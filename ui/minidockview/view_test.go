package minidockview

import (
	"fmt"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/ui/theme"
)

func dockTester(v *View, o Options) *ui.Tester {
	tt := ui.NewTester(func(c *ui.Context) { c.SetTheme(theme.Window()); v.Draw(c, o) }, Width, ExpandedHeight)
	tt.Frame()
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	return tt
}

func TestMiniDockExpandsFiltersAndOpensTheCorrectHost(t *testing.T) {
	var v View
	v.Show()
	opened := ""
	height := 0
	o := Options{Hosts: []Host{{ID: "local", Name: "This PC"}, {ID: "vps", Name: "VPS"}},
		Sessions: []Session{{ID: "same", HostID: "vps", HostName: "VPS", Name: "OMP", Project: "remote", Status: NeedsInput},
			{ID: "same", HostID: "local", HostName: "This PC", Name: "Codex", Project: "local", Status: Working}},
		Open: func(h, id string) { opened = h + ":" + id }, Resize: func(h int) { height = h }}
	tt := dockTester(&v, o)
	if !tt.HasText("1 working") || !tt.HasText("1 need attention") || tt.HasText("remote") {
		t.Fatalf("compact: %q", tt.Texts())
	}
	if err := tt.Click("Show sessions"); err != nil {
		t.Fatal(err)
	}
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	if height != ExpandedHeight || !tt.HasText("remote") || !tt.HasText("local") {
		t.Fatalf("expanded: %q", tt.Texts())
	}
	if err := tt.Click("Filter mini dock host"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if err := tt.Click("Filter VPS"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if tt.HasText("Codex · local") || !tt.HasText("OMP · remote") {
		t.Fatalf("filter: %q", tt.Texts())
	}
	if err := tt.Click("Open OMP on VPS"); err != nil {
		t.Fatal(err)
	}
	if opened != "vps:same" {
		t.Fatalf("opened %q", opened)
	}
	tt.TypeKey(0, ui.KeyEscape, "")
	tt.Advance(300 * time.Millisecond)
	tt.Frame()
	if height != CompactHeight || tt.HasText("remote") {
		t.Fatalf("collapse: %q", tt.Texts())
	}
}

func TestMiniDockManySessionsScrollAndRestore(t *testing.T) {
	v := View{expanded: true}
	opened := ""
	restored := false
	o := Options{ReduceMotion: true, Hosts: []Host{{ID: "vps", Name: "VPS"}}, Open: func(_, id string) { opened = id }, Restore: func() { restored = true }}
	for i := 0; i < 100; i++ {
		o.Sessions = append(o.Sessions, Session{ID: fmt.Sprint(i), HostID: "vps", HostName: "VPS", Name: fmt.Sprintf("Agent %d", i), Project: "project", Status: Working})
	}
	tt := dockTester(&v, o)
	v.list.ScrollToEnd()
	tt.Frame()
	tt.Frame()
	if err := tt.Click("Open Agent 99 on VPS"); err != nil {
		t.Fatalf("last session inaccessible: %v; %q", err, tt.Texts())
	}
	if opened != "99" {
		t.Fatalf("opened %q", opened)
	}
	if err := tt.Click("Open enowX"); err != nil {
		t.Fatal(err)
	}
	if !restored {
		t.Fatal("restore button did nothing")
	}
}

func TestMiniDockEmptyHostAndReducedMotion(t *testing.T) {
	v := View{}
	v.Show()
	height := 0
	o := Options{ReduceMotion: true, Hosts: []Host{{ID: "local", Name: "This PC"}}, Resize: func(h int) { height = h }}
	tt := dockTester(&v, o)
	if err := tt.Click("Show sessions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if height != ExpandedHeight || !tt.HasText("No sessions on this host.") {
		t.Fatalf("empty: %q", tt.Texts())
	}
	if err := tt.Click("Collapse sessions"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if height != CompactHeight {
		t.Fatal("reduced motion did not collapse immediately")
	}
}

func TestMiniDockDoesNotMoveRowsUnderThePointer(t *testing.T) {
	v := View{expanded: true}
	o := Options{ReduceMotion: true, Sessions: []Session{
		{ID: "1", HostID: "local", HostName: "This PC", Name: "First", Status: Working},
		{ID: "2", HostID: "local", HostName: "This PC", Name: "Second", Status: Working},
	}}
	tt := ui.NewTester(func(c *ui.Context) { v.Draw(c, o) }, Width, ExpandedHeight)
	tt.Frame()
	before, ok := tt.Find("Open First on This PC")
	if !ok {
		t.Fatal("missing first row")
	}
	tt.Move(before.X+before.W/2, before.Y+before.H/2)
	tt.Frame()
	o.Sessions[1].Status = NeedsInput
	o.Sessions[0], o.Sessions[1] = o.Sessions[1], o.Sessions[0]
	tt.Frame()
	after, _ := tt.Find("Open First on This PC")
	if before.Y != after.Y {
		t.Fatal("status update moved the row under the pointer")
	}
	tt.Move(0, 0)
	tt.Frame()
	tt.Frame()
	after, _ = tt.Find("Open First on This PC")
	if before.Y == after.Y {
		t.Fatal("attention order was not restored after leaving the list")
	}
}
