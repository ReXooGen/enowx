package ui

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"github.com/enowx/enowx/ui/minidockview"
)

// miniDockController belongs to the app's UI thread, including window events.
type miniDockController struct {
	window     *mygo.Window
	view       minidockview.View
	shown      bool
	positioned bool
}

func (a *App) setUpMiniDock() {
	a.win.OnMinimize(a.showMiniDock)
	a.win.OnRestore(a.hideMiniDock)
	a.win.OnShow(a.hideMiniDock)
	a.win.OnClose(func(*mygo.CloseEvent) { a.hideMiniDock() })
	mygo.App.OnWillQuit(func(*mygo.QuitEvent) {
		if a.miniDock.window != nil {
			a.miniDock.window.Destroy()
		}
	})
	offDisplays := mygo.Screen.OnDisplaysChanged(func() {
		if a.miniDock.shown {
			a.resizeMiniDock(a.miniDock.window.Bounds().Height)
		}
	})
	a.win.OnClosed(offDisplays)
}

func (a *App) setMiniDock(enabled bool) {
	a.state.MiniDock = enabled
	if !enabled {
		a.hideMiniDock()
	} else if a.win != nil && a.win.IsMinimized() {
		a.showMiniDock()
	}
	a.saveState()
}

func (a *App) showMiniDock() {
	if !a.state.MiniDock || a.quitting {
		return
	}
	d := &a.miniDock
	d.shown = true
	d.view.Show()
	a.hideFloating()
	if a.win == nil {
		return
	} // the same state transition is exercised headlessly
	if d.window == nil {
		d.window = mygo.NewWindow(mygo.WindowOptions{Title: "enowX mini dock", Panel: true, Frameless: true,
			Transparent: true, Vibrancy: mygo.VibrancyMica, Hidden: true, DisableShadow: true,
			DisableResize: true, DisableMinimize: true, DisableMaximize: true, DisableFullScreen: true,
			Width: minidockview.Width, Height: minidockview.CompactHeight, Content: ui.View(a.miniDockView)})
		d.window.SetVibrancy(mygo.VibrancyNone)
		dropFloatingFrame(d.window)
		d.window.OnClose(func(e *mygo.CloseEvent) {
			if a.quitting {
				return
			}
			e.PreventDefault()
			a.showWindow() // closing the dock never strands a minimized main window
		})
		d.window.OnMove(a.rememberMiniDockPosition)
	}
	a.resizeMiniDock(minidockview.CompactHeight)
	d.window.ShowInactive() // minimizing must leave focus with the other app
	d.window.Invalidate()
}

func (a *App) hideMiniDock() {
	a.miniDock.shown = false
	if a.miniDock.window != nil {
		a.miniDock.window.Hide()
	}
}

func (a *App) miniDockView(c *ui.Context) {
	c.SetTheme(a.theme)
	sessions, sources := a.miniDockSessions()
	a.miniDock.view.Draw(c, minidockview.Options{Sessions: sessions, Hosts: sources,
		ReduceMotion: a.state.MiniDockReduceMotion, Restore: a.showWindow, Open: a.openMiniDockSession, Resize: a.resizeMiniDock})
}

func (a *App) rememberMiniDockPosition() {
	if !a.miniDock.shown || a.miniDock.window == nil {
		return
	}
	b := a.miniDock.window.Bounds()
	a.state.MiniDockPosition = &miniDockPosition{X: b.X, Y: b.Y + b.Height}
	a.saveState()
}

func (a *App) resizeMiniDock(height int) {
	d := &a.miniDock
	if d.window == nil {
		return
	}
	b := d.window.Bounds()
	if !d.positioned {
		area := mygo.Screen.DisplayMatching(a.win.Bounds()).WorkArea
		b = mygo.Rectangle{X: area.X + area.Width - minidockview.Width - 24, Y: area.Y + area.Height - minidockview.CompactHeight - 24, Width: minidockview.Width, Height: minidockview.CompactHeight}
		if p := a.state.MiniDockPosition; p != nil {
			b.X, b.Y = p.X, p.Y-b.Height
		}
		d.positioned = true
	}
	area := mygo.Screen.DisplayMatching(b).WorkArea
	next := miniDockBounds(b, area, height)
	if next != d.window.Bounds() {
		d.window.SetBounds(next)
	}
}

// miniDockBounds preserves the bottom edge as the list opens upwards, and
// brings a remembered position back on screen after displays change.
func miniDockBounds(old, area mygo.Rectangle, height int) mygo.Rectangle {
	w, h := min(minidockview.Width, area.Width), min(height, area.Height)
	x := max(area.X, min(old.X, area.X+area.Width-w))
	y := max(area.Y, min(old.Y+old.Height-h, area.Y+area.Height-h))
	return mygo.Rectangle{X: x, Y: y, Width: w, Height: h}
}
