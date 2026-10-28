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
	id, dir, env, err := s.agentCatalogContext(p)
	if err != nil {
		return proto.AgentModels{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	return agent.Models(ctx, id, dir, env)
}

func (s *Server) agentCommands(p proto.AgentModelsParams) (proto.AgentCommands, error) {
	id, dir, _, err := s.agentCatalogContext(p)
	if err != nil {
		return proto.AgentCommands{}, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	list, err := agent.Commands(ctx, id, dir)
	return proto.AgentCommands{Commands: list}, err
}

func (s *Server) agentCatalogContext(p proto.AgentModelsParams) (string, string, []string, error) {
	var dir, account string
	if p.Session != "" {
		sess, err := s.sessions.Get(p.Session)
		if err != nil {
			return "", "", nil, err
		}
		info := sess.Info()
		p.Agent, dir, account = info.Agent, info.Dir, info.Account
	} else if p.Canvas != "" {
		if _, err := s.canvases.Get(p.Canvas); err != nil {
			return "", "", nil, err
		}
		dir = filepath.Join(filepath.Dir(s.WorktreeRoot), "canvas-chat", p.Canvas)
		if err := os.MkdirAll(dir, 0700); err != nil {
			return "", "", nil, err
		}
	}
	if _, ok := agent.Find(p.Agent); !ok {
		return "", "", nil, fmt.Errorf("agent: unknown agent")
	}
	env, err := s.accounts.Env(agent.Base(p.Agent), account)
	if err != nil {
		return "", "", nil, err
	}
	refreshLoginPath(s.logger())
	return p.Agent, dir, env, nil
}
