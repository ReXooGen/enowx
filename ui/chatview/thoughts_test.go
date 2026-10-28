package chatview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestThoughtPreviewIsVisibleWithToolsCollapsed(t *testing.T) {
	ch := New()
	entries := []proto.TranscriptEntry{{Index: 0, Kind: proto.TranscriptUser, Text: "Design it"}, {Index: 1, Kind: proto.TranscriptThinking, Text: "Review the layout before making changes."}, {Index: 2, Kind: proto.TranscriptTool, Text: "Bash", Result: "Raw shell output", Done: true}}
	turns := splitTurns(entries)
	ch.open[0] = false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(400).Height(500).Children(func() { ch.turn(c, turns[0], false, false) })
	}, 400, 500)
	tt.Frame()
	if !tt.HasText("Thoughts") || !tt.HasText("Review the layout before making changes.") {
		t.Fatal("thought hidden inside tool history")
	}
	if tt.HasText("Raw shell output") {
		t.Fatal("collapsed tool output leaked into thought preview")
	}
}
