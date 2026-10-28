package canvasview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

func (v *View) connectorRouting(c *ui.Context, o Options, el proto.CanvasElement) {
	labels := []string{"Straight", "Elbow 90°", "Elbow horizontal", "Elbow vertical"}
	values := []string{"straight", "elbow", "elbow-x", "elbow-y"}
	selected := labels[0]
	for i, value := range values {
		if el.Routing == value {
			selected = labels[i]
		}
	}
	ui.Text(c, "Routing").FontSize(11).TextColor(theme.Muted)
	if widget.Choice(c.Key("routing"), "Connector routing", &selected, labels) {
		for i, label := range labels {
			if selected == label {
				el.Routing = values[i]
				v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpElementSet, Element: &el})
			}
		}
	}
}
