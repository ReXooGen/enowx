package daemon

import (
	"strings"
	"testing"

	"github.com/enowx/enowx/proto"
)

func TestCanvasWorkflowThroughDaemon(t *testing.T) {
	srv := startServer(t)
	a, b := connect(t, srv, nil), connect(t, srv, nil)
	cv, err := a.CreateCanvas(proto.CanvasCreateParams{Title: "Workflow", Frames: []proto.CanvasFrame{{Name: "Home", HTML: `<button id="go">Continue</button>`}, {Name: "Done", HTML: `<h1>Done</h1>`}}})
	if err != nil {
		t.Fatal(err)
	}
	lib, err := a.CanvasLibrary(proto.CanvasLibraryParams{Op: "save", Canvas: cv.ID, Frame: cv.Frames[0].ID, Target: "go", Component: proto.CanvasLibraryComponent{Name: "Continue button"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = b.CanvasLibrary(proto.CanvasLibraryParams{Op: "insert", Canvas: cv.ID, Frame: cv.Frames[1].ID, Component: lib.Components[0]}); err != nil {
		t.Fatal(err)
	}
	got, err := a.Canvas(cv.ID)
	if err != nil || !strings.Contains(got.Frames[1].HTML, "data-enowx-component") {
		t.Fatalf("linked instance missing across clients: %v", err)
	}
	if _, err = b.EditCanvas(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpTransition, Transition: &proto.CanvasTransition{Element: "go", Target: cv.Frames[1].ID, Trigger: "click", Effect: "fade", Easing: "linear", Duration: 100}}); err != nil {
		t.Fatal(err)
	}
	out, err := a.RenderCanvas(proto.CanvasRenderParams{ID: cv.ID, Format: "offline-html"})
	if err != nil || !strings.Contains(string(out.Data), "__enowxNavigate") || !strings.Contains(string(out.Data), "Content-Security-Policy") {
		t.Fatalf("offline prototype export: %v", err)
	}
	if _, err = a.ControlCanvas(cv.ID, proto.CanvasPause); err != nil {
		t.Fatal(err)
	}
	if _, err = b.CanvasLibrary(proto.CanvasLibraryParams{Op: "remove", Session: "agent", Component: lib.Components[0]}); err == nil {
		t.Fatal("agent changed library while canvas was paused")
	}
	if _, err = a.CanvasLibrary(proto.CanvasLibraryParams{Op: "remove", Component: lib.Components[0]}); err != nil {
		t.Fatalf("user could not edit paused canvas library: %v", err)
	}
}
