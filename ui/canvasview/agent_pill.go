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
	waiting := false
	agentID := cv.Agent
	fromSession := false
	if o.AgentActivity != nil {
		if session, ok := o.AgentActivity(cv.ID); ok {
			fromSession = true
			agentID = session.ID
			waiting = session.State == proto.StateWaiting
			active = session.State == proto.StateWorking || waiting
		}
	}
	if !active && !cv.Paused {
		return
	}
	if active && !fromSession {
		c.After(designingFor - now.Sub(cv.AgentAt) + 50*time.Millisecond)
	}
	who := "An agent"
	if o.AgentName != nil && agentID != "" {
		who = o.AgentName(agentID)
	}
	label := who + " is designing"
	if fromSession {
		label = who + " is working"
	}
	if waiting {
		label = who + " needs input"
	}
	if cv.Paused {
		label = "Agents paused"
	}
	pill := ui.Row(c.Key("agent-pill")).Absolute().Top(12).Right(12).Height(34).PaddingX(10).Gap(7).AlignItems(ui.Center).
		Radius(17).Background(theme.Panel).Border(1, theme.AccentLine).Shadow(0, 3, 12, 0, theme.Fade(theme.TextStrong, .12)).Label("Canvas agent status")
	pill.Children(func() {
		if cv.Paused {
			ui.Icon(c, icon.Pause).FontSize(12).TextColor(theme.Warning)
		} else if waiting {
			ui.Icon(c, icon.Alert).FontSize(12).TextColor(theme.Warning)
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
