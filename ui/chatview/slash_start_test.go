package chatview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestFreshOMPCommandsDoNotWaitForAConversation(t *testing.T) {
	ch := New()
	ch.input = "/"
	var finish func(proto.AgentCommands, error)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(420).Height(600).Children(func() {
			ch.View(c, Options{Agent: "omp", AgentName: "OMP", Fresh: true, LoadCommands: func(done func(proto.AgentCommands, error)) { finish = done }})
		})
	}, 420, 600)
	tt.Frame()
	if !tt.HasText("/model") || tt.HasText("Reading the agent's commands…") {
		t.Fatalf("fresh slash menu is stuck: %q", tt.Texts())
	}
	if finish == nil {
		t.Fatal("CLI discovery did not start")
	}
	finish(proto.AgentCommands{Commands: []proto.AgentCommand{{Name: "project:verify", Description: "Project command"}}}, nil)
	tt.Frame()
	if !tt.HasText("/project:verify") {
		t.Fatal("discovered command did not replace starter list")
	}
}
