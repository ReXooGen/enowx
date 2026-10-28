package ui

import (
	"strings"

	"github.com/enowx/enowx/core/agent"
)

// Carry the preset's brief in the app's context, including for a conversation
// started by an older build. The transcript still shows the user's short ask.
func canvasGuidedMessage(context, text string) string {
	if strings.TrimSpace(text) == agent.CanvasVariationsRequest {
		context = strings.TrimSuffix(context, "]\n") + "\n" + agent.CanvasVariationsGuide + "]\n"
	}
	return context + text
}
