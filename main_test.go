package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("copy stdout error: %v", err)
	}
	_ = r.Close()
	return buf.String()
}

func TestMainPrintsUsageWhenTooManyArgs(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"TCPChat", "8080", "extra"}
	out := captureStdout(t, main)

	if !strings.Contains(out, "[USAGE]: ./TCPChat $port") {
		t.Fatalf("expected usage output, got %q", out)
	}
}

func TestMainPrintsErrorForInvalidPortArg(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	os.Args = []string{"TCPChat", "invalid-port"}
	out := captureStdout(t, main)

	if !strings.Contains(out, "Error:") {
		t.Fatalf("expected startup error output, got %q", out)
	}
}
