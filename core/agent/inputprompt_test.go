package agent

import "testing"

func TestInputPromptRecognizesInteractiveFooters(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		want       bool
	}{
		{"choice dialog", "Choose a theme\nDark\nLight\nspace select · n note · ↑/↓ move · ←/→ question · esc cancel · Alt+H hide", true},
		{"selection", "Choose a project\n↑/↓ move • enter select • esc cancel", true},
		{"confirmation", "Save changes?\nEnter to confirm | Esc to cancel", true},
		{"question in output", "Would you like to choose a theme?", false},
		{"ordinary output", "Select the new theme, move the config, then cancel the old job.", false},
		{"idle shell", "user@host:~$", false},
		{"old dialog", "↑/↓ move · enter select · esc cancel\nRunning tests\nTest one passed\nTest two passed\nAll done\nReady", false},
		{"not actionable", "esc cancel", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := InputPrompt(tc.text, 24); got != tc.want {
				t.Fatalf("InputPrompt = %v, want %v", got, tc.want)
			}
		})
	}
}
