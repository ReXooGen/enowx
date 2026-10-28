package canvas

import (
	"fmt"
	"math"

	"github.com/enowx/enowx/proto"
)

// CheckIssue identifies a possible layout or readability problem at a width.
type CheckIssue struct {
	Element string `json:"element"`
	Message string `json:"message"`
	Width   int    `json:"width"`
}

// CheckShapes checks deterministic shape geometry; browser checks handle HTML.
func CheckShapes(f proto.CanvasFrame) []CheckIssue {
	var issues []CheckIssue
	add := func(id, msg string) { issues = append(issues, CheckIssue{Element: id, Message: msg, Width: f.Width}) }
	for i, el := range f.Elements {
		if el.Type == proto.ShapeLine || el.Type == proto.ShapeArrow {
			continue
		}
		if el.X < 0 || el.Y < 0 || el.X+el.W > float64(f.Width) || el.Y+el.H > float64(f.Height) {
			add(el.ID, "Extends beyond the frame")
		}
		st := StyleOf(el)
		if el.Text != "" {
			fg, ok1 := ParseColor(st.Color)
			bg, bgOK := ParseColor(st.Fill)
			if !bgOK {
				bg, bgOK = ParseColor(f.Background)
			}
			if ok1 && bgOK {
				lum := func(r, g, b uint8) float64 {
					channel := func(v uint8) float64 {
						x := float64(v) / 255
						if x <= .04045 {
							return x / 12.92
						}
						return math.Pow((x+.055)/1.055, 2.4)
					}
					return .2126*channel(r) + .7152*channel(g) + .0722*channel(b)
				}
				a, b := lum(fg.R, fg.G, fg.B), lum(bg.R, bg.G, bg.B)
				ratio := (max(a, b) + .05) / (min(a, b) + .05)
				if ratio < 4.5 {
					add(el.ID, fmt.Sprintf("Text contrast %.1f:1; review readability", ratio))
				}
			}
			if el.H < st.FontSize*1.2 {
				add(el.ID, "Text may be clipped vertically")
			}
			for _, other := range f.Elements[:i] {
				if other.Text != "" && other.Type == proto.ShapeText && el.Type == proto.ShapeText && el.X < other.X+other.W && el.X+el.W > other.X && el.Y < other.Y+other.H && el.Y+el.H > other.Y {
					add(el.ID, "Text overlaps "+other.ID)
				}
			}
		}
	}
	return issues
}
