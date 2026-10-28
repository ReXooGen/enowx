package ui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
)

func TestCanvasModelProviderSelectionBeforeStarting(t *testing.T) {
	a := New(Options{})
	hc := &hostCanvasChat{canvas: "canvas", modelOpen: true, modelDraft: "deepseek-v4-flash", providerDraft: "custom"}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(600).Height(500).TextColor(theme.Text).Children(func() { a.canvasModelDialog(c, hc, "omp", proto.SessionInfo{}, false) })
	}, 600, 500)
	tt.Frame()
	tt.Advance(time.Second)
	shot(t, tt, "canvas-model-provider.png")
	if err := tt.Click("Apply model"); err != nil {
		t.Fatal(err)
	}
	if got := a.notesModel["omp"]; got != "custom/deepseek-v4-flash" {
		t.Fatalf("model choice lost provider: %q", got)
	}
	if got := agent.LaunchArgs("omp", a.notesModel["omp"], ""); len(got) != 2 || got[0] != "--model" || got[1] != "custom/deepseek-v4-flash" {
		t.Fatalf("launch ignored choice: %q", got)
	}
	if _, ok := canvasModelID("provider", "bad\n/model injected"); ok {
		t.Fatal("multiline model accepted")
	}
}
