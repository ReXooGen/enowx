package minidockview

import (
	"fmt"
	"slices"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

// Draw shows the overview. Window resizing stays with the caller, so a collapse
// finishes before the native window gives the desktop its space back.
func (v *View) Draw(c *ui.Context, o Options) {
	c.Root().Background(ui.Transparent)
	if v.host != "" && !slices.ContainsFunc(o.Hosts, func(h Host) bool { return h.ID == v.host }) {
		v.host = ""
	}
	if c.Shortcut(0, ui.KeyEscape) {
		v.expanded = false
	}
	root := ui.Column(c.Key("mini-dock")).Fill().Justify(ui.End).Padding(6)
	if v.fresh {
		root.Jump("appear", 0)
		root.Jump("expand", 0)
		v.fresh = false
	}
	appear, fraction := float32(1), float32(0)
	if v.expanded {
		fraction = 1
	}
	if !o.ReduceMotion {
		appear = root.AnimateTo("appear", 1, ui.Animation{Duration: 220 * time.Millisecond, Ease: widget.Out})
		fraction = root.AnimateTo("expand", fraction, ui.Animation{Duration: 200 * time.Millisecond, Ease: widget.Out})
	}
	if o.Resize != nil {
		if v.expanded {
			o.Resize(ExpandedHeight)
		} else if fraction == 0 {
			o.Resize(CompactHeight)
		}
	}
	attention, working := 0, 0
	nowAttention := map[string]bool{}
	newAttention := false
	for _, s := range o.Sessions {
		if s.Status == NeedsInput || s.Status == Failed {
			attention++
			key := s.HostID + ":" + s.ID
			nowAttention[key] = true
			if v.lastAttention != nil && !v.lastAttention[key] {
				newAttention = true
			}
		}
		if s.Status == Working {
			working++
		}
	}
	v.lastAttention = nowAttention
	if newAttention {
		root.Jump("attention", 0)
	}
	notice := float32(1)
	if !o.ReduceMotion {
		notice = root.AnimateTo("attention", 1, ui.Animation{Duration: 900 * time.Millisecond, Ease: widget.Out})
	}
	root.Children(func() {
		ui.Column(c.Key("card")).FillWidth().Shrink(0).Height(float32(CompactHeight-12)+fraction*float32(ExpandedHeight-CompactHeight)).
			Radius(14).Background(theme.Dialog).TextColor(theme.Text).Clip().
			Opacity(appear).TransformOrigin(.5, 1).Scale(.97+.03*appear, .97+.03*appear).
			Shadow(0, 3, 6, 0, ui.RGBA(0, 0, 0, 50)).Children(func() {
			v.header(c, o)
			v.summary(c, working, attention, 1-notice)
			if fraction > 0 {
				v.details(c, o)
			}
		})
	})
	c.After(time.Second) // redraw cached host snapshots; no network work here
}

func (v *View) header(c *ui.Context, o Options) {
	ui.Row(c).FillWidth().Height(48).Shrink(0).Padding(8, 10, 0, 14).Gap(8).Children(func() {
		ui.Box(c).Size(27, 27).Radius(7).Background(theme.Hover).Center().Children(func() { ui.Icon(c, icon.Logo).FontSize(18).TextColor(theme.Accent) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(2).DragWindow().Children(func() {
			ui.Text(c, "enowX").FontSize(12.5).FontWeight(600)
			ui.Text(c, fmt.Sprintf("%d sessions · %d hosts", len(o.Sessions), len(o.Hosts))).FontSize(10.5).TextColor(theme.Muted).SingleLine()
		})
		ui.Box(c).Size(22, 28).Center().DragWindow().Tooltip("Drag the mini dock").Children(func() { ui.Icon(c, icon.More).FontSize(14).TextColor(theme.Muted) })
		if widget.IconButton(c.Key("restore"), icon.External, "Open enowX", false).Clicked() && o.Restore != nil {
			o.Restore()
		}
	})
}

func (v *View) summary(c *ui.Context, working, attention int, notice float32) {
	b := ui.ButtonBase(c.Key("summary")).FillWidth().Height(40).Shrink(0).PaddingX(14).Gap(12).Label("Show sessions").Cursor(ui.CursorPointer)
	if v.expanded {
		b.Label("Collapse sessions")
	}
	if b.Hovered() {
		b.Background(theme.Hover)
	}
	b.Children(func() {
		ui.Row(c).Gap(5).Children(func() {
			widget.Dot(c, theme.Working, false, 6)
			ui.Text(c, fmt.Sprintf("%d working", working)).FontSize(11).SingleLine()
		})
		ui.Row(c).Gap(5).Background(theme.Fade(theme.Warning, .15*float64(notice))).Radius(4).Children(func() {
			widget.Dot(c, theme.Warning, false, 6)
			ui.Text(c, fmt.Sprintf("%d need attention", attention)).FontSize(11).TextColor(theme.Warning).SingleLine()
		})
		ui.Spacer(c)
		angle := float32(0)
		if v.expanded {
			angle = 180
		}
		ui.Icon(c, icon.ChevronDown).FontSize(12).Rotate(angle).TextColor(theme.Muted)
	})
	if b.Clicked() {
		v.expanded = !v.expanded
	}
}

func (v *View) details(c *ui.Context, o Options) {
	ui.Column(c).FillWidth().Grow(1).MinHeight(0).BorderWidth(1, 0, 0, 0).BorderColor(theme.Border).Children(func() {
		ui.Row(c).FillWidth().Height(42).Shrink(0).PaddingX(12).Gap(8).Children(func() {
			ui.Text(c, "Host").FontSize(11).TextColor(theme.Muted)
			ui.Column(c).Grow(1).MinWidth(0).Children(func() { v.hostFilter(c.Key("host-filter"), o.Hosts) })
		})
		var list []Session
		for _, s := range o.Sessions {
			if v.host == "" || v.host == s.HostID {
				list = append(list, s)
			}
		}
		if v.holding {
			order := map[string]int{}
			for i, key := range v.order {
				order[key] = i + 1
			}
			slices.SortStableFunc(list, func(a, b Session) int {
				x, y := order[a.HostID+":"+a.ID], order[b.HostID+":"+b.ID]
				if x == 0 {
					x = len(v.order) + 1
				}
				if y == 0 {
					y = len(v.order) + 1
				}
				return x - y
			})
		}
		v.order = v.order[:0]
		for _, s := range list {
			v.order = append(v.order, s.HostID+":"+s.ID)
		}
		if len(list) == 0 {
			ui.Column(c).Grow(1).FillWidth().Center().Padding(16).Children(func() { ui.Text(c, "No sessions on this host.").FontSize(12).TextColor(theme.Muted) })
		} else {
			box := ui.List(c.Key("sessions"), &v.list, len(list), func(i int) { v.row(c.Key(list[i].HostID+":"+list[i].ID), list[i], o) }).Grow(1).FillWidth().MinHeight(0).PaddingX(7)
			v.holding = box.Hovered() || box.FocusWithin()
		}
		ui.Row(c).FillWidth().Shrink(0).Padding(6, 10, 10, 10).Children(func() {
			if widget.GhostButton(c.Key("all"), "Open main window").FillWidth().Height(32).Clicked() && o.Restore != nil {
				o.Restore()
			}
		})
	})
}

func (v *View) row(c *ui.Context, s Session, o Options) {
	b := ui.ButtonBase(c).FillWidth().Height(52).PaddingX(9).Gap(10).Radius(7).Label("Open " + s.Name + " on " + s.HostName).Cursor(ui.CursorPointer)
	if b.Hovered() {
		b.Background(theme.Hover)
	}
	b.Children(func() {
		svg := s.Icon
		if svg == nil {
			svg = icon.Terminal
		}
		ui.Box(c).Size(28, 28).Center().Radius(6).Background(theme.Hover).Children(func() { ui.Icon(c, svg).FontSize(17).TextColor(theme.TextSoft) })
		ui.Column(c).Grow(1).MinWidth(0).Gap(3).Children(func() {
			ui.Text(c, s.Name+" · "+s.Project).FontSize(11.5).SingleLine().Shrink(1)
			ui.Text(c, s.HostName+" · "+s.Status.label()).FontSize(10.5).TextColor(theme.Muted).SingleLine().Shrink(1)
		})
		col := theme.Muted
		switch s.Status {
		case NeedsInput, Failed:
			col = theme.Warning
		case Working:
			col = theme.Working
		case Done:
			col = theme.Done
		}
		widget.Dot(c, col, false, 6)
	})
	if b.Clicked() && o.Open != nil {
		o.Open(s.HostID, s.ID)
	}
}
