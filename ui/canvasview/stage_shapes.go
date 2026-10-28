package canvasview

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
)

// recent says element id is the one an agent changed moments ago.
func (v *View) recent(id string) bool {
	return id != "" && v.last.Element == id && time.Since(v.lastAt) < highlightFor
}

// placed is where shape el shows: as it is, or as being dragged.
func (v *View) placed(el proto.CanvasElement) proto.CanvasElement {
	if d := v.drag; d != nil && !d.resize && d.id != el.ID && v.chosen[el.ID] {
		if anchor, ok := shapeOf(v.frameOf(), d.id); ok {
			el.X += d.x - anchor.X
			el.Y += d.y - anchor.Y
		}
	}
	if d := v.drag; d != nil && d.id == el.ID {
		el.X, el.Y, el.W, el.H = d.x, d.y, d.w, d.h
		el = snapped(v.frameOf(), el) // where it goes, on the grid
	}
	return el
}

// shape draws a shape and handles the pointer on it.
func (v *View) shape(c *ui.Context, o Options, el proto.CanvasElement, s float64, now time.Time) {
	shown := v.placed(el)
	st := canvas.StyleOf(shown)
	box := ui.Box(c).Absolute().Left(float32(shown.X*s)).Top(float32(shown.Y*s)).Size(float32(shown.W*s), float32(shown.H*s)).
		Label(el.ID)
	fill, stroke := color(st.Fill, ui.Color{}), color(st.Stroke, ui.Color{})
	switch el.Type {
	case proto.ShapeEllipse, proto.ShapeDiamond:
		typ, width := el.Type, float32(st.StrokeWidth*s)
		box.Draw(func(p *ui.Painter, r ui.Rect) {
			var path ui.Path
			if typ == proto.ShapeEllipse {
				ellipsePath(&path, r)
			} else {
				path.MoveTo(r.X+r.W/2, r.Y).LineTo(r.X+r.W, r.Y+r.H/2).LineTo(r.X+r.W/2, r.Y+r.H).LineTo(r.X, r.Y+r.H/2).Close()
			}
			if st.Fill != "" {
				p.FillPath(&path, fill)
			}
			if st.Stroke != "" {
				p.StrokePath(&path, width, stroke)
			}
		})
	case proto.ShapeImage:
		box.Background(ui.RGB(226, 232, 240)).Radius(float32(el.Radius * s))
	default:
		if st.Fill != "" {
			box.Background(fill)
		}
		if st.Stroke != "" {
			box.Border(float32(math.Max(st.StrokeWidth*s, 1)), stroke)
		}
		box.Radius(float32(el.Radius * s))
		if el.Type == proto.ShapeSticky {
			box.Shadow(0, 1, 4, 0, ui.RGBA(15, 23, 42, 0.15))
		}
	}
	selected := v.sel == el.ID || v.chosen[el.ID]
	if v.recent(el.ID) {
		box.Shadow(0, 0, 0, 3, theme.Accent)
	} else if selected {
		box.Shadow(0, 0, 0, 2, ui.RGB(99, 102, 241))
	} else if v.connect != "" && box.Hovered() {
		box.Shadow(0, 0, 0, 2, ui.RGBA(99, 102, 241, 0.5))
	}
	if v.mode == modeDesign {
		box.Cursor(ui.CursorPointer)
		v.handlePointer(c, o, box, el, s)
	} else {
		v.transitionPointer(c, o, box, el.ID)
	}
	box.Children(func() {
		switch {
		case v.editing != nil && v.editing.id == el.ID:
			in := ui.TextInput(c.Key("edit"), &v.editing.text).FillWidth().FillHeight().FontSize(float32(st.FontSize * s)).
				TextColor(color(st.Color, ui.RGB(15, 23, 42))).Padding(4).AutoFocus().Label("Text of " + el.ID)
			if in.Submitted() || in.Shortcut(0, ui.KeyEscape) || !in.Focused() && v.editing.text != el.Text {
				el.Text = v.editing.text
				v.editing = nil
				v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &el})
			}
		case el.Type == proto.ShapeImage:
			ui.Column(c).FillWidth().FillHeight().Center().Gap(4).Children(func() {
				ui.Icon(c, icon.File).FontSize(float32(math.Max(14*s, 10))).TextColor(ui.RGB(100, 116, 139))
				ui.Text(c, el.Src).FontSize(float32(math.Max(11*s, 9))).TextColor(ui.RGB(100, 116, 139)).SingleLine()
			})
		case el.Text != "":
			shapeText(c, shown, st, s)
		}
		if selected && !el.Locked && len(v.selectedShapes()) <= 1 && v.mode == modeDesign && v.editing == nil {
			v.resizeHandle(c, o, el, s)
		}
		if v.recent(el.ID) && v.last.Session != "" {
			who := "Agent"
			if o.AgentName != nil {
				who = o.AgentName(v.last.Session)
			}
			ui.Row(c.Key("who")).Absolute().Left(0).Top(-22).Height(18).PaddingX(6).Radius(4).Background(theme.Accent).Children(func() {
				ui.Text(c, who).FontSize(10.5).TextColor(ui.RGB(255, 255, 255)).SingleLine()
			})
		}
	})
	_ = now
}

