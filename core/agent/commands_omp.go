package agent

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/enowx/enowx/internal/quietcmd"
	"github.com/enowx/enowx/proto"
)

// ImmediateCommands keeps the composer useful while the CLI discovers plugins.
func ImmediateCommands(id string) []proto.AgentCommand {
	if Base(id) != "omp" {
		return nil
	}
	return []proto.AgentCommand{{Name: "model", Description: "Select model and provider", Kind: "built-in"}, {Name: "session", Description: "Session information", Kind: "built-in"}, {Name: "usage", Description: "Provider usage", Kind: "built-in"}, {Name: "compact", Description: "Compact the conversation", Kind: "built-in"}, {Name: "switch", Description: "Switch model", Kind: "built-in"}, {Name: "fast", Description: "Fast mode", Kind: "built-in"}, {Name: "slow", Description: "Slow mode", Kind: "built-in"}, {Name: "security", Description: "Security settings", Kind: "built-in"}}
}

func ompCommands(ctx context.Context, dir string) ([]proto.AgentCommand, error) {
	cmd := quietcmd.CommandContext(ctx, "omp", "--mode", "rpc", "--no-ui", "--no-session")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader("{\"id\":\"enowx-commands\",\"type\":\"get_available_commands\"}\n")
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("agent: OMP command discovery failed: %w", err)
	}
	return parseOMPCommands(data)
}

func parseOMPCommands(data []byte) ([]proto.AgentCommand, error) {
	scan := bufio.NewScanner(bytes.NewReader(data))
	scan.Buffer(make([]byte, 64<<10), 4<<20)
	for scan.Scan() {
		var frame struct {
			Type, Command string
			Success       bool
			Data          struct {
				Commands []struct{ Name, Description, Source string }
			}
		}
		if json.Unmarshal(scan.Bytes(), &frame) != nil || frame.Type != "response" || frame.Command != "get_available_commands" {
			continue
		}
		if !frame.Success {
			return nil, fmt.Errorf("agent: OMP cannot list commands")
		}
		list := []proto.AgentCommand{}
		seen := map[string]bool{}
		for _, item := range frame.Data.Commands {
			name := strings.TrimPrefix(item.Name, "/")
			if name == "" || strings.ContainsAny(name, " \t\r\n\x00") || seen[name] {
				continue
			}
			seen[name] = true
			kind := "command"
			if item.Source == "builtin" {
				kind = "built-in"
			}
			if item.Source == "skill" {
				kind = "skill"
			}
			list = append(list, proto.AgentCommand{Name: name, Description: item.Description, Kind: kind})
		}
		return list, nil
	}
	return nil, fmt.Errorf("agent: OMP returned no command catalog")
}
