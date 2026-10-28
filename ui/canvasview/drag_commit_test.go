package canvasview

import (
	"errors"
	"testing"

	"github.com/enowx/enowx/proto"
)

func TestDragKeepsPreviewUntilRemoteReply(t *testing.T) {
	for _, fail := range []bool{false, true} {
		v := &View{frame: "f", canvas: &proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "c", Version: 1}, Frames: []proto.CanvasFrame{{ID: "f", Elements: []proto.CanvasElement{{ID: "shape", Type: proto.ShapeRect, X: 20, Y: 20, W: 100, H: 60}}}}}}
		v.drag = &dragState{id: "shape", x: 240, y: 180, w: 100, h: 60, moved: true}
		var reply func(proto.Canvas, error)
		sent := 0
		o := Options{Edit: func(e proto.CanvasEdit, done func(proto.Canvas, error)) { sent++; reply = done }}
		shape := v.canvas.Frames[0].Elements[0]
		moved := shape
		moved.X, moved.Y = 240, 180
		v.commitDrag(o, v.drag, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &moved})
		if sent != 1 || v.drag == nil || !v.drag.pending {
			t.Fatal("drag preview cleared before acknowledgement")
		}
		for range 5 {
			if got := v.placed(shape); got.X != 240 || got.Y != 180 {
				t.Fatalf("preview snapped back: %+v", got)
			}
		}
		if fail {
			reply(proto.Canvas{}, errors.New("host disconnected"))
			if v.err == "" {
				t.Fatal("write error hidden")
			}
		} else {
			cv := *v.canvas
			cv.Version++
			cv.Frames = append([]proto.CanvasFrame(nil), cv.Frames...)
			cv.Frames[0].Elements = []proto.CanvasElement{moved}
			reply(cv, nil)
		}
		if v.drag != nil {
			t.Fatal("pending preview did not clear on reply")
		}
		got := v.placed(v.canvas.Frames[0].Elements[0])
		if !fail && got.X != 240 {
			t.Fatal("acknowledged position lost")
		}
		if fail && got.X != 20 {
			t.Fatal("failed edit did not restore host geometry")
		}
	}
}
