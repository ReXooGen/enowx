package canvasview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestVariationArrowsStayAboveBackgroundSheet(t *testing.T) {
	f := proto.CanvasFrame{ID: "variation", Name: "Variation", Width: 800, Height: 600, Elements: []proto.CanvasElement{
		{ID: "sheet", Type: proto.ShapeRect, X: 20, Y: 20, W: 760, H: 560, Fill: "#ffffff", Stroke: "none"},
		{ID: "a", Type: proto.ShapeRect, X: 100, Y: 200, W: 120, H: 60},
		{ID: "b", Type: proto.ShapeEllipse, X: 500, Y: 200, W: 120, H: 60},
		{ID: "arrow", Type: proto.ShapeArrow, From: "a", To: "b", Stroke: "#0000ff", StrokeWidth: 4, Routing: "elbow"},
	}}
	v := &View{canvas: &proto.Canvas{Frames: []proto.CanvasFrame{f}}, frame: f.ID}
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Width(900).Height(700).Children(func() { v.stage(c, Options{}) }) }, 900, 700)
	tt.Frame()
	tt.Frame()
	a, _ := tt.Find("a")
	b, _ := tt.Find("b")
	pixel := tt.Image().RGBAAt(int((a.X+a.W+b.X)/2), int(a.Y+a.H/2))
	if pixel.B < 180 || pixel.R > 80 {
		t.Fatalf("background sheet covers connector: %+v", pixel)
	}
}
