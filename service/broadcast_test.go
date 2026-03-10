package service

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestNewServerInitializesState(t *testing.T) {
	s := NewServer(10)
	if s == nil {
		t.Fatal("expected server instance")
	}
	if s.Clients == nil || len(s.Clients) != 0 {
		t.Fatal("expected empty clients map")
	}
	if s.Broadcast == nil || s.Join == nil || s.Leave == nil {
		t.Fatal("expected all channels to be initialized")
	}
	if s.History == nil || len(s.History) != 0 {
		t.Fatal("expected empty history")
	}
}

func TestBroadcastToOthersSkipsSender(t *testing.T) {
	s := NewServer(10)
	sender := &Client{Name: "alice", Messages: make(chan string, 1)}
	receiver := &Client{Name: "bob", Messages: make(chan string, 1)}
	s.Clients[sender.Name] = sender
	s.Clients[receiver.Name] = receiver

	s.broadcastToOthers("hello", sender)

	select {
	case got := <-receiver.Messages:
		if got != "hello" {
			t.Fatalf("unexpected receiver message: %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("receiver did not get broadcast")
	}

	select {
	case got := <-sender.Messages:
		t.Fatalf("sender should not receive message, got %q", got)
	default:
	}
}

func TestRunJoinSendsHistoryAndBroadcastsSystemMessage(t *testing.T) {
	s := NewServer(10)
	s.History = []string{"old-1", "old-2"}

	existing := &Client{Name: "bob", Messages: make(chan string, 2)}
	joining := &Client{Name: "alice", Messages: make(chan string, 4)}
	s.Clients[existing.Name] = existing

	go s.Broadcasts()
	s.Join <- joining

	for i, want := range []string{"old-1", "old-2"} {
		select {
		case got := <-joining.Messages:
			if got != want {
				t.Fatalf("history[%d]: got %q want %q", i, got, want)
			}
		case <-time.After(500 * time.Millisecond):
			t.Fatalf("timed out waiting for history[%d]", i)
		}
	}

	select {
	case msg := <-existing.Messages:
		if !strings.Contains(msg, "[System]: alice has joined our chat.") {
			t.Fatalf("unexpected join broadcast: %q", msg)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for join broadcast")
	}

	if _, ok := s.Clients[joining.Name]; !ok {
		t.Fatal("joining client not tracked in server")
	}

	if len(s.History) == 0 || !strings.Contains(s.History[len(s.History)-1], "alice has joined our chat.") {
		t.Fatal("join message not added to history")
	}
}

func TestRunLeaveBroadcastsSystemMessageAndAddsHistory(t *testing.T) {
	s := NewServer(10)
	leaving := &Client{Name: "alice", Messages: make(chan string, 1)}
	other := &Client{Name: "bob", Messages: make(chan string, 1)}
	s.Clients[leaving.Name] = leaving
	s.Clients[other.Name] = other

	go s.Broadcasts()
	s.Leave <- leaving

	select {
	case msg := <-other.Messages:
		if !strings.Contains(msg, "[System]: alice has left our chat.") {
			t.Fatalf("unexpected leave broadcast: %q", msg)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for leave broadcast")
	}

	if len(s.History) == 0 || !strings.Contains(s.History[len(s.History)-1], "alice has left our chat.") {
		t.Fatal("leave message not added to history")
	}
}

func TestRunBroadcastIgnoresEmptyAndFormatsValidMessages(t *testing.T) {
	s := NewServer(10)
	sender := &Client{Name: "alice", Messages: make(chan string, 1)}
	receiver := &Client{Name: "bob", Messages: make(chan string, 1)}
	s.Clients[sender.Name] = sender
	s.Clients[receiver.Name] = receiver

	go s.Broadcasts()
	s.Broadcast <- Message{Sender: sender, Content: "   "}

	select {
	case got := <-receiver.Messages:
		t.Fatalf("did not expect message for blank content, got %q", got)
	case <-time.After(150 * time.Millisecond):
	}

	s.Broadcast <- Message{Sender: sender, Content: " hello world "}

	select {
	case got := <-receiver.Messages:
		if !regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\[alice\]: hello world$`).MatchString(got) {
			t.Fatalf("unexpected formatted message: %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for user broadcast")
	}

	if len(s.History) == 0 || !strings.Contains(s.History[len(s.History)-1], "[alice]: hello world") {
		t.Fatal("valid message not added to history")
	}
}

func TestFormatHelpers(t *testing.T) {
	user := formatUserMessage("alice", "hi")
	system := formatSystemMessage("welcome")

	userPattern := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\[alice\]: hi$`)
	systemPattern := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\[System\]: welcome$`)

	if !userPattern.MatchString(user) {
		t.Fatalf("unexpected user format: %q", user)
	}
	if !systemPattern.MatchString(system) {
		t.Fatalf("unexpected system format: %q", system)
	}
}
