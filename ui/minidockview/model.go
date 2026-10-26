// Package minidockview shows a compact overview of sessions while the main
// window is minimized. Its state belongs to the UI thread.
package minidockview

import (
	"time"

	"github.com/egoist/mygo/ui"
)

const (
	Width          = 344
	CompactHeight  = 100
	ExpandedHeight = 432
)

// Status orders sessions by the attention they need.
type Status int

const (
	NeedsInput Status = iota
	Failed
	Working
	Done
	Queued
	Sleeping
	Idle
	Stopped
	Offline
)

// Host is a session source, identified independently of its display name.
type Host struct{ ID, Name string }

// Session is a main session of one host; IDs are only unique within a host.
type Session struct {
	ID, HostID, HostName, Name, Project string
	Icon                                *ui.SVG
	Status                              Status
	Created                             time.Time
}

// Options supplies a fresh snapshot and the app actions available in the dock.
type Options struct {
	Sessions     []Session
	Hosts        []Host
	ReduceMotion bool
	Restore      func()
	Open         func(hostID, sessionID string)
	Resize       func(height int)
}

func (s Status) label() string {
	return [...]string{"Needs input", "Failed", "Working", "Done", "Queued", "Sleeping", "Idle", "Stopped", "Offline"}[s]
}

// View holds the dock's filter and disclosure state between frames.
type View struct {
	expanded      bool
	host          string
	list          ui.ListState
	fresh         bool
	lastAttention map[string]bool
	order         []string
	holding       bool
}

// Show starts a new compact presentation without forgetting the host filter.
func (v *View) Show() { v.fresh = true; v.expanded = false; v.lastAttention = nil }
