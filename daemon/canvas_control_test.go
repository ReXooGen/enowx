package daemon

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/core/session"
	"github.com/enowx/enowx/proto"
	"golang.org/x/term"
)

// A real PTY reader lets the protocol test verify both the paste and Enter,
// without launching an installed CLI or spending a model request.
func TestCanvasControlProcess(t *testing.T) {
	if os.Getenv("ENOWX_TEST_CANVAS_INPUT") != "1" {
		return
	}
	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer term.Restore(int(os.Stdin.Fd()), state)
	fmt.Println("canvas reader ready")
	p := make([]byte, 4096)
	for {
		n, err := os.Stdin.Read(p)
		fmt.Printf("%x", p[:n])
		if err != nil {
			return
		}
	}
}

func TestCanvasPauseResumeDeliversContinuation(t *testing.T) {
	srv := startServer(t)
	c := connect(t, srv, nil)
	cv, err := c.CreateCanvas(proto.CanvasCreateParams{Title: "Resume task"})
	if err != nil {
		t.Fatal(err)
	}
	start := func(canvas string) (*session.Session, *screen) {
		t.Helper()
		sess, err := srv.sessions.Create(session.Spec{
			Command: []string{os.Args[0], "-test.run=^TestCanvasControlProcess$"},
			Env:     []string{"ENOWX_TEST_CANVAS_INPUT=1"}, Dir: t.TempDir(), Cols: 200, Rows: 30,
			Meta: session.Meta{Agent: "omp", CanvasFor: canvas},
		})
		if err != nil {
			t.Fatal(err)
		}
		st, err := c.Attach(sess.ID())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { st.Close() })
		out := watch(st)
		out.waitFor(t, "canvas reader ready")
		return sess, out
	}
	sess, out := start(cv.ID) // No canvas edit yet: Canvas.Agent is still empty.
	_, other := start("another-canvas")
	if _, err := c.ControlCanvas(cv.ID, proto.CanvasPause); err != nil {
		t.Fatal(err)
	}
	preset, _ := agent.Find("omp")
	out.waitFor(t, hex.EncodeToString([]byte(preset.InterruptKey())))
	_, replacement := start(cv.ID)
	title := "User edit during pause"
	if _, err := c.EditCanvas(proto.CanvasEdit{ID: cv.ID, Op: proto.CanvasOpMeta, Title: &title}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ControlCanvas(cv.ID, proto.CanvasResume); err != nil {
		t.Fatal(err)
	}
	needle := hex.EncodeToString([]byte("The user resumed this canvas."))
	out.waitFor(t, needle)
	out.waitFor(t, hex.EncodeToString([]byte("User edit during pause")))
	out.waitFor(t, hex.EncodeToString([]byte(pasteEnd+"\r")))
	if _, err := c.ControlCanvas(cv.ID, proto.CanvasResume); err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	got := strings.ReplaceAll(sess.Terminal().Text(), "\n", "")
	if strings.Count(got, needle) != 1 {
		t.Fatalf("resume sent multiple continuation prompts: %q", got)
	}
	other.mu.Lock()
	defer other.mu.Unlock()
	if strings.Contains(other.buf.String(), needle) {
		t.Fatal("resume sent to an unrelated conversation")
	}
	replacement.mu.Lock()
	defer replacement.mu.Unlock()
	if strings.Contains(replacement.buf.String(), needle) {
		t.Fatal("resume switched from the paused conversation to a newer one")
	}
	current, err := c.Canvas(cv.ID)
	if err != nil || current.Paused {
		t.Fatalf("canvas remains paused: %+v, %v", current.CanvasInfo, err)
	}
}

func TestCanvasResumeMissingConversationStaysPaused(t *testing.T) {
	srv := startServer(t)
	c := connect(t, srv, nil)
	cv, err := c.CreateCanvas(proto.CanvasCreateParams{Title: "Lost conversation"})
	if err != nil {
		t.Fatal(err)
	}
	srv.canvasState.setPaused(cv.ID, true)
	srv.canvasState.pausedSession = map[string]string{cv.ID: "removed-session"}
	if _, err := c.ControlCanvas(cv.ID, proto.CanvasResume); err == nil || !strings.Contains(err.Error(), "no longer available") {
		t.Fatalf("resume silently accepted an undeliverable prompt: %v", err)
	}
	current, err := c.Canvas(cv.ID)
	if err != nil || !current.Paused {
		t.Fatalf("failed resume cleared pause: %+v, %v", current.CanvasInfo, err)
	}
}
