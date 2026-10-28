package canvasview

import "github.com/enowx/enowx/proto"

func (v *View) chooseShape(id string, add bool) {
	if !add {
		v.chosen = map[string]bool{}
	}
	if v.chosen == nil {
		v.chosen = map[string]bool{}
	}
	el, ok := shapeOf(v.frameOf(), id)
	if !ok {
		v.chosen = nil
		v.sel = id
		return
	}
	selected := !v.chosen[id]
	for _, candidate := range v.frameOf().Elements {
		if candidate.ID == id || (el.Group != "" && candidate.Group == el.Group) {
			v.chosen[candidate.ID] = selected
		}
	}
	v.sel = id
	if !selected {
		v.sel = ""
		for _, candidate := range v.frameOf().Elements {
			if v.chosen[candidate.ID] {
				v.sel = candidate.ID
				break
			}
		}
	}
}

func (v *View) selectedShapes() []string {
	var ids []string
	for _, el := range v.frameOf().Elements {
		if v.chosen[el.ID] || el.ID == v.sel {
			ids = append(ids, el.ID)
		}
	}
	return ids
}

func (v *View) selectAllShapes() {
	v.chosen = map[string]bool{}
	for _, el := range v.frameOf().Elements {
		v.chosen[el.ID] = true
		v.sel = el.ID
	}
}

func (v *View) batch(o Options, action string) {
	v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpBatch, Targets: v.selectedShapes(), Action: action})
}
