package ui

import (
	"slices"

	"github.com/enowx/enowx/core/hosts"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/minidockview"
	"github.com/enowx/enowx/ui/textfmt"
)

// miniDockSessions reads the same daemon snapshots as the host rail. Keep the
// host alongside each ID: different daemons can have the same session ID.
func (a *App) miniDockSessions() ([]minidockview.Session, []minidockview.Host) {
	var out []minidockview.Session
	var sources []minidockview.Host
	for _, h := range a.allHosts() {
		name := h.Name
		if h.ID == hosts.LocalID {
			name = "This PC"
		}
		sources = append(sources, minidockview.Host{ID: h.ID, Name: name})
		st, known := a.statusOfHost(h)
		online := known && st.state == hostOnline
		for _, s := range st.sessions {
			if s.Parent != "" || s.NotesFor != "" || s.CanvasFor != "" {
				continue
			}
			project := textfmt.LastPathPart(floatingProjectPath(s))
			if project == "" {
				project = "Session"
			}
			out = append(out, minidockview.Session{ID: s.ID, HostID: h.ID, HostName: name,
				Name: a.floatingSessionName(s), Project: project, Icon: floatingSessionIcon(s),
				Status: miniDockStatus(s, online), Created: s.Created})
		}
	}
	slices.SortStableFunc(out, func(a, b minidockview.Session) int {
		if a.Status != b.Status {
			return int(a.Status) - int(b.Status)
		}
		return b.Created.Compare(a.Created)
	})
	return out, sources
}

func miniDockStatus(s proto.SessionInfo, online bool) minidockview.Status {
	switch {
	case !online:
		return minidockview.Offline
	case s.Hibernated:
		return minidockview.Sleeping
	case s.Exited && s.Restored && s.ExitCode == -1:
		return minidockview.Stopped
	case s.Exited && s.ExitCode != 0:
		return minidockview.Failed
	case s.Exited:
		return minidockview.Stopped
	case s.Queued:
		return minidockview.Queued
	case s.State == proto.StateWaiting:
		return minidockview.NeedsInput
	case isWorking(s):
		return minidockview.Working
	case s.State == proto.StateDone:
		return minidockview.Done
	default:
		return minidockview.Idle
	}
}

func (a *App) openMiniDockSession(hostID, id string) {
	for _, h := range a.allHosts() {
		if h.ID != hostID {
			continue
		}
		a.page = ""
		if h.ID != a.host.ID || a.client == nil {
			a.openSession = id
			a.connect(h)
		} else if s, ok := a.sessionByID(id); ok {
			a.selectSession(s)
		}
		a.showWindow()
		return
	}
}
