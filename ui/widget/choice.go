package widget

import (
	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
)

// Choice is a labeled drop-down of strings; it reports a changed selection.
func Choice(c *ui.Context, label string, value *string, options []string) bool {
	before := *value
	sel := ui.SelectBase(c, value)
	Trigger(sel.Trigger, sel.Open()).Label(label).Children(func() {
		text := *value
		if text == "" {
			text = "Choose…"
		}
		ui.Text(c, text).FontSize(12).TextColor(theme.Text).SingleLine().Shrink(1)
		ui.Spacer(c)
		ui.Icon(c, icon.ChevronDown).FontSize(12).TextColor(theme.Muted)
	})
	sel.Popup(func(panel ui.Element) {
		Popup(panel.MinWidth(180).MaxHeight(240))
		ui.Scroll(c).FillWidth().Children(func() {
			for _, option := range options {
				item := sel.Item(option)
				PopupItem(item, item.Highlighted()).Label(label + ": " + option).Children(func() { ui.Text(c, option).FontSize(12).SingleLine().Shrink(1) })
			}
		})
	})
	return before != *value
}
