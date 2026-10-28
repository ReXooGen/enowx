package canvas

import (
	"fmt"
	"golang.org/x/net/html"
	"strings"
	"testing"

	"github.com/enowx/enowx/proto"
)

func workflowCanvas(t *testing.T, s *Store) proto.Canvas {
	t.Helper()
	cv, err := s.Create(proto.CanvasCreateParams{Title: "Workflow", Frames: []proto.CanvasFrame{{Name: "Desktop", Width: 400, Height: 300, HTML: `<section id="parent"><button id="first">One</button><button id="second">Two</button></section>`, Elements: []proto.CanvasElement{{ID: "a", Type: proto.ShapeRect, X: 10, Y: 20, W: 40, H: 30}, {ID: "b", Type: proto.ShapeRect, X: 100, Y: 80, W: 30, H: 40}, {ID: "c", Type: proto.ShapeRect, X: 200, Y: 50, W: 40, H: 30}}}}})
	if err != nil {
		t.Fatal(err)
	}
	return cv
}

func TestCanvasWorkflowBatchIsUndoableAndRespectsLocks(t *testing.T) {
	s := Open(t.TempDir())
	cv := workflowCanvas(t, s)
	edit := func(action string) (proto.Canvas, error) {
		v, _, _, err := s.Edit(proto.CanvasEdit{ID: cv.ID, Frame: cv.Frames[0].ID, Op: proto.CanvasOpBatch, Targets: []string{"a", "b", "c"}, Action: action, DX: 12, DY: 8})
		return v, err
	}
	grouped, err := edit("group")
	if err != nil {
		t.Fatal(err)
	}
	if grouped.Frames[0].Elements[0].Group == "" || grouped.Frames[0].Elements[0].Group != grouped.Frames[0].Elements[1].Group {
		t.Fatal("group was not persisted")
	}
	moved, err := edit("move")
	if err != nil {
		t.Fatal(err)
	}
	if moved.Frames[0].Elements[0].X != 22 || moved.Frames[0].Elements[1].Y != 88 {
		t.Fatal("batch movement did not move all shapes")
	}
	undo, err := s.Undo(cv.ID)
	if err != nil {
		t.Fatal(err)
	}
	if undo.Frames[0].Elements[0].X != 10 || undo.Frames[0].Elements[1].Y != 80 {
		t.Fatal("one undo did not reverse the whole batch")
	}
	if _, err = edit("lock"); err != nil {
		t.Fatal(err)
	}
	if _, err = edit("left"); err == nil {
		t.Fatal("locked shapes were aligned")
	}
	if _, err = edit("unlock"); err != nil {
		t.Fatal(err)
	}
	distributed, err := edit("distribute-x")
	if err != nil {
		t.Fatal(err)
	}
	x := distributed.Frames[0].Elements
	if x[1].X-x[0].X-x[0].W != x[2].X-x[1].X-x[1].W {
		t.Fatal("distribution gaps differ")
	}
}

func TestCanvasWorkflowElementReviewPreservesOtherEdits(t *testing.T) {
	s := Open(t.TempDir())
	cv := workflowCanvas(t, s)
	base := cv.Version
	head, _, _, err := s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpHTMLPatch, Target: "first", HTML: `<button id="first">Agent edit</button>`})
	if err != nil {
		t.Fatal(err)
	}
	_, _, _, err = s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpHTMLPatch, Target: "second", HTML: `<button id="second">User edit</button>`})
	if err != nil {
		t.Fatal(err)
	}
	got, _, _, err := s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpReview, Review: &proto.CanvasElementReview{Frame: cv.Frames[0].ID, Element: "first", Base: base, Head: head.Version, Decision: "reject"}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got.Frames[0].HTML, ">One</button>") || !strings.Contains(got.Frames[0].HTML, ">User edit</button>") {
		t.Fatal("review overwrote an unrelated user edit")
	}
	if len(got.Reviews) != 1 {
		t.Fatal("review decision was not recorded")
	}
	_, _, _, err = s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpReview, Review: &proto.CanvasElementReview{Frame: cv.Frames[0].ID, Element: "first", Base: base, Head: head.Version, Decision: "keep"}})
	if err == nil {
		t.Fatal("stale review overwrote a newer change")
	}
}

