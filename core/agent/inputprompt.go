package agent

import "strings"

// InputPrompt recognizes a keyboard-driven choice dialog at the bottom of a
// terminal. Ordinary questions in prose and an idle command prompt do not count.
// This is a fallback for programs that do not report their input state via hooks.
func InputPrompt(text string, rows int) bool {
	// Only the current screen's tail can describe the active interaction.
	if len(text) > 8192 {
		text = text[len(text)-8192:]
	}
	lines := strings.Split(strings.TrimRight(text, "\n\r \t"), "\n")
	if rows > 0 && len(lines) > rows {
		lines = lines[len(lines)-rows:]
	}
	if len(lines) > 3 {
		lines = lines[len(lines)-3:]
	}
	for _, line := range lines {
		line = strings.ToLower(line)
		if !strings.Contains(line, "cancel") {
			continue
		}
		separator := strings.Contains(line, "·") || strings.Contains(line, "•") || strings.Contains(line, " | ")
		if !separator {
			continue
		}
		selecting := strings.Contains(line, "select") && (strings.Contains(line, "move") || strings.Contains(line, "↑") || strings.Contains(line, "↓"))
		confirming := strings.Contains(line, "enter") && (strings.Contains(line, "confirm") || strings.Contains(line, "submit")) && (strings.Contains(line, "esc") || strings.Contains(line, "escape"))
		if selecting || confirming {
			return true
		}
	}
	return false
}
