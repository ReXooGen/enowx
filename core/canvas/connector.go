package canvas

import (
	"github.com/enowx/enowx/proto"
	"math"
)

// ConnectorPoint is a point on a connector in frame coordinates.
type ConnectorPoint struct{ X, Y float64 }

// ConnectorPath follows connected shapes, using right-angle segments when
// requested. Cardinal ports keep the first and last segments outside shapes.
func ConnectorPath(f proto.CanvasFrame, el proto.CanvasElement) []ConnectorPoint {
	x1, y1, x2, y2 := Ends(f, el)
	if el.Routing == "" || el.Routing == "straight" {
		return []ConnectorPoint{{x1, y1}, {x2, y2}}
	}
	from, hasFrom := byID(f, el.From)
	to, hasTo := byID(f, el.To)
	x1, y1, x2, y2 = el.X, el.Y, el.X+el.W, el.Y+el.H
	if hasFrom {
		x1, y1 = from.X+from.W/2, from.Y+from.H/2
	}
	if hasTo {
		x2, y2 = to.X+to.W/2, to.Y+to.H/2
	}
	horizontal := math.Abs(x2-x1) >= math.Abs(y2-y1)
	if el.Routing == "elbow-x" {
		horizontal = true
	}
	if el.Routing == "elbow-y" {
		horizontal = false
	}
	var points []ConnectorPoint
	if horizontal {
		direction := 1.0
		if x2 < x1 {
			direction = -1
		}
		if hasFrom {
			x1 += direction * from.W / 2
		}
		if hasTo {
			x2 -= direction * to.W / 2
		}
		mid := (x1 + x2) / 2
		points = []ConnectorPoint{{x1, y1}, {mid, y1}, {mid, y2}, {x2, y2}}
	} else {
		direction := 1.0
		if y2 < y1 {
			direction = -1
		}
		if hasFrom {
			y1 += direction * from.H / 2
		}
		if hasTo {
			y2 -= direction * to.H / 2
		}
		mid := (y1 + y2) / 2
		points = []ConnectorPoint{{x1, y1}, {x1, mid}, {x2, mid}, {x2, y2}}
	}
	// Drop empty segments so arrowheads always use the final visible segment.
	out := points[:1]
	for _, p := range points[1:] {
		if p != out[len(out)-1] {
			out = append(out, p)
		}
	}
	if len(out) == 1 {
		out = append(out, out[0])
	}
	return out
}

// ConnectorMiddle places a label halfway along the visible route.
func ConnectorMiddle(points []ConnectorPoint) (x, y float64) {
	var length float64
	for i := 1; i < len(points); i++ {
		length += math.Hypot(points[i].X-points[i-1].X, points[i].Y-points[i-1].Y)
	}
	remaining := length / 2
	for i := 1; i < len(points); i++ {
		a, b := points[i-1], points[i]
		d := math.Hypot(b.X-a.X, b.Y-a.Y)
		if d > 0 && remaining <= d {
			return a.X + (b.X-a.X)*remaining/d, a.Y + (b.Y-a.Y)*remaining/d
		}
		remaining -= d
	}
	if len(points) > 0 {
		return points[0].X, points[0].Y
	}
	return 0, 0
}