func TestCanvasWorkflowLibraryLinksAcrossCanvases(t *testing.T) {
	s := Open(t.TempDir())
	a := workflowCanvas(t, s)
	b := workflowCanvas(t, s)
	lib, err := s.Library(proto.CanvasLibraryParams{Op: "save", Component: proto.CanvasLibraryComponent{Name: "Contact form", HTML: `<form id="form"><label for="email">Email</label><input id="email"></form>`}})
	if err != nil {
		t.Fatal(err)
	}
	comp := lib.Components[0]
	for _, cv := range []proto.Canvas{a, b} {
		if _, err := s.Library(proto.CanvasLibraryParams{Op: "insert", Canvas: cv.ID, Frame: cv.Frames[0].ID, Component: comp}); err != nil {
			t.Fatal(err)
		}
	}
	comp.HTML = `<form id="form"><label for="email">Updated label</label><input id="email"></form>`
	updated, err := s.Library(proto.CanvasLibraryParams{Op: "save", Component: comp})
	if err != nil {
		t.Fatal(err)
	}
	if len(updated.Canvases) != 2 {
		t.Fatalf("updated %d canvases", len(updated.Canvases))
	}
	for _, cv := range updated.Canvases {
		page := cv.Frames[0].HTML
		if !strings.Contains(page, "Updated label") {
			t.Fatal("linked instance did not update")
		}
		d, _ := parseDoc(page)
		input := find(d.root, func(n *html.Node) bool { return n.Data == "input" })
		label := find(d.root, func(n *html.Node) bool { return n.Data == "label" })
		if attr(input, "id") != attr(label, "for") {
			t.Fatal("instance broke label references")
		}
	}
}

func TestCanvasWorkflowOfflineIncludesCSSFontsImagesAndScripts(t *testing.T) {
	resources := map[string]string{"https://assets.example/base.css": `@import "theme.css"; .title{background:url(image.png)} @font-face{font-family:demo;src:url(font.woff2)}`, "https://assets.example/theme.css": `.title{color:#123456}`, "https://assets.example/image.png": "image-data", "https://assets.example/font.woff2": "font-data", "https://assets.example/demo.js": "window.demoLoaded=true;"}
	fetch := func(u string) ([]byte, string, error) {
		if data, ok := resources[u]; ok {
			return []byte(data), "application/octet-stream", nil
		}
		return nil, "", fmt.Errorf("missing resource")
	}
	page := `<!doctype html><html><head><link rel="stylesheet" href="https://assets.example/base.css"></head><body><h1 class="title">Demo</h1><script src="https://assets.example/demo.js"></script></body></html>`
	got, err := OfflineDocument(page, fetch)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, "https://assets.example") || !strings.Contains(got, "base64,") || !strings.Contains(got, "demoLoaded") {
		t.Fatal("export still depends on external assets")
	}
	shared := Standalone(proto.Canvas{Frames: []proto.CanvasFrame{{HTML: page, Width: 400, Height: 300}}}, nil)
	if _, err := OfflineDocument(shared, fetch); err != nil {
		t.Fatalf("nested shared preview: %v", err)
	}
}

func TestCanvasWorkflowTransitionsPersistAndExport(t *testing.T) {
	s := Open(t.TempDir())
	cv := workflowCanvas(t, s)
	_, _, _, err := s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpFrameAdd, FrameInfo: &proto.CanvasFrame{Name: "Next"}})
	if err != nil {
		t.Fatal(err)
	}
	got, _, _, err := s.Edit(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpTransition, Transition: &proto.CanvasTransition{Element: "first", Target: "Next", Trigger: "click", Effect: "slide", Duration: 220, Easing: "ease-out"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Frames[0].Transitions) != 1 || !strings.Contains(FrameDocument(got.Frames[0], nil), "translateX(-18px)") {
		t.Fatal("transition did not reach the preview document")
	}
}
