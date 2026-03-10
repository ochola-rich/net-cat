package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net-cat/server"
	"net-cat/service"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

const (
	viewChat  = "chat"
	viewInput = "input"

	namePrompt = "[ENTER YOUR NAME]:"
)

type inputMode int

const (
	modeName inputMode = iota
	modeChat
)

type app struct {
	gui           *gocui.Gui
	conn          net.Conn
	reader        *bufio.Reader
	mode          inputMode
	name          string
	banner        string
	bannerPrinted bool
	readStarted   bool
}

func main() {
	withServer := flag.Bool("with-server", false, "start a local server before connecting")
	flag.BoolVar(withServer, "local", false, "start a local server before connecting")
	flag.Parse()

	addr, port, err := parseAddr(flag.Args(), *withServer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	var conn net.Conn
	if *withServer {
		serverErr := startLocalServer(port)
		conn, err = dialWithRetry(addr, 2*time.Second, serverErr)
	} else {
		conn, err = net.Dial("tcp", addr)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	banner, err := readBanner(reader)
	if err != nil && !errors.Is(err, io.EOF) {
		fmt.Fprintln(os.Stderr, "read banner:", err)
	}

	gui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer gui.Close()

	app := &app{
		gui:    gui,
		conn:   conn,
		reader: reader,
		mode:   modeName,
		banner: banner,
	}

	gui.Cursor = true
	gui.SetManagerFunc(app.layout)

	if err := gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, app.quit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := gui.SetKeybinding("", gocui.KeyCtrlQ, gocui.ModNone, app.quit); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if err := gui.SetKeybinding(viewInput, gocui.KeyEnter, gocui.ModNone, app.handleEnter); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func parseAddr(args []string, withServer bool) (string, string, error) {
	var host string
	var port string

	switch len(args) {
	case 0:
		host = "localhost"
		port = service.DefaultPort
	case 1:
		if strings.Contains(args[0], ":") {
			parsedHost, parsedPort, err := net.SplitHostPort(args[0])
			if err != nil {
				return "", "", fmt.Errorf("invalid address: %w", err)
			}
			if parsedHost == "" {
				parsedHost = "localhost"
			}
			host = parsedHost
			port = parsedPort
		} else {
			host = "localhost"
			port = args[0]
		}
	default:
		return "", "", fmt.Errorf("usage: %s [--with-server] [host:port|port]", filepath.Base(os.Args[0]))
	}

	if port == "" {
		return "", "", fmt.Errorf("invalid address: missing port")
	}

	if withServer && !isLocalHost(host) {
		return "", "", fmt.Errorf("with-server mode only supports localhost")
	}

	return net.JoinHostPort(host, port), port, nil
}

func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func startLocalServer(port string) <-chan error {
	errCh := make(chan error, 1)
	logger := log.New(io.Discard, "", 0)

	go func() {
		errCh <- server.StartWithOptions(port, server.Options{
			InfoWriter:  io.Discard,
			ErrorLogger: logger,
		})
	}()

	return errCh
}

func dialWithRetry(addr string, timeout time.Duration, errCh <-chan error) (net.Conn, error) {
	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 200*time.Millisecond)
		if err == nil {
			return conn, nil
		}
		lastErr = err

		select {
		case serr := <-errCh:
			if serr != nil {
				return nil, serr
			}
			return nil, fmt.Errorf("server exited before accept")
		default:
		}

		time.Sleep(100 * time.Millisecond)
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("timed out")
	}
	return nil, fmt.Errorf("connect: %w", lastErr)
}

func (a *app) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	chatHeight := maxY - 4
	if chatHeight < 1 {
		chatHeight = 1
	}

	if v, err := g.SetView(viewChat, 0, 0, maxX-1, chatHeight); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "Messages"
		v.Wrap = true
		v.Autoscroll = true
		if !a.bannerPrinted {
			a.bannerPrinted = true
			a.printBanner(v)
		}
	}

	if v, err := g.SetView(viewInput, 0, chatHeight+1, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = a.inputTitle()
		v.Editable = true
		v.Wrap = true
		if _, err := g.SetCurrentView(viewInput); err != nil {
			return err
		}
	}

	return nil
}

func (a *app) inputTitle() string {
	if a.mode == modeName {
		return "Name"
	}
	return "Message"
}

func (a *app) printBanner(v *gocui.View) {
	if strings.TrimSpace(a.banner) != "" {
		fmt.Fprintln(v, a.banner)
	}
	fmt.Fprintln(v, "Enter your name in the input below.")
}

func (a *app) handleEnter(g *gocui.Gui, v *gocui.View) error {
	input := strings.TrimSpace(v.Buffer())
	v.Clear()
	v.SetCursor(0, 0)

	if input == "" {
		return nil
	}

	switch a.mode {
	case modeName:
		if err := a.sendLine(input); err != nil {
			a.printChat(fmt.Sprintf("Failed to send name: %v", err))
			return nil
		}
		a.name = input
		a.mode = modeChat
		v.Title = a.inputTitle()
		a.printChat(fmt.Sprintf("Connected as %s.", a.name))
		if !a.readStarted {
			a.readStarted = true
			go a.readLoop()
		}
	case modeChat:
		if err := a.sendLine(input); err != nil {
			a.printChat(fmt.Sprintf("Failed to send message: %v", err))
			return nil
		}
		a.printChat(formatLocalMessage(a.name, input))
	}

	return nil
}

func (a *app) sendLine(line string) error {
	_, err := fmt.Fprintln(a.conn, line)
	return err
}

func (a *app) readLoop() {
	for {
		line, err := a.reader.ReadString('\n')
		if line != "" {
			a.printChatAsync(strings.TrimRight(line, "\n"))
		}
		if err != nil {
			if !errors.Is(err, io.EOF) {
				a.printChatAsync(fmt.Sprintf("Connection closed: %v", err))
			} else {
				a.printChatAsync("Connection closed by server.")
			}
			return
		}
	}
}

func (a *app) printChat(message string) {
	v, err := a.gui.View(viewChat)
	if err != nil {
		return
	}
	fmt.Fprintln(v, message)
}

func (a *app) printChatAsync(message string) {
	a.gui.Update(func(g *gocui.Gui) error {
		v, err := g.View(viewChat)
		if err != nil {
			return nil
		}
		fmt.Fprintln(v, message)
		return nil
	})
}

func (a *app) quit(g *gocui.Gui, v *gocui.View) error {
	if a.conn != nil {
		_ = a.conn.Close()
	}
	return gocui.ErrQuit
}

func readBanner(r *bufio.Reader) (string, error) {
	var buf bytes.Buffer
	tmp := make([]byte, 256)

	for {
		n, err := r.Read(tmp)
		if n > 0 {
			buf.Write(tmp[:n])
			if bytes.Contains(buf.Bytes(), []byte(namePrompt)) {
				return trimBanner(buf.String()), nil
			}
		}
		if err != nil {
			return trimBanner(buf.String()), err
		}
	}
}

func trimBanner(banner string) string {
	if idx := strings.Index(banner, namePrompt); idx >= 0 {
		banner = banner[:idx]
	}
	return strings.TrimRight(banner, "\n")
}

func formatLocalMessage(name, msg string) string {
	return fmt.Sprintf("[%s][%s]: %s",
		time.Now().Format("2006-01-02 15:04:05"),
		name,
		msg,
	)
}
