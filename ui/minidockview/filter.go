package minidockview

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

func (v *View) hostFilter(c *ui.Context, hosts []Host) {
	before := v.host
	name := "All hosts"
	for _, h := range hosts {
		if h.ID == v.host {
			name = h.Name
		}
	}
	sel := ui.SelectBase(c, &v.host)
	widget.Trigger(sel.Trigger, sel.Open()).Height(28).Label("Filter mini dock host").Children(func() {
		ui.Text(c, name).FontSize(11).SingleLine().Shrink(1)
		ui.Spacer(c)
		ui.Icon(c, icon.ChevronDown).FontSize(11).TextColor(theme.Muted)
	})
	sel.Popup(func(panel ui.Element) {
		widget.Popup(panel.MinWidth(220).MaxHeight(240))
		ui.Scroll(c).FillWidth().Children(func() {
			for _, h := range append([]Host{{Name: "All hosts"}}, hosts...) {
				item := sel.Item(h.ID)
				widget.PopupItem(item, item.Highlighted()).Label("Filter " + h.Name).Children(func() {
					ui.Text(c, h.Name).FontSize(12).SingleLine().Shrink(1)
				})
			}
		})
	})
	if before != v.host {
		v.list = ui.ListState{}
		v.holding = false
	}
}
