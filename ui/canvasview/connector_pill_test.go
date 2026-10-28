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

func TestCanvasElbowRoutingAndAgentPill(t *testing.T) {
	v := &View{selected: "canvas", frame: "home", sel: "arrow", panel: "layers", width: 1100, canvas: &proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "canvas", Title: "Checkout flow", Agent: "omp", AgentAt: time.Now()}, Frames: []proto.CanvasFrame{{ID: "home", Name: "Board", Width: 800, Height: 600, Elements: []proto.CanvasElement{{ID: "a", Type: proto.ShapeRect, X: 60, Y: 80, W: 180, H: 90, Text: "Start"}, {ID: "b", Type: proto.ShapeEllipse, X: 500, Y: 360, W: 180, H: 90, Text: "Finish"}, {ID: "arrow", Type: proto.ShapeArrow, From: "a", To: "b"}}}}}}
	o := Options{CanAsk: true, AgentName: func(string) string { return "OMP" }}
	o.Edit = func(e proto.CanvasEdit, done func(proto.Canvas, error)) {
		v.canvas.Frames[0].Elements[2] = *e.Element
		done(*v.canvas, nil)
	}
	o.Control = func(id, action string, done func(error)) { v.canvas.Paused = action == proto.CanvasPause; done(nil) }
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(1100).Height(760).Background(theme.Background).TextColor(theme.Text).Children(func() { v.page(c, o) })
	}, 1100, 760)
	tt.Frame()
	if err := tt.Click("Connector routing"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Connector routing: Elbow 90°"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if v.canvas.Frames[0].Elements[2].Routing != "elbow" {
		t.Fatal("elbow choice was not saved")
	}
	pill, ok := tt.Find("Canvas agent status")
	if !ok || pill.W > 340 || pill.H > 40 || pill.Y < 90 || pill.X < 400 {
		t.Fatalf("pill is not over the upper-right canvas: %+v", pill)
	}
	if dir := os.Getenv("CANVAS_QA_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		file, err := os.Create(filepath.Join(dir, "elbow-pill.png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, tt.Image())
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tt.Click("Pause"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("Agents paused") {
		t.Fatal("pause state missing")
	}
	if err := tt.Click("Resume"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if v.canvas.Paused {
		t.Fatal("resume action was lost")
	}
}
