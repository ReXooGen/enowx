package session

import (
	"time"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/proto"
)

// inputObserver belongs to watchActivity. Read text only after the screen
// settles, at most once a second, and never again until output or input changes.
// Terminal.Text includes scrollback, so reading it on every frame is expensive.
type inputObserver struct {
	checked, screen, input time.Time
}

func (o *inputObserver) check(s *Session, now time.Time) {
	s.mu.Lock()
	screen, input := s.screenAt, s.lastInput
	hookAt := s.stateAt
	skip := s.exited || s.meta.Queued || s.state == proto.StateWaiting || s.state == proto.StateDone ||
		screen.IsZero() || now.Sub(screen) < 300*time.Millisecond || now.Sub(input) < 300*time.Millisecond ||
		now.Sub(o.checked) < time.Second || (screen == o.screen && input == o.input)
	s.mu.Unlock()
	if skip {
		return
	}
	o.checked = now
	text := s.term.Text()
	_, rows := s.term.Size()
	waiting := agent.InputPrompt(text, rows)
	s.mu.Lock()
	// A key or a new frame while text was copied makes that snapshot obsolete.
	if s.screenAt == screen && s.lastInput == input && s.stateAt == hookAt {
		s.promptWaiting = waiting
		o.screen, o.input = screen, input
	}
	s.mu.Unlock()
}
