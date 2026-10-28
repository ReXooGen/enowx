package canvasview

import (
	"bytes"
	"image"
	imgcolor "image/color"
	"image/png"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
)

func TestZoomedMockupUsesBackgroundWithoutOldShapes(t *testing.T) {
	cv := proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "c"}, Frames: []proto.CanvasFrame{{ID: "f", Name: "Home", Width: 800, Height: 600, HTML: `<main>Background</main>`, Elements: []proto.CanvasElement{{ID: "shape", Type: proto.ShapeRect, X: 20, Y: 20, W: 100, H: 60, Fill: "#0000ff"}}}}}
	var data bytes.Buffer
	white := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			white.SetRGBA(x, y, imgcolor.RGBA{255, 255, 255, 255})
		}
	}
	if err := png.Encode(&data, white); err != nil {
		t.Fatal(err)
	}
	backgrounds, full := 0, 0
	v := &View{canvas: &cv, frame: "f"}
	o := Options{Render: func(string, string, string, func([]byte, error)) { full++ }, RenderBackground: func(_, _, _ string, done func([]byte, error)) { backgrounds++; done(data.Bytes(), nil) }, Web: &Web{NewPage: func(done func(Page, error)) { done(livePage{}, nil) }, Update: func(fn func()) { fn() }}}
	o.Edit = func(proto.CanvasEdit, func(proto.Canvas, error)) {}
	tt := ui.NewTester(func(c *ui.Context) { ui.Column(c).Width(800).Height(600).Children(func() { v.stage(c, o) }) }, 800, 600)
	tt.Frame()
	tt.Frame()
	v.cam.zoomAt(0, 0, .25/v.cam.z)
	v.drag = &dragState{id: "shape", x: 400, y: 300, w: 100, h: 60, moved: true, pending: true}
	tt.Frame()
	tt.Frame()
	if full != 0 || backgrounds != 1 {
		t.Fatalf("wrong preview source: full=%d background=%d", full, backgrounds)
	}
	box, ok := tt.Find("shape")
	if !ok {
		t.Fatal("shape not rendered")
	}
	img := tt.Image()
	blue := func(x, y int) bool { p := img.RGBAAt(x, y); return p.B > 180 && p.R < 60 && p.G < 60 }
	x, y := int(box.X+box.W/2), int(box.Y+box.H/2)
	if !blue(x, y) {
		t.Fatal("dragged shape is missing")
	}
	oldX, oldY := x-int(380*v.cam.z), y-int(280*v.cam.z)
	if blue(oldX, oldY) {
		t.Fatal("old shape remains in the background during drag")
	}
	if len(v.backgrounds) != 1 {
		t.Fatal("missing independent background cache")
	}
}
