package cmd

import (
	"io"
	"net"
	"net-cat/service"
	"net-cat/utils"
	"testing"
	"time"
)

func TestHandleClientSendsBannerAndEnqueuesClient(t *testing.T) {
	server := &service.Server{
		Clients:   make(map[string]*service.Client),
		Broadcast: make(chan service.Message, 4),
		Join:      make(chan *service.Client),
		Leave:     make(chan *service.Client, 4),
		History:   []string{},
	}

	clientConn, peerConn := net.Pipe()
	defer peerConn.Close()

	go HandleClient(clientConn, server)

	bannerBuf := make([]byte, len(utils.Banner))
	if _, err := io.ReadFull(peerConn, bannerBuf); err != nil {
		t.Fatalf("failed reading banner: %v", err)
	}
	if string(bannerBuf) != utils.Banner {
		t.Fatal("banner mismatch")
	}

	_, _ = peerConn.Write([]byte("alice\n"))

	select {
	case joined := <-server.Join:
		if joined.Name != "alice" {
			t.Fatalf("expected queued client name alice, got %q", joined.Name)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("client was not queued on Join channel")
	}
}
