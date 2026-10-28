package ui

import (
	"strings"
	"testing"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/proto"
)

func TestVariationPresetSendsBriefWithoutExpandingTheChatBubble(t *testing.T) {
	cv := proto.Canvas{CanvasInfo: proto.CanvasInfo{ID: "canvas", Title: "Client flow"}, Frames: []proto.CanvasFrame{{ID: "frame", Name: "Source", Width: 800, Height: 600}}}
	context := canvas.ChatContext(cv, "frame", "client")
	message := canvasGuidedMessage(context, agent.CanvasVariationsRequest)
	if !strings.Contains(message, agent.CanvasVariationsGuide) || !strings.Contains(message, "selected #client") {
		t.Fatal("variation request lost its brief or source selection")
	}
	if got := canvas.WithoutContext(message); got != agent.CanvasVariationsRequest {
		t.Fatalf("app guidance leaked into the visible ask: %q", got)
	}
	if got := canvasGuidedMessage(context, "Change this label"); got != context+"Change this label" {
		t.Fatal("unrelated request was rewritten")
	}
}
