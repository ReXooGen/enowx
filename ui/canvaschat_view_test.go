package ui

import (
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
)

func TestCanvasAgentCanSwitchToTerminalWhileWorking(t *testing.T) {
	a := New(Options{})
	hc := &hostCanvasChat{}
	s := proto.SessionInfo{Agent: "omp", State: proto.StateWorking}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(400).Height(100).Children(func() { a.canvasAgentViewToggle(c, hc, s) })
	}, 400, 100)
	tt.Frame()
	if !tt.HasText("working") {
		t.Fatal("working status is missing")
	}
	if err := tt.Click("Terminal"); err != nil {
		t.Fatal(err)
	}
	if !hc.terminal {
		t.Fatal("terminal selection was lost")
	}
	if err := tt.Click("Chat"); err != nil {
		t.Fatal(err)
	}
	if hc.terminal {
		t.Fatal("could not return to the transcript")
	}
}
