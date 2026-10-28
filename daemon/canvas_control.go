package daemon

import (
	"fmt"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/core/session"
	"github.com/enowx/enowx/proto"
)

// canvasAgent finds the conversation even before it has made its first edit.
func (s *Server) canvasAgent(c proto.Canvas) *session.Session {
	list := s.sessions.List()
	for i := len(list) - 1; i >= 0; i-- {
		info := list[i].Info()
		if info.CanvasFor == c.ID && !info.Exited && !info.Hibernated {
			return list[i]
		}
	}
	sess, _ := s.sessions.Get(c.Agent)
	return sess
}

func (s *Server) interruptCanvasAgent(c proto.Canvas) {
	sess := s.canvasAgent(c)
	id := ""
	if sess != nil {
		id = sess.ID()
	}
	s.canvasState.mu.Lock()
	if s.canvasState.pausedSession == nil {
		s.canvasState.pausedSession = make(map[string]string)
	}
	s.canvasState.pausedSession[c.ID] = id
	s.canvasState.mu.Unlock()
	if sess == nil || sess.Info().Exited || sess.Info().Hibernated {
		return
	}
	key := "\x03"
	if preset, ok := agent.Find(sess.Meta().Agent); ok {
		key = preset.InterruptKey()
	}
	s.sessions.Input(sess, []byte(key))
}

func (s *Server) resumeCanvasAgent(c proto.Canvas) error {
	s.canvasState.mu.Lock()
	id := s.canvasState.pausedSession[c.ID]
	s.canvasState.mu.Unlock()
	sess := s.canvasAgent(c)
	if id != "" {
		var err error
		sess, err = s.sessions.Get(id)
		if err != nil {
			return fmt.Errorf("canvas: paused conversation is no longer available; open Ask an agent to continue: %w", err)
		}
	}
	if sess == nil {
		return nil // A canvas can be paused without a conversation.
	}
	info := sess.Info()
	if info.Exited || info.Hibernated || info.Queued {
		return fmt.Errorf("canvas: conversation is not running; open Ask an agent to continue")
	}
	text := canvas.ChatContext(c, "", "") + "The user resumed this canvas. Continue the unfinished task from this conversation. Read the current canvas and comments first, preserve edits made while paused, and continue from the last completed step without duplicating designs or elements. If the task was already complete, report that instead of repeating it."
	if err := s.input(proto.SessionInputParams{ID: info.ID, Text: text, Paste: true, Enter: true}); err != nil {
		return fmt.Errorf("canvas: resume conversation: %w", err)
	}
	return nil
}
