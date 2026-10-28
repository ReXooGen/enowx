package agent

import (
	"testing"
)

func TestOMPCommandCatalogUsesItsRPCDiscovery(t *testing.T) {
	data := []byte(`{"type":"ready"}
{"type":"available_commands_update","commands":[]}
{"type":"response","command":"get_available_commands","success":true,"data":{"commands":[{"name":"model","description":"Choose model","source":"builtin"},{"name":"skill:review","source":"skill"},{"name":"project:check","source":"file"},{"name":"model","source":"builtin"}]}}
`)
	list, err := parseOMPCommands(data)
	if err != nil || len(list) != 3 {
		t.Fatalf("catalog: %+v %v", list, err)
	}
	if list[0].Kind != "built-in" || list[1].Kind != "skill" || list[2].Name != "project:check" {
		t.Fatalf("command metadata lost: %+v", list)
	}
}
