package canvasview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/enowx/enowx/proto"
)

type connectorTestPage struct {
	livePage
	loaded    string
	evaluated chan string
}

func (p *connectorTestPage) LoadHTML(doc string)           { p.loaded = doc }
func (p *connectorTestPage) Eval(code string) (any, error) { p.evaluated <- code; return true, nil }

func TestLiveMockupKeepsConnectorsAboveItsBrowserSurface(t *testing.T) {
	f := proto.CanvasFrame{ID: "home", Width: 600, Height: 400, HTML: `<div style="position:absolute;left:40px;top:40px;width:120px;height:60px;background:#edf2ff">Client</div><div style="position:absolute;left:400px;top:260px;width:120px;height:60px;background:#edf2ff;border-radius:50%">Provider</div>`, Elements: []proto.CanvasElement{{ID: "a", Type: proto.ShapeRect, X: 40, Y: 40, W: 120, H: 60}, {ID: "b", Type: proto.ShapeEllipse, X: 400, Y: 260, W: 120, H: 60}, {ID: "connector", Type: proto.ShapeArrow, From: "a", To: "b", Routing: "elbow", Stroke: "#245bc1", StrokeWidth: 3}}}
	cv := proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "test"}, Frames: []proto.CanvasFrame{f}}
	p := &connectorTestPage{evaluated: make(chan string, 1)}
	w := &Web{page: p}
	w.show(cv, f, nil)
	if !strings.Contains(p.loaded, "data-enowx-connectors") || !strings.Contains(p.loaded, `points="160,70 280,70 280,290 400,290 `) {
		t.Fatal("live document omitted the connector overlay")
	}
	if dir := os.Getenv("CANVAS_QA_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "live-connectors.html"), []byte(p.loaded), 0600); err != nil {
			t.Fatal(err)
		}
	}
	before := htmlKey(f)
	f.Elements[1].X += 40
	if htmlKey(f) == before {
		t.Fatal("shape edits did not invalidate the thumbnail")
	}
	oldDoc := p.loaded
	w.show(cv, f, nil)
	select {
	case code := <-p.evaluated:
		if !strings.Contains(code, "innerHTML") || !strings.Contains(code, "440,290") {
			t.Fatalf("endpoint update missing: %s", code)
		}
	case <-time.After(time.Second):
		t.Fatal("moving a shape did not update its live connector")
	}
	if p.loaded != oldDoc {
		t.Fatal("shape drag unnecessarily reloaded the mockup")
	}
}
