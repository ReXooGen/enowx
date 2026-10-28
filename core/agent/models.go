package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/enowx/enowx/internal/quietcmd"
	"github.com/enowx/enowx/proto"
)

// Models reads the CLI's own catalog where available. Only safe display fields
// cross the protocol; configuration and authentication stay on the host.
func Models(ctx context.Context, id, dir string, env []string) (proto.AgentModels, error) {
	id = Base(id)
	if id != "omp" {
		result := proto.AgentModels{Models: []proto.AgentModel{}}
		for _, choice := range SettingsOf(id).Models {
			result.Models = append(result.Models, proto.AgentModel{ID: choice.ID, Name: choice.Label, Selector: choice.ID})
		}
		return result, nil
	}
	cmd := quietcmd.CommandContext(ctx, "omp", "models", "--json")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	data, err := cmd.Output()
	if err != nil {
		return proto.AgentModels{}, fmt.Errorf("agent: OMP model catalog unavailable: %w", err)
	}
	return parseOMPModels(data)
}

func parseOMPModels(data []byte) (proto.AgentModels, error) {
	result := proto.AgentModels{Registered: true, Models: []proto.AgentModel{}}
	if len(data) > 8<<20 {
		return result, fmt.Errorf("agent: OMP model catalog exceeds 8 MB")
	}
	var catalog struct {
		Models []struct{ Provider, ID, Name, Selector, Kind string }
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		return result, fmt.Errorf("agent: OMP returned an unreadable model catalog")
	}
	seen := map[string]bool{}
	for _, m := range catalog.Models {
		if m.Kind != "" && m.Kind != "chat" {
			continue
		}
		if m.Provider == "" || m.ID == "" || strings.ContainsAny(m.Provider+m.ID, " \t\r\n\x00") {
			continue
		}
		selector := m.Provider + "/" + m.ID
		if m.Selector != "" && m.Selector != selector {
			continue
		}
		if seen[selector] {
			continue
		}
		seen[selector] = true
		if m.Name == "" {
			m.Name = m.ID
		}
		result.Models = append(result.Models, proto.AgentModel{Provider: m.Provider, ID: m.ID, Name: m.Name, Selector: selector})
		if len(result.Models) > 10000 {
			return proto.AgentModels{}, fmt.Errorf("agent: OMP model catalog exceeds 10000 entries")
		}
	}
	slices.SortFunc(result.Models, func(a, b proto.AgentModel) int { return strings.Compare(a.Selector, b.Selector) })
	return result, nil
}
