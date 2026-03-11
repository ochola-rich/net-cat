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
	if s.Groups == nil || len(s.Groups) != 1 {
		t.Fatal("expected lobby group to be initialized")
	}
	group := s.GetOrCreateGroup("lobby")
	if group.Clients == nil || len(group.Clients) != 0 {
		t.Fatal("expected empty clients map")
	}
	if group.Broadcast == nil || group.Join == nil || group.Leave == nil {
		t.Fatal("expected all channels to be initialized")
	}
	if group.History == nil || len(group.History) != 0 {
		t.Fatal("expected empty history")
	}
}

func TestBroadcastToOthersSkipsSender(t *testing.T) {
	s := NewServer(10)
	group := s.GetOrCreateGroup("lobby")
	sender := &Client{Name: "alice", Messages: make(chan string, 1)}
	receiver := &Client{Name: "bob", Messages: make(chan string, 1)}
	group.Clients[sender.Name] = sender
	group.Clients[receiver.Name] = receiver

	group.broadcastToOthers("hello", sender)

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
	group := s.GetOrCreateGroup("lobby")
	group.History = []string{"old-1", "old-2"}

	existing := &Client{Name: "bob", Messages: make(chan string, 2)}
	joining := &Client{Name: "alice", Messages: make(chan string, 4)}
	group.Clients[existing.Name] = existing

	group.Join <- joining

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

	if _, ok := group.Clients[joining.Name]; !ok {
		t.Fatal("joining client not tracked in group")
	}

	if len(group.History) == 0 || !strings.Contains(group.History[len(group.History)-1], "alice has joined our chat.") {
		t.Fatal("join message not added to history")
	}
}

func TestRunLeaveBroadcastsSystemMessageAndAddsHistory(t *testing.T) {
	s := NewServer(10)
	group := s.GetOrCreateGroup("lobby")
	leaving := &Client{Name: "alice", Messages: make(chan string, 1)}
	other := &Client{Name: "bob", Messages: make(chan string, 1)}
	group.Clients[leaving.Name] = leaving
	group.Clients[other.Name] = other

	group.Leave <- leaving

	select {
	case msg := <-other.Messages:
		if !strings.Contains(msg, "[System]: alice has left our chat.") {
			t.Fatalf("unexpected leave broadcast: %q", msg)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for leave broadcast")
	}

	if len(group.History) == 0 || !strings.Contains(group.History[len(group.History)-1], "alice has left our chat.") {
		t.Fatal("leave message not added to history")
	}
}

func TestRunBroadcastIgnoresEmptyAndFormatsValidMessages(t *testing.T) {
	s := NewServer(10)
	group := s.GetOrCreateGroup("lobby")
	sender := &Client{Name: "alice", Messages: make(chan string, 1)}
	receiver := &Client{Name: "bob", Messages: make(chan string, 1)}
	group.Clients[sender.Name] = sender
	group.Clients[receiver.Name] = receiver

	group.Broadcast <- Message{Sender: sender, Content: "   "}

	select {
	case got := <-receiver.Messages:
		t.Fatalf("did not expect message for blank content, got %q", got)
	case <-time.After(150 * time.Millisecond):
	}

	group.Broadcast <- Message{Sender: sender, Content: " hello world "}

	select {
	case got := <-receiver.Messages:
		if !regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\]\[alice\]: hello world$`).MatchString(got) {
			t.Fatalf("unexpected formatted message: %q", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("timed out waiting for user broadcast")
	}

	if len(group.History) == 0 || !strings.Contains(group.History[len(group.History)-1], "[alice]: hello world") {
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
