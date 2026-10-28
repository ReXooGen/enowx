package canvasview

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

// Agent activity floats below the design menu, leaving the board's layout still.
func (v *View) agentPill(c *ui.Context, o Options, cv proto.Canvas) {
	now := time.Now()
	active := v.designing(now)
	if !active && !cv.Paused {
		return
	}
	if active {
		c.After(designingFor - now.Sub(cv.AgentAt) + 50*time.Millisecond)
	}
	who := "An agent"
	if o.AgentName != nil && cv.Agent != "" {
		who = o.AgentName(cv.Agent)
	}
	label := who + " is designing"
	if cv.Paused {
		label = "Agents paused"
	}
	pill := ui.Row(c.Key("agent-pill")).Absolute().Top(12).Right(12).Height(34).PaddingX(10).Gap(7).AlignItems(ui.Center).
		Radius(17).Background(theme.Panel).Border(1, theme.AccentLine).Shadow(0, 3, 12, 0, theme.Fade(theme.TextStrong, .12)).Label("Canvas agent status")
	pill.Children(func() {
		if cv.Paused {
			ui.Icon(c, icon.Pause).FontSize(12).TextColor(theme.Warning)
		} else {
			widget.Spinner(c, 12)
		}
		ui.Text(c, label).FontSize(12).TextColor(theme.Text).SingleLine().MaxWidth(180).Shrink(1)
		if cv.Paused {
			if widget.IconButton(c, icon.Play, "Resume", false).Size(24, 24).Clicked() && o.Control != nil {
				o.Control(cv.ID, proto.CanvasResume, v.reportErr)
			}
		} else {
			if widget.IconButton(c, icon.Pause, "Pause", false).Size(24, 24).Clicked() && o.Control != nil {
				o.Control(cv.ID, proto.CanvasPause, v.reportErr)
			}
			if widget.IconButton(c, icon.Stop, "Stop", false).Size(24, 24).Clicked() && o.Control != nil {
				o.Control(cv.ID, proto.CanvasStop, v.reportErr)
			}
		}
	})
}
