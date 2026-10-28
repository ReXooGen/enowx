package chatview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/ui/theme"
	"testing"
)

func TestOMPComposerOffersModelProviderButton(t *testing.T) {
	ch := New()
	opened := false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(414).Height(400).Background(theme.Background).TextColor(theme.Text).Children(func() {
			ch.View(c, Options{Agent: "omp", AgentName: "OMP", Fresh: true, ChooseModel: func() { opened = true }})
		})
	}, 414, 400)
	tt.Frame()
	if err := tt.Click("Model / provider"); err != nil {
		t.Fatal(err)
	}
	if !opened {
		t.Fatal("model picker did not open from the narrow composer")
	}
}