func shapeText(c *ui.Context, el proto.CanvasElement, st canvas.Style, s float64) {
	align := ui.Center
	switch el.Align {
	case "left":
		align = ui.Start
	case "right":
		align = ui.End
	}
	ui.Column(c).FillWidth().FillHeight().Padding(float32(8 * s)).Justify(ui.Center).AlignItems(align).Children(func() {
		t := ui.Text(c, el.Text).FontSize(float32(st.FontSize * s)).TextColor(color(st.Color, ui.RGB(15, 23, 42))).TextAlign(align)
		if el.Bold {
			t.FontWeight(700)
		}
	})
}

func ellipsePath(p *ui.Path, r ui.Rect) {
	const k = 0.5523
	cx, cy, rx, ry := r.X+r.W/2, r.Y+r.H/2, r.W/2, r.H/2
	p.MoveTo(cx+rx, cy).
		CubeTo(cx+rx, cy+ry*k, cx+rx*k, cy+ry, cx, cy+ry).
		CubeTo(cx-rx*k, cy+ry, cx-rx, cy+ry*k, cx-rx, cy).
		CubeTo(cx-rx, cy-ry*k, cx-rx*k, cy-ry, cx, cy-ry).
		CubeTo(cx+rx*k, cy-ry, cx+rx, cy-ry*k, cx+rx, cy).Close()
}

// handlePointer selects shape el on a click, connects an arrow to it, edits
// its text on a double click, and moves it as it is dragged.
func (v *View) handlePointer(c *ui.Context, o Options, box ui.Element, el proto.CanvasElement, s float64) {
	if v.drag != nil && v.drag.pending {
		return
	}
	if el.Locked {
		if box.Clicked() {
			v.chooseShape(el.ID, box.ClickModifiers()&ui.Shift != 0)
		}
		return
	}
	if dx, dy, ok := box.Dragged(); ok {
		if v.drag == nil || v.drag.id != el.ID {
			if !v.chosen[el.ID] {
				v.chooseShape(el.ID, false)
			}
			v.drag = &dragState{id: el.ID, x: el.X, y: el.Y, w: el.W, h: el.H}
		}
		if (dx != 0 || dy != 0) && !v.drag.resize {
			v.drag.x += float64(dx) / s
			v.drag.y += float64(dy) / s
			v.drag.moved = true
		}
	} else if d := v.drag; d != nil && d.id == el.ID && !d.resize {
		if d.moved {
			if len(v.selectedShapes()) > 1 {
				v.commitDrag(o, d, proto.CanvasEdit{Op: proto.CanvasOpBatch, Targets: v.selectedShapes(), Action: "move", DX: math.Round(d.x - el.X), DY: math.Round(d.y - el.Y)})
			} else {
				el.X, el.Y = math.Round(d.x), math.Round(d.y)
				v.commitDrag(o, d, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &el})
			}
		} else {
			v.drag = nil
		}
	}
	if box.DoubleClicked() && el.Type != proto.ShapeImage {
		v.sel, v.editing = el.ID, &textEdit{id: el.ID, text: el.Text}
		return
	}
	if box.Clicked() {
		if v.connect != "" && v.connect != el.ID {
			arrow := proto.CanvasElement{Type: proto.ShapeArrow, From: v.connect, To: el.ID, Routing: "elbow"}
			v.connect = ""
			v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &arrow})
			return
		}
		v.chooseShape(el.ID, box.ClickModifiers()&ui.Shift != 0)
		if v.panel != "workflow" {
			v.panel = "layers"
		}
	}
}

