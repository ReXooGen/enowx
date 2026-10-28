package daemon

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/proto"
)

func (s *Server) agentModels(p proto.AgentModelsParams) (proto.AgentModels, error) {
	var dir, account string
	if p.Session != "" {
		sess, err := s.sessions.Get(p.Session)
		if err != nil {
			return proto.AgentModels{}, err
		}
		info := sess.Info()
		p.Agent, dir, account = info.Agent, info.Dir, info.Account
	} else if p.Canvas != "" {
		if _, err := s.canvases.Get(p.Canvas); err != nil {
			return proto.AgentModels{}, err
		}
		dir = filepath.Join(filepath.Dir(s.WorktreeRoot), "canvas-chat", p.Canvas)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return proto.AgentModels{}, err
		}
	}
	if _, ok := agent.Find(p.Agent); !ok {
		return proto.AgentModels{}, fmt.Errorf("agent: unknown agent")
	}
	env, err := s.accounts.Env(agent.Base(p.Agent), account)
	if err != nil {
		return proto.AgentModels{}, err
	}
	refreshLoginPath(s.logger())
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return agent.Models(ctx, p.Agent, dir, env)
}
