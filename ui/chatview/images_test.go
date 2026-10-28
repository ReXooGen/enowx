package chatview

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
)

func TestOMPWorkShowsThinkingToolsAndVision(t *testing.T) {
	var b bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 160, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 160; x++ {
			img.SetRGBA(x, y, color.RGBA{R: 40, G: 80, B: 140, A: 255})
		}
	}
	if err := png.Encode(&b, img); err != nil {
		t.Fatal(err)
	}
	entries := []proto.TranscriptEntry{
		{Index: 0, Kind: proto.TranscriptUser, Text: "Review this design", Images: []proto.TranscriptImage{{Name: "Image", Thumbnail: b.Bytes()}}},
		{Index: 1, Kind: proto.TranscriptThinking, Text: "Inspecting spacing and alignment."},
		{Index: 2, Kind: proto.TranscriptTool, Text: "Bash", Detail: "enowx-app canvas show checkout", Result: "Design 1 · Checkout", Done: true},
		{Index: 3, Kind: proto.TranscriptTool, Text: "canvas_screenshot", Detail: "Design 1", Images: []proto.TranscriptImage{{Name: "Image", Thumbnail: b.Bytes()}}, Done: true},
		{Index: 4, Kind: proto.TranscriptAssistant, Text: "The design is ready to review."},
	}
	turns := splitTurns(entries)
	if len(turns) != 1 || len(turns[0].work) != 3 || len(turns[0].answer) != 1 {
		t.Fatalf("work was mixed into the answer: %+v", turns)
	}
	ch := New()
	for i := 0; i < 4; i++ {
		ch.open[i] = true
	}
	tt := ui.NewTester(func(c *ui.Context) {
		ui.Column(c).Width(420).Height(900).Padding(14).Gap(10).Background(theme.Background).TextColor(theme.Text).Children(func() { ch.turn(c, turns[0], true, false) })
	}, 420, 900)
	tt.Frame()
	for _, text := range []string{"Thinking", "Inspecting spacing and alignment.", "Design 1 · Checkout", "Image", "The design is ready to review."} {
		if !tt.HasText(text) {
			t.Fatalf("missing %q: %q", text, tt.Texts())
		}
	}
	if dir := os.Getenv("CANVAS_QA_DIR"); dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		file, err := os.Create(filepath.Join(dir, "omp-work.png"))
		if err != nil {
			t.Fatal(err)
		}
		err = png.Encode(file, tt.Image())
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
}