// resizeHandle is the selected shape's corner, which drags its size.
func (v *View) resizeHandle(c *ui.Context, o Options, el proto.CanvasElement, s float64) {
	h := ui.Box(c.Key("resize")).Absolute().Right(-5).Bottom(-5).Size(10, 10).Radius(2).
		Background(ui.RGB(255, 255, 255)).Border(1.5, ui.RGB(99, 102, 241)).Cursor(ui.CursorResizeNWSE).Label("Resize " + el.ID)
	if v.drag != nil && v.drag.pending {
		return
	}
	if dx, dy, ok := h.Dragged(); ok {
		if v.drag == nil || v.drag.id != el.ID || !v.drag.resize {
			v.drag = &dragState{id: el.ID, resize: true, x: el.X, y: el.Y, w: el.W, h: el.H}
		}
		v.drag.w = math.Max(v.drag.w+float64(dx)/s, 8)
		v.drag.h = math.Max(v.drag.h+float64(dy)/s, 8)
		v.drag.moved = v.drag.moved || dx != 0 || dy != 0
	} else if d := v.drag; d != nil && d.id == el.ID && d.resize {
		if d.moved {
			el.W, el.H = math.Round(d.w), math.Round(d.h)
			v.commitDrag(o, d, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &el})
		} else {
			v.drag = nil
		}
	}
}

// arrows draws the frame's lines and arrows, following the shapes they
// join as those move.
func (v *View) arrows(c *ui.Context, f proto.CanvasFrame, s float64) {
	shown := f
	shown.Elements = make([]proto.CanvasElement, len(f.Elements))
	for i, el := range f.Elements {
		shown.Elements[i] = v.placed(el)
	}
	ui.Box(c.Key("arrows")).Absolute().Left(0).Top(0).FillWidth().FillHeight().PassThrough().Draw(func(p *ui.Painter, r ui.Rect) {
		for _, el := range shown.Elements {
			if el.Type != proto.ShapeLine && el.Type != proto.ShapeArrow {
				continue
			}
			st := canvas.StyleOf(el)
			col := color(st.Stroke, ui.RGB(71, 85, 105))
			if v.sel == el.ID || v.recent(el.ID) {
				col = ui.RGB(99, 102, 241)
			}
			points := canvas.ConnectorPath(shown, el)
			w := float32(math.Max(st.StrokeWidth*s, 1))
			// Painter.Line only draws axis-aligned lines; connectors can be diagonal.
			var shaft ui.Path
			shaft.MoveTo(r.X+float32(points[0].X*s), r.Y+float32(points[0].Y*s))
			for _, point := range points[1:] {
				shaft.LineTo(r.X+float32(point.X*s), r.Y+float32(point.Y*s))
			}
			p.StrokePath(&shaft, w, col)
			if el.Type == proto.ShapeArrow {
				a, b := points[len(points)-2], points[len(points)-1]
				x1, y1, x2, y2 := a.X, a.Y, b.X, b.Y
				ax, ay, bx, by := canvas.ArrowHead(x1, y1, x2, y2, 8+st.StrokeWidth*2)
				var head ui.Path
				head.MoveTo(r.X+float32(x2*s), r.Y+float32(y2*s)).LineTo(r.X+float32(ax*s), r.Y+float32(ay*s)).
					LineTo(r.X+float32(bx*s), r.Y+float32(by*s)).Close()
				p.FillPath(&head, col)
			}
		}
	})
}

// arrowHandle is an arrow's label at its middle, which selects it.
func (v *View) arrowHandle(c *ui.Context, o Options, f proto.CanvasFrame, el proto.CanvasElement, s float64) {
	shown := f
	shown.Elements = make([]proto.CanvasElement, len(f.Elements))
	for i, x := range f.Elements {
		shown.Elements[i] = v.placed(x)
	}
	mx, my := canvas.ConnectorMiddle(canvas.ConnectorPath(shown, el))
	mx, my = mx*s, my*s
	st := canvas.StyleOf(el)
	label := ui.Row(c).Absolute().Left(float32(mx) - 60).Top(float32(my) - 11).Width(120).Height(22).Justify(ui.Center).AlignItems(ui.Center).Label(el.ID)
	if v.mode == modeDesign {
		label.Cursor(ui.CursorPointer)
		if label.Clicked() {
			v.sel, v.panel = el.ID, "layers"
		}
	}
	label.Children(func() {
		if el.Text == "" {
			return
		}
		ui.Text(c, el.Text).FontSize(float32(st.FontSize * s * 0.85)).TextColor(color(st.Color, ui.RGB(15, 23, 42))).
			Background(ui.RGBA(255, 255, 255, 0.9)).PaddingX(4).Radius(3).SingleLine()
	})
	_ = o
}
