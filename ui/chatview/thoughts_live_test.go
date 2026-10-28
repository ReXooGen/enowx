package chatview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"testing"
)

func TestRecordedThinkingUpdatesWhileToolsRun(t *testing.T) {
	ch := New()
	entries := []proto.TranscriptEntry{{Index: 0, Kind: proto.TranscriptThinking, Text: "Initial recorded thought"}, {Index: 1, Kind: proto.TranscriptTool, Text: "Bash", Detail: "check design", Done: false}}
	ch.open[0] = false
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(420).Height(500).Children(func() { ch.turn(c, splitTurns(entries)[0], true, false) })
	}, 420, 500)
	tt.Frame()
	if !tt.HasText("Running tools") || !tt.HasText("Initial recorded thought") {
		t.Fatal("tool execution hides recorded thought")
	}
	if err := tt.Click("Thinking details"); err != nil {
		t.Fatal(err)
	}
	if ch.open[0] {
		t.Fatal("thinking toggle changed tool history")
	}
	entries = append(entries, proto.TranscriptEntry{Index: 2, Kind: proto.TranscriptThinking, Text: "New recorded thought"})
	tt.Frame()
	if !tt.HasText("New recorded thought") {
		t.Fatal("new thought did not replace the preview while tools run")
	}
}
