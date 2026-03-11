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
	server := service.NewServer(10)
	group := server.GetOrCreateGroup("lobby")

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

	deadline := time.Now().Add(1 * time.Second)
	for time.Now().Before(deadline) {
		group.Mutex.Lock()
		_, ok := group.Clients["alice"]
		group.Mutex.Unlock()
		if ok {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("client was not added to lobby group")
}
