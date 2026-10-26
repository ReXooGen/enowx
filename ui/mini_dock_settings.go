package ui

import "github.com/egoist/mygo/ui"

func (a *App) miniDockSettings(c *ui.Context) {
	enabled := a.state.MiniDock
	if switchRowKey(c, "Mini dock switch", &enabled, "Mini dock when minimized", "Keep a small, movable overview of sessions over other apps. Click a session to return to it.") {
		a.setMiniDock(enabled)
	}
	if enabled {
		if switchRowKey(c, "Mini dock motion switch", &a.state.MiniDockReduceMotion, "Reduce mini dock motion", "Show and expand the dock without animation.") {
			a.saveState()
		}
	}
}
