package ui

import (
	"path/filepath"
	"testing"

	"github.com/egoist/mygo"
	"github.com/enowx/enowx/client"
	"github.com/enowx/enowx/core/hosts"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/minidockview"
)

func TestMiniDockSettingPersistsAndDisablingHidesIt(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	a := New(Options{StateFile: p})
	a.host = LocalHost()
	a.showMiniDock()
	if a.miniDock.shown {
		t.Fatal("mini dock was enabled by default")
	}
	tt := floatingSectionFrame(t, a, a.miniDockSettings)
	if err := tt.Click("Mini dock switch"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	a.showMiniDock()
	if !a.miniDock.shown {
		t.Fatal("enabled mini dock did not show")
	}
	if err := tt.Click("Mini dock motion switch"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	b := New(Options{StateFile: p})
	if !b.state.MiniDock || !b.state.MiniDockReduceMotion {
		t.Fatal("mini dock settings were not saved")
	}
	if err := tt.Click("Mini dock switch"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if a.miniDock.shown || loadState(p).MiniDock {
		t.Fatal("disabled mini dock stayed visible or enabled")
	}
	a.quitting = true
	a.state.MiniDock = true
	a.showMiniDock()
	if a.miniDock.shown {
		t.Fatal("mini dock appeared while quitting")
	}
}

func TestMiniDockSnapshotsKeepHostIdentityAndStaleState(t *testing.T) {
	a := New(Options{})
	a.host = LocalHost()
	a.client = new(client.Client)
	a.sessions = []proto.SessionInfo{{ID: "same", Agent: "codex", Project: "/work/local", State: proto.StateWorking},
		{ID: "tab", Parent: "same", Agent: "shell"}, {ID: "note", NotesFor: "note"}}
	a.hostList = []hosts.Host{{ID: "vps-1", Name: "Production"}, {ID: "vps-2", Name: "Production"}}
	a.watches = map[string]*hostWatch{
		"vps-1": {status: hostStatus{state: hostOnline, sessions: []proto.SessionInfo{{ID: "same", Agent: "omp", Project: "/srv/remote", State: proto.StateWaiting}}}},
		"vps-2": {status: hostStatus{state: hostOffline, sessions: []proto.SessionInfo{{ID: "stale", State: proto.StateWorking}}}},
	}
	list, sources := a.miniDockSessions()
	if len(list) != 3 || len(sources) != 3 {
		t.Fatalf("sessions %d, hosts %d", len(list), len(sources))
	}
	if list[0].ID != "same" || list[0].HostID != "vps-1" || list[0].Status != minidockview.NeedsInput {
		t.Fatalf("attention session: %+v", list[0])
	}
	if list[1].HostID != hosts.LocalID || list[1].Status != minidockview.Working {
		t.Fatalf("local session: %+v", list[1])
	}
	if list[2].Status != minidockview.Offline {
		t.Fatal("stale host still reports a working session")
	}
}

func TestMiniDockBoundsExpandUpwardsAndStayOnScreen(t *testing.T) {
	area := mygo.Rectangle{X: -1920, Y: 0, Width: 1920, Height: 1080}
	old := mygo.Rectangle{X: -400, Y: 900, Width: 344, Height: 100}
	expanded := miniDockBounds(old, area, minidockview.ExpandedHeight)
	if expanded.Y+expanded.Height != old.Y+old.Height {
		t.Fatal("expanding moved the bottom edge")
	}
	for _, r := range []mygo.Rectangle{old, {X: 5000, Y: -500, Width: 344, Height: 100}} {
		b := miniDockBounds(r, area, minidockview.ExpandedHeight)
		if b.X < area.X || b.Y < area.Y || b.X+b.Width > area.X+area.Width || b.Y+b.Height > area.Y+area.Height {
			t.Fatalf("off screen: %+v", b)
		}
	}
}
