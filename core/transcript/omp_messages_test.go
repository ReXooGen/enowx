package transcript

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
	"testing"

	"github.com/enowx/enowx/proto"
)

func TestOMPToolsThinkingAndVisionStayInOneTurn(t *testing.T) {
	r := NewOMPReader("")
	p := &ompParser{}
	var pic bytes.Buffer
	if err := png.Encode(&pic, image.NewRGBA(image.Rect(0, 0, 300, 200))); err != nil {
		t.Fatal(err)
	}
	p.parse([]byte(fmt.Sprintf(`{"type":"message","message":{"role":"user","content":[{"type":"text","text":"Review this design"},{"type":"image","data":"%s"}]}}`, base64.StdEncoding.EncodeToString(pic.Bytes()))), r)
	p.parse([]byte(`{"type":"message","timestamp":"2026-10-28T01:00:00Z","message":{"role":"assistant","content":[{"type":"thinking","thinking":"Inspect the layout"},{"type":"toolCall","id":"call-1","name":"bash","arguments":{"command":"enowx-app canvas show test"}}]}}`), r)
	entries, _ := r.Since(0)
	if len(entries) != 3 || entries[1].Kind != proto.TranscriptThinking || entries[2].Kind != proto.TranscriptTool || entries[2].Done {
		t.Fatalf("pending work: %+v", entries)
	}
	if len(entries[0].Images) != 1 || len(entries[0].Images[0].Thumbnail) == 0 {
		t.Fatal("user vision image was dropped")
	}
	rev := r.rev
	p.parse([]byte(`{"type":"message","timestamp":"2026-10-28T01:00:02Z","message":{"role":"toolResult","toolCallId":"call-1","toolName":"bash","isError":true,"content":[{"type":"text","text":"Permission denied"}]}}`), r)
	changes, _ := r.Since(rev)
	if len(changes) != 1 || changes[0].Index != 2 || !changes[0].Done || !changes[0].Failed || changes[0].Text != "Bash" || changes[0].Detail != "enowx-app canvas show test" || changes[0].Result != "Permission denied" || changes[0].DurationMs != 2000 {
		t.Fatalf("tool result not linked: %+v", changes)
	}
	p.parse([]byte(`{"type":"message","message":{"role":"assistant","content":[{"type":"text","text":"Please allow access."}]}}`), r)
	entries, _ = r.Since(0)
	if len(entries) != 4 || entries[3].Kind != proto.TranscriptAssistant {
		t.Fatalf("tool output leaked as a reply: %+v", entries)
	}
	thumb := entries[0].Images[0].Thumbnail
	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb))
	if err != nil || cfg.Width > 160 || cfg.Height > 160 || len(thumb) > 8<<10 {
		t.Fatalf("unbounded vision preview: %+v %v", cfg, err)
	}
}

func TestOMPOrphanToolResultsAndErrorsRemainVisible(t *testing.T) {
	r := NewOMPReader("")
	p := &ompParser{}
	p.parse([]byte(`{"type":"message","message":{"role":"toolResult","toolCallId":"missing","toolName":"read","content":"Contents"}}`), r)
	p.parse([]byte(`{"type":"message","message":{"role":"assistant","content":[],"errorMessage":"Model request failed"}}`), r)
	e, _ := r.Since(0)
	if len(e) != 2 || e[0].Kind != proto.TranscriptTool || !e[0].Done || e[0].Result != "Contents" || e[1].Kind != proto.TranscriptNotice || !e[1].Failed {
		t.Fatalf("events: %+v", e)
	}
	if got := transcriptThumbnail("https://example.com/private.png"); len(got.Thumbnail) > 0 {
		t.Fatal("unsupported image source should remain a placeholder")
	}
}
