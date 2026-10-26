package session

import (
	"testing"
	"time"

	"github.com/enowx/enowx/proto"
)

func TestInputPromptTracksTheRenderedDialogAndAnswers(t *testing.T) {
	m := NewManager()
	defer m.CloseAll()
	// An idle terminal gives the test control of its screen without a shell.
	s, err := m.Create(Spec{Exited: true, Cols: 120, Rows: 12})
	if err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.exited = false
	s.state = proto.StateWorking
	s.stateAt = time.Now()
	s.mu.Unlock()
	s.term.Feed([]byte("\x1b[2J\x1b[HChoose a theme\r\nDark\r\nLight\r\nspace select · ↑/↓ move · esc cancel"))
	now := time.Now()
	s.noteScreen(now.Add(-time.Second), true)
	var observer inputObserver
	observer.check(s, now)
	if got := s.Info().State; got != proto.StateWaiting {
		t.Fatalf("dialog state %q", got)
	}
	m.Input(s, []byte("\x1b[B"))
	if got := s.Info().State; got == proto.StateWaiting {
		t.Fatal("typing did not clear inferred input state")
	}
	observer.check(s, now.Add(2*time.Second))
	if got := s.Info().State; got != proto.StateWaiting {
		t.Fatalf("navigating an open dialog lost input state: %q", got)
	}
	s.term.Feed([]byte("\x1b[2J\x1b[HGenerating a preview..."))
	s.noteScreen(now.Add(3*time.Second), true)
	observer.check(s, now.Add(4*time.Second))
	if got := s.Info().State; got != proto.StateWorking {
		t.Fatalf("closed dialog state %q", got)
	}
}

func TestInputPromptHonorsExplicitHooksAndSkipsUnchangedScreens(t *testing.T) {
	m := NewManager()
	defer m.CloseAll()
	s, err := m.Create(Spec{Exited: true, Cols: 100, Rows: 12})
	if err != nil {
		t.Fatal(err)
	}
	s.mu.Lock()
	s.exited = false
	s.mu.Unlock()
	s.term.Feed([]byte("Enter to confirm | Esc to cancel"))
	now := time.Now()
	s.noteScreen(now.Add(-time.Second), true)
	var observer inputObserver
	observer.check(s, now)
	checked := observer.checked
	observer.check(s, now.Add(10*time.Second))
	if observer.checked != checked {
		t.Fatal("unchanged screen was read again")
	}
	if err := m.SetState(s, proto.StateDone); err != nil {
		t.Fatal(err)
	}
	observer.check(s, now.Add(11*time.Second))
	if got := s.Info().State; got != proto.StateDone {
		t.Fatalf("explicit hook state was overridden: %q", got)
	}
}
