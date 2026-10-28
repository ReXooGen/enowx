package ui

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

// The transcript has no terminal menus, live reasoning or connection errors.
// Keep the same running session accessible when a reply has not arrived yet.
func (a *App) canvasAgentViewToggle(c *ui.Context, hc *hostCanvasChat, s proto.SessionInfo) {
	ui.Row(c.Key("canvas-agent-view")).FillWidth().Height(38).Shrink(0).PaddingX(12).Gap(10).AlignItems(ui.Center).
		BorderWidth(0, 0, 1, 0).BorderColor(theme.Border).Children(func() {
		selected := 0
		if hc.terminal {
			selected = 1
		}
		if i := widget.Segmented(c.Key("mode"), []string{"Chat", "Terminal"}, selected, nil); i >= 0 {
			hc.terminal = i == 1
		}
		ui.Spacer(c)
		st := statusOf(s)
		ui.Text(c, st.label).FontSize(11.5).TextColor(st.color).SingleLine()
	})
}
