package canvasview

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/proto"
)

// A diagonal connector must paint between its endpoints, including while a
// connected shape is being dragged; the axis-aligned painter loses that path.
func TestDiagonalConnectorsFollowShapes(t *testing.T) {
	for _, typ := range []string{proto.ShapeLine, proto.ShapeArrow} {
		t.Run(typ, func(t *testing.T) {
			f := proto.CanvasFrame{ID: "frame", Width: 600, Height: 400, Elements: []proto.CanvasElement{
				{ID: "from", Type: proto.ShapeRect, X: 40, Y: 40, W: 120, H: 60},
				{ID: "to", Type: proto.ShapeEllipse, X: 400, Y: 260, W: 120, H: 60},
				{ID: "connector", Type: typ, From: "from", To: "to", Stroke: "#ffffff", StrokeWidth: 4},
			}}
			v := &View{canvas: &proto.Canvas{Frames: []proto.CanvasFrame{f}}}
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Width(600).Height(400).Background(ui.RGB(0, 0, 0)).Children(func() { v.arrows(c, f, 1) })
			}, 600, 400)
			check := func(name string) {
				t.Helper()
				tt.Frame()
				shown := f
				shown.Elements = append([]proto.CanvasElement(nil), f.Elements...)
				for i := range shown.Elements {
					shown.Elements[i] = v.placed(shown.Elements[i])
				}
				x1, y1, x2, y2 := canvas.Ends(shown, shown.Elements[2])
				img := tt.Image()
				for _, fraction := range []float64{.1, .5, .9} {
					x, y := int(x1+(x2-x1)*fraction), int(y1+(y2-y1)*fraction)
					if !connectorPixel(img, x, y) {
						t.Fatalf("%s: line missing at (%d,%d)", name, x, y)
					}
				}
				if connectorPixel(img, int((x1+x2)/2), int(y1)) {
					t.Fatalf("%s: stray horizontal line", name)
				}
				if dir := os.Getenv("CANVAS_QA_DIR"); dir != "" {
					if err := os.MkdirAll(dir, 0755); err != nil {
						t.Fatal(err)
					}
					file, err := os.Create(filepath.Join(dir, typ+"-"+name+".png"))
					if err != nil {
						t.Fatal(err)
					}
					err = png.Encode(file, img)
					file.Close()
					if err != nil {
						t.Fatal(err)
					}
				}
			}
			check("initial")
			v.drag = &dragState{id: "to", x: 330, y: 180, w: 120, h: 60}
			check("dragged")
		})
	}
}

func connectorPixel(img *image.RGBA, x, y int) bool {
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			pixel := img.RGBAAt(x+dx, y+dy)
			if pixel.R > 180 && pixel.G > 180 && pixel.B > 180 {
				return true
			}
		}
	}
	return false
}
