package ui

import (
	"strings"
	"testing"
	"time"

	"github.com/enowx/enowx/proto"
)

func TestCanvasWorkflowArrangeAndLibraryThroughDaemon(t *testing.T) {
	h := newCanvasHarness(t)
	cv, err := h.cl.CreateCanvas(proto.CanvasCreateParams{Title: "Workflow UI", Kind: proto.CanvasWireframe, Frames: []proto.CanvasFrame{{Name: "Desktop", Width: 800, Height: 500, Elements: []proto.CanvasElement{{ID: "a", Type: proto.ShapeRect, X: 20, Y: 100, W: 100, H: 60, Text: "First"}, {ID: "b", Type: proto.ShapeRect, X: 200, Y: 160, W: 100, H: 60, Text: "Second"}}}}})
	if err != nil {
		t.Fatal(err)
	}
	h.settle("canvas", func() bool { return h.app.canvases.Selected() == cv.ID && h.tt.HasText("First") })
	click := func(label string) {
		t.Helper()
		if err := h.tt.Click(label); err != nil {
			t.Fatalf("%s: %v; %q", label, err, h.tt.Texts())
		}
		h.tt.Advance(time.Second)
		h.tt.Frame()
	}
	click("Canvas tools")
	click("Arrange")
	click("Select all shapes")
	click("left")
	h.settle("aligned shapes", func() bool { c, _ := h.cl.Canvas(cv.ID); return c.Frames[0].Elements[0].X == c.Frames[0].Elements[1].X })
	click("group")
	h.settle("grouped", func() bool { c, _ := h.cl.Canvas(cv.ID); return c.Frames[0].Elements[0].Group != "" })
	click("lock")
	h.settle("locked", func() bool { c, _ := h.cl.Canvas(cv.ID); return c.Frames[0].Elements[0].Locked })
	click("unlock")
	h.settle("unlocked", func() bool { c, _ := h.cl.Canvas(cv.ID); return !c.Frames[0].Elements[0].Locked })
	click("Checks")
	click("Check this design")
	if !h.tt.HasText("Checking three preview widths") && !h.tt.HasText("findings") {
		t.Fatalf("checks are not available: %q", h.tt.Texts())
	}
	click("Library")
	if _, err := h.cl.EditCanvas(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpHTMLSet, HTML: `<button id="cta" style="color:white;background:#245BC1">Start</button>`}); err != nil {
		t.Fatal(err)
	}
	lib, err := h.cl.CanvasLibrary(proto.CanvasLibraryParams{Op: "save", Canvas: cv.ID, Frame: cv.Frames[0].ID, Target: "cta", Component: proto.CanvasLibraryComponent{Name: "CTA"}})
	if err != nil {
		t.Fatal(err)
	}
	// Reopen to read the host's newly saved component through the same RPC.
	click("Canvas tools")
	click("Canvas tools")
	h.settle("library item", func() bool { return h.tt.HasText("CTA") })
	click("Insert component CTA")
	h.settle("linked instance", func() bool {
		c, _ := h.cl.Canvas(cv.ID)
		return strings.Contains(c.Frames[0].HTML, lib.Components[0].ID)
	})
}

func TestCanvasWorkflowReviewUsesHTMLIDs(t *testing.T) {
	h := newCanvasHarness(t)
	cv, err := h.cl.CreateCanvas(proto.CanvasCreateParams{Title: "Review UI", Frames: []proto.CanvasFrame{{Name: "Home", Width: 800, Height: 500, HTML: `<button id="cta">Before</button><p id="other">Other</p>`}}})
	if err != nil {
		t.Fatal(err)
	}
	base := cv.Version
	_, err = h.cl.EditCanvas(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpHTMLPatch, Target: "cta", HTML: `<button id="cta">After</button>`, Session: "agent-1"})
	if err != nil {
		t.Fatal(err)
	}
	h.settle("review canvas", func() bool {
		c := h.app.canvases.Canvas()
		return h.app.canvases.Selected() == cv.ID && c != nil && c.Version > base
	})
	h.app.canvases.ShowVersions(h.app.canvasOptions(), base)
	h.settle("comparison", func() bool { return h.tt.HasText("Version 1 → 2") })
	if err := h.tt.Click("Canvas tools"); err != nil {
		t.Fatal(err)
	}
	h.tt.Frame()
	h.tt.Advance(time.Second)
	h.tt.Frame()
	if err := h.tt.Click("Review"); err != nil {
		t.Fatal(err)
	}
	h.tt.Frame()
	if err := h.tt.Click("reject #cta"); err != nil {
		t.Fatalf("%v; %q", err, h.tt.Texts())
	}
	h.settle("rejected element", func() bool { c, _ := h.cl.Canvas(cv.ID); return strings.Contains(c.Frames[0].HTML, ">Before</button>") })
}
