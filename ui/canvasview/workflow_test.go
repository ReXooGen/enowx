package canvasview

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
)

func TestCanvasWorkflowPanelsAndMotion(t *testing.T) {
	v := &View{canvas: &proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "workflow", Title: "Checkout"}, Frames: []proto.CanvasFrame{{ID: "home", Name: "Home", Width: 800, Height: 600, Elements: []proto.CanvasElement{{ID: "cta", Type: proto.ShapeRect, X: 30, Y: 40, W: 130, H: 48, Text: "Continue"}}}, {ID: "done", Name: "Done", Width: 800, Height: 600}}}, selected: "workflow", sel: "cta", panel: "workflow", mode: modeDesign}
	o := Options{Library: func(p proto.CanvasLibraryParams, done func(proto.CanvasLibraryResult, error)) {
		done(proto.CanvasLibraryResult{}, nil)
	}}
	v.openWorkflow(o)
	var edit proto.CanvasEdit
	o.Edit = func(e proto.CanvasEdit, done func(proto.Canvas, error)) { edit = e; done(*v.canvas, nil) }
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(1280).Height(800).Background(theme.Background).TextColor(theme.Text).Children(func() { v.page(c, o) })
	}, 1280, 800)
	for _, name := range []string{"Library", "Review", "Checks", "Arrange", "Motion", "Offline"} {
		tt.Frame()
		if err := tt.Click(name); err != nil {
			t.Fatal(err)
		}
		tt.Advance(time.Second)
		tt.Frame()
		if dir := os.Getenv("CANVAS_QA_DIR"); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			file, err := os.Create(filepath.Join(dir, name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			err = png.Encode(file, tt.Image())
			file.Close()
			if err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := tt.Click("Motion"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(time.Second)
	if err := tt.Click("Target design"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Target design: done · Done"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Apply transition"); err != nil {
		t.Fatal(err)
	}
	if edit.Op != proto.CanvasOpTransition || edit.Transition == nil || edit.Transition.Target != "done" {
		t.Fatalf("transition not submitted: %+v", edit)
	}
}
