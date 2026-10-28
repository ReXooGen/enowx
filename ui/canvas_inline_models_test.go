package ui

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/chatview"
	"github.com/enowx/enowx/ui/theme"
	"testing"
	"time"
)

func TestCanvasComposerOffersTwoRegisteredDropdowns(t *testing.T) {
	a := New(Options{})
	a.canvasChat = &hostCanvasChat{canvas: "c", modelRequest: 1, models: []proto.AgentModel{{Provider: "custom", ID: "one", Selector: "custom/one", Name: "Model One"}, {Provider: "custom", ID: "two", Selector: "custom/two", Name: "Model Two"}}}
	ch := chatview.New()
	controls := a.canvasInlineModels("c", "omp", proto.SessionInfo{}, false)
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(414).Height(500).Background(theme.Background).TextColor(theme.Text).Children(func() {
			ch.View(c, chatview.Options{Agent: "omp", AgentName: "OMP", Fresh: true, ModelControls: controls})
		})
	}, 414, 500)
	tt.Frame()
	tt.Advance(time.Second)
	if err := tt.Click("Canvas provider"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Canvas provider: custom"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Canvas model"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Canvas model: two"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.notesModel["omp"] != "custom/two" {
		t.Fatalf("dropdown choice not applied: %q", a.notesModel["omp"])
	}
	if tt.HasText("Model / provider") {
		t.Fatal("old modal button is still in the composer")
	}
	shot(t, tt, "canvas-inline-models.png")
}
