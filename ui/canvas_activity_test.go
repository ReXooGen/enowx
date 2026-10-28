package ui

import (
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestCanvasActivityShowsBeforeFirstCanvasEdit(t *testing.T) {
	h := newCanvasHarness(t)
	h.app.agents = []proto.AgentInfo{{ID: "omp", Name: "OMP", Installed: true}}
	cv, err := h.cl.CreateCanvas(proto.CanvasCreateParams{Title: "Activity"})
	if err != nil {
		t.Fatal(err)
	}
	h.settle("canvas open", func() bool { return h.app.canvases.Selected() == cv.ID && h.app.canvases.Canvas() != nil })
	h.app.sessions = append(h.app.sessions, proto.SessionInfo{ID: "canvas-worker", Agent: "omp", CanvasFor: cv.ID, State: proto.StateWorking})
	h.tt.Frame()
	if !h.tt.HasText("OMP is working") {
		t.Fatalf("working session has no pill before its first edit: %q", h.tt.Texts())
	}
	shot(t, h.tt, "canvas-working-pill.png")
	h.app.sessions[len(h.app.sessions)-1].State = proto.StateWaiting
	h.tt.Frame()
	if !h.tt.HasText("OMP needs input") {
		t.Fatal("waiting state did not reach pill")
	}
	h.app.sessions[len(h.app.sessions)-1].State = proto.StateDone
	h.tt.Frame()
	if _, ok := h.tt.Find("Canvas agent status"); ok {
		t.Fatal("activity pill remains after work is done")
	}
}
