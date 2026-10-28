package canvasview

import "github.com/enowx/enowx/proto"

// Keep the preview until the reply is applied on the main thread. Clearing it
// on pointer-up briefly restores the host's old position, especially over SSH.
func (v *View) commitDrag(o Options, drag *dragState, edit proto.CanvasEdit) {
	drag.pending = true
	send := o.Edit
	o.Edit = func(e proto.CanvasEdit, done func(proto.Canvas, error)) {
		send(e, func(cv proto.Canvas, err error) {
			if v.drag == drag {
				v.drag = nil
			}
			done(cv, err)
		})
	}
	v.edit(o, edit)
}
