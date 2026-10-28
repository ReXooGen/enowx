package canvas

import (
	"fmt"
	"math"
	"slices"

	"github.com/enowx/enowx/proto"
)

func batchElements(f *proto.CanvasFrame, e proto.CanvasEdit) error {
	if len(e.Targets) == 0 {
		return fmt.Errorf("canvas: select shapes first")
	}
	var indices []int
	seen := map[string]bool{}
	for _, id := range e.Targets {
		if seen[id] {
			continue
		}
		seen[id] = true
		i := slices.IndexFunc(f.Elements, func(el proto.CanvasElement) bool { return el.ID == id })
		if i < 0 {
			return fmt.Errorf("canvas: no shape %s", id)
		}
		if f.Elements[i].Locked && e.Action != "unlock" {
			return fmt.Errorf("canvas: shape %s is locked", id)
		}
		indices = append(indices, i)
	}
	if math.IsNaN(e.DX) || math.IsInf(e.DX, 0) || math.IsNaN(e.DY) || math.IsInf(e.DY, 0) {
		return fmt.Errorf("canvas: movement must be finite")
	}
	left, top, right, bottom := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, i := range indices {
		x := f.Elements[i]
		left = min(left, x.X)
		top = min(top, x.Y)
		right = max(right, x.X+x.W)
		bottom = max(bottom, x.Y+x.H)
	}
	group := "g" + newID(4)
	switch e.Action {
	case "move", "left", "right", "top", "bottom", "center-x", "center-y", "group", "ungroup", "lock", "unlock":
		for _, i := range indices {
			x := &f.Elements[i]
			switch e.Action {
			case "move":
				x.X += e.DX
				x.Y += e.DY
			case "left":
				x.X = left
			case "right":
				x.X = right - x.W
			case "top":
				x.Y = top
			case "bottom":
				x.Y = bottom - x.H
			case "center-x":
				x.X = (left + right - x.W) / 2
			case "center-y":
				x.Y = (top + bottom - x.H) / 2
			case "group":
				x.Group = group
			case "ungroup":
				x.Group = ""
			case "lock":
				x.Locked = true
			case "unlock":
				x.Locked = false
			}
		}
	case "distribute-x", "distribute-y":
		if len(indices) < 3 {
			return fmt.Errorf("canvas: distribution needs three shapes")
		}
		horizontal := e.Action == "distribute-x"
		slices.SortFunc(indices, func(a, b int) int {
			av, bv := f.Elements[a].Y, f.Elements[b].Y
			if horizontal {
				av, bv = f.Elements[a].X, f.Elements[b].X
			}
			if av < bv {
				return -1
			}
			if av > bv {
				return 1
			}
			return 0
		})
		span := bottom - top
		start := top
		if horizontal {
			span = right - left
			start = left
		}
		for _, i := range indices {
			if horizontal {
				span -= f.Elements[i].W
			} else {
				span -= f.Elements[i].H
			}
		}
		gap := span / float64(len(indices)-1)
		for _, i := range indices {
			x := &f.Elements[i]
			if horizontal {
				x.X = start
				start += x.W + gap
			} else {
				x.Y = start
				start += x.H + gap
			}
		}
	case "delete":
		f.Elements = slices.DeleteFunc(f.Elements, func(el proto.CanvasElement) bool { return seen[el.ID] || seen[el.From] || seen[el.To] })
	default:
		return fmt.Errorf("canvas: no batch action %q", e.Action)
	}
	return nil
}
