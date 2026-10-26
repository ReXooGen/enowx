//go:build windows

package ui

import (
	"syscall"
	"unsafe"

	"github.com/egoist/mygo"
)

// Windows 11 window attributes and values this file needs. They are DWM's
// own numbers, not MyGo's: MyGo keeps its copies unexported.
const (
	// dwmwaWindowCornerPreference is DWMWA_WINDOW_CORNER_PREFERENCE.
	dwmwaWindowCornerPreference = 33
	// dwmwcpDoNotRound is DWMWCP_DONOTROUND: square corners, no outline.
	dwmwcpDoNotRound = 1
)

var (
	dwmapi               = syscall.NewLazyDLL("dwmapi.dll")
	procDwmSetWindowAttr = dwmapi.NewProc("DwmSetWindowAttribute")
)

// dropFloatingFrame tells Windows not to round — and so not to outline — the
// control's window.
//
// MyGo rounds a Panel's corners as it creates it (internal/windows/window.go,
// `if o.Panel { w.roundFlyout() }`), which sets DWMWA_WINDOW_CORNER_PREFERENCE
// to DWMWCP_ROUND (internal/windows/flyout.go). Windows 11 clips, and
// outlines, any window it rounds: that outline is the hairline around the
// whole square canvas, well outside the pie. Nothing clears it afterwards —
// SetVibrancy writes only DWMWA_SYSTEMBACKDROP_TYPE and the frame extension —
// so it is still set when the panel first shows.
//
// Square corners keep the Panel, and with it WS_EX_TOOLWINDOW (out of
// Alt-Tab), the owner that keeps the button off every taskbar, and the
// creation-time WS_EX_TOPMOST. DWMWCP_DONOTROUND draws no outline at all, so
// only the pie shows on the desktop.
//
// Best effort: Windows 10 knows the attribute and ignores it, and a build
// without DWM fails the call. Either way the window is as it was, and the
// hairline is the only thing left over.
func dropFloatingFrame(w *mygo.Window) {
	if w == nil || procDwmSetWindowAttr.Find() != nil {
		return
	}
	hwnd := w.NativeHandle()
	if hwnd == 0 {
		return
	}
	pref := int32(dwmwcpDoNotRound)
	// The HRESULT is deliberately ignored: on Windows 10 and without DWM
	// this fails harmlessly, and there is nothing to fall back to.
	procDwmSetWindowAttr.Call(hwnd, dwmwaWindowCornerPreference, uintptr(unsafe.Pointer(&pref)), unsafe.Sizeof(pref))
}
