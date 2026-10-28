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
	hc := &hostCanvasChat{canvas: "canvas", modelOpen: true, modelDraft: "deepseek-v4-flash", providerDraft: "custom", models: []proto.AgentModel{{Provider: "custom", ID: "deepseek-v4-flash", Name: "DeepSeek V4 Flash", Selector: "custom/deepseek-v4-flash"}}}
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

func TestCanvasProviderChangeOnlyOffersItsRegisteredModels(t *testing.T) {
	a := New(Options{})
	hc := &hostCanvasChat{modelOpen: true, providerDraft: "alpha", modelDraft: "a", models: []proto.AgentModel{{Provider: "alpha", ID: "a", Name: "A", Selector: "alpha/a"}, {Provider: "beta", ID: "b", Name: "B", Selector: "beta/b"}}}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(600).Height(500).TextColor(theme.Text).Children(func() { a.canvasModelDialog(c, hc, "omp", proto.SessionInfo{}, false) })
	}, 600, 500)
	tt.Frame()
	tt.Advance(time.Second)
	if err := tt.Click("Registered provider"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Registered provider: beta"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if hc.providerDraft != "beta" || hc.modelDraft != "b" {
		t.Fatalf("provider retained another provider's model: %s/%s", hc.providerDraft, hc.modelDraft)
	}
	if canvasModelRegistered(hc.models, "beta/a") || canvasModelRegistered(hc.models, "unknown/model") {
		t.Fatal("unregistered selection accepted")
	}
}
