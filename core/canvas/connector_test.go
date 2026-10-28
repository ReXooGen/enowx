package canvas

import (
	"strings"
	"testing"

	"github.com/enowx/enowx/proto"
)

func TestElbowConnectorFollowsShapesAndExports(t *testing.T) {
	for _, routing := range []string{"elbow", "elbow-x", "elbow-y"} {
		for _, target := range []ConnectorPoint{{400, 260}, {-200, 260}, {400, -200}, {40, 260}, {400, 40}} {
			f := proto.CanvasFrame{Width: 700, Height: 500, Elements: []proto.CanvasElement{
				{ID: "a", Type: proto.ShapeRect, X: 40, Y: 40, W: 120, H: 60},
				{ID: "b", Type: proto.ShapeEllipse, X: target.X, Y: target.Y, W: 120, H: 60},
			}}
			el := proto.CanvasElement{Type: proto.ShapeArrow, From: "a", To: "b", Routing: routing}
			points := ConnectorPath(f, el)
			for i := 1; i < len(points); i++ {
				a, b := points[i-1], points[i]
				if a.X != b.X && a.Y != b.Y {
					t.Fatalf("%s has diagonal segment: %v", routing, points)
				}
				if a == b {
					t.Fatalf("%s has empty segment: %v", routing, points)
				}
			}
			before := points[len(points)-1]
			f.Elements[1].X += 80
			after := ConnectorPath(f, el)
			if before == after[len(after)-1] {
				t.Fatal("endpoint did not follow target")
			}
		}
	}
	f := proto.CanvasFrame{Width: 600, Height: 400, Elements: []proto.CanvasElement{{ID: "a", Type: proto.ShapeRect, X: 40, Y: 40, W: 120, H: 60}, {ID: "b", Type: proto.ShapeEllipse, X: 400, Y: 260, W: 120, H: 60}, {ID: "line", Type: proto.ShapeArrow, From: "a", To: "b", Routing: "elbow-x", Stroke: "#ffffff", StrokeWidth: 4}}}
	points := ConnectorPath(f, f.Elements[2])
	if points[0] != (ConnectorPoint{160, 70}) || points[len(points)-1] != (ConnectorPoint{400, 290}) {
		t.Fatalf("ports: %v", points)
	}
	if got := SVG(f, nil, false); !strings.Contains(got, `points="160,70 280,70 280,290 400,290 `) {
		t.Fatalf("SVG lost elbow: %s", got)
	}
	img := Raster(f, nil, 1)
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		pixel := img.RGBAAt(int((a.X+b.X)/2), int((a.Y+b.Y)/2))
		if pixel.R < 200 {
			t.Fatalf("export lost segment %v → %v", a, b)
		}
	}
}
