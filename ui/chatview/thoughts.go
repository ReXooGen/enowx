package chatview

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
)

// Recorded thoughts remain visible independently of the tool history. A short
// preview avoids an empty turn without flooding it with long reasoning blocks.
func (ch *Chat) thoughts(c *ui.Context, entries []proto.TranscriptEntry, thinking, working bool) {
	key := entries[0].Index
	if ch.thoughtOpen == nil {
		ch.thoughtOpen = map[int]bool{}
	}
	open, set := ch.thoughtOpen[key]
	if !set {
		open = working
	}
	title := "Thoughts"
	if thinking {
		title = "Thinking"
	}
	var duration int64
	for _, entry := range entries {
		duration += entry.DurationMs
	}
	ui.Column(c).FillWidth().Gap(6).Padding(9, 10).Radius(8).Background(theme.Sunken).Children(func() {
		head := ui.ButtonBase(c).FillWidth().Gap(7).AlignItems(ui.Center).Cursor(ui.CursorPointer).Label("Thinking details")
		head.Children(func() {
			ui.Icon(c, icon.Brain).FontSize(13).TextColor(theme.Accent)
			ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(theme.TextSoft)
			if working && !thinking {
				ui.Text(c, "Running tools").FontSize(11).TextColor(theme.Muted)
			}
			if len(entries) > 1 {
				ui.Text(c, fmt.Sprintf("%d entries", len(entries))).FontSize(11).TextColor(theme.Muted)
			}
			if duration > 0 {
				ui.Text(c, (time.Duration(duration) * time.Millisecond).Round(time.Second).String()).FontSize(11).TextColor(theme.Muted)
			}
			ui.Spacer(c)
			chevron := icon.ChevronRight
			if open {
				chevron = icon.ChevronDown
			}
			ui.Icon(c, chevron).FontSize(11).TextColor(theme.Muted)
		})
		if head.Clicked() {
			ch.thoughtOpen[key] = !open
		}
		if open {
			ui.Scroll(c).FillWidth().MaxHeight(200).Children(func() {
				ui.Column(c).FillWidth().Gap(8).Children(func() {
					for i := len(entries) - 1; i >= 0; i-- {
						entry := entries[i]
						ui.Text(c.Key(entry.Index), entry.Text).FontSize(12).TextColor(theme.Muted).Selectable()
					}
				})
			})
		} else {
			preview := []rune(strings.TrimSpace(entries[len(entries)-1].Text))
			if len(preview) > 260 {
				preview = append(preview[:260], '…')
			}
			ui.Text(c, string(preview)).FontSize(12).TextColor(theme.Muted).Selectable()
		}
	})
}
