package service

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestWriteOutputWritesToConnection(t *testing.T) {
	clientConn, peerConn := net.Pipe()
	defer peerConn.Close()

	c := &Client{
		Conn:     clientConn,
		Messages: make(chan string, 1),
	}

	done := make(chan struct{})
	go func() {
		c.WriteOutput()
		close(done)
	}()

	c.Messages <- "hello"
	close(c.Messages)

	buf := make([]byte, len("hello"))
	if _, err := io.ReadFull(peerConn, buf); err != nil {
		t.Fatalf("failed reading message: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("unexpected payload: %q", string(buf))
	}

	clientConn.Close()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("WriteOutput did not exit")
	}
}

func TestReadInputPublishesMessagesAndLeavesOnDisconnect(t *testing.T) {
	clientConn, peerConn := net.Pipe()
	defer peerConn.Close()

	s := &Server{
		Broadcast: make(chan Message, 2),
		Leave:     make(chan *Client, 1),
	}
	c := &Client{Conn: clientConn, Name: "alice"}

	done := make(chan struct{})
	go func() {
		c.ReadInput(s)
		close(done)
	}()

	_, _ = peerConn.Write([]byte("  hello  \n\n"))
	peerConn.Close()

	select {
	case msg := <-s.Broadcast:
		if msg.Sender != c {
			t.Fatal("broadcast sender mismatch")
		}
		if msg.Content != "hello" {
			t.Fatalf("expected trimmed content, got %q", msg.Content)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for broadcast")
	}

	select {
	case <-s.Broadcast:
		t.Fatal("blank line should not produce a broadcast message")
	default:
	}

	select {
	case left := <-s.Leave:
		if left != c {
			t.Fatal("leave event client mismatch")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for leave event")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ReadInput did not exit")
	}
}
