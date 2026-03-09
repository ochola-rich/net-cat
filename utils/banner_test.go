package utils

import (
	"strings"
	"testing"
)

func TestBannerContainsWelcomeAndPrompt(t *testing.T) {
	if !strings.Contains(Banner, "Welcome to TCP-Chat!") {
		t.Fatal("banner must include welcome text")
	}
	if !strings.Contains(Banner, "[ENTER YOUR NAME]: ") {
		t.Fatal("banner must include name prompt")
	}
	if strings.Count(Banner, "\n") < 5 {
		t.Fatal("banner should be multi-line")
	}
}
