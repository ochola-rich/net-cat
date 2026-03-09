package server

import (
	"io"
	"net"
	"net-cat/service"
	"net-cat/utils"
	"testing"
	"time"
)

func TestStartReturnsErrorForInvalidPort(t *testing.T) {
	if err := Start("invalid-port"); err == nil {
		t.Fatal("expected invalid port error")
	}
}

func TestHandleConnectionRejectsWhenFull(t *testing.T) {
	s := service.NewServer(10)
	for i := 0; i < maxClients; i++ {
		name := "client"
		s.Clients[name+string(rune('A'+i))] = &service.Client{Name: name, Messages: make(chan string, 1)}
	}

	serverConn, peerConn := net.Pipe()
	defer peerConn.Close()

	done := make(chan struct{})
	go func() {
		handleConnection(s, serverConn)
		close(done)
	}()

	msg, err := io.ReadAll(peerConn)
	if err != nil {
		t.Fatalf("failed to read full-response message: %v", err)
	}

	if string(msg) != "Server full. Maximum 10 clients allowed.\n" {
		t.Fatalf("unexpected full-response message: %q", string(msg))
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handleConnection did not return")
	}
}

func TestHandleConnectionStartsClientHandlerWhenNotFull(t *testing.T) {
	s := service.NewServer(10)
	serverConn, peerConn := net.Pipe()
	defer peerConn.Close()

	done := make(chan struct{})
	go func() {
		handleConnection(s, serverConn)
		close(done)
	}()

	bannerBuf := make([]byte, len(utils.Banner))
	if _, err := io.ReadFull(peerConn, bannerBuf); err != nil {
		t.Fatalf("failed reading banner from client handler: %v", err)
	}
	if string(bannerBuf) != utils.Banner {
		t.Fatal("banner mismatch")
	}
	_, _ = peerConn.Write([]byte("alice\n"))

	select {
	case joined := <-s.Join:
		if joined.Name != "alice" {
			t.Fatalf("expected joined name alice, got %q", joined.Name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for join event")
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
	}
}
