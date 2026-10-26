package daemon

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/enowx/enowx/proto"
)

const inputPromptTestEnv = "ENOWX_TEST_INPUT_PROMPT"

func serveInputPromptTest() {
	if os.Getenv(inputPromptTestEnv) != "1" {
		return
	}
	fmt.Print("\x1b[2J\x1b[HChoose a theme\r\nDark\r\nLight\r\nspace select · ↑/↓ move · esc cancel")
	r := bufio.NewReader(os.Stdin)
	_, _ = r.ReadString('\n')
	fmt.Print("\x1b[2J\x1b[HResponse received. Continuing...\r\n")
	_, _ = r.ReadString('\n')
	os.Exit(0)
}

func TestInputPromptStatusTravelsThroughTheDaemon(t *testing.T) {
	for _, attached := range []bool{false, true} {
		t.Run(fmt.Sprintf("attached=%v", attached), func(t *testing.T) {
			srv := startServer(t)
			events := make(chan proto.SessionInfo, 64)
			c := connect(t, srv, func(ev proto.Event) {
				if ev.Name == proto.EventSessionUpdated {
					var s proto.SessionInfo
					if json.Unmarshal(ev.Data, &s) == nil {
						select {
						case events <- s:
						default:
						}
					}
				}
			})
			exe, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			info, err := c.CreateSession(proto.SessionCreateParams{Command: []string{exe}, Env: []string{inputPromptTestEnv + "=1"}, Dir: t.TempDir(), Cols: 100, Rows: 12})
			if err != nil {
				t.Fatal(err)
			}
			if attached {
				st, err := c.Attach(info.ID)
				if err != nil {
					t.Fatal(err)
				}
				defer st.Close()
				go func() { _, _ = io.Copy(io.Discard, st) }()
			}
			waitFor(t, "interactive dialog state", func() bool {
				list, err := c.Sessions()
				return err == nil && len(list) == 1 && list[0].State == proto.StateWaiting
			})
			found := false
			for len(events) > 0 {
				if s := <-events; s.ID == info.ID && s.State == proto.StateWaiting {
					found = true
				}
			}
			if !found {
				t.Fatal("waiting state was not broadcast to clients")
			}
			if err := c.Input(info.ID, "answer", false, true); err != nil {
				t.Fatal(err)
			}
			waitFor(t, "answered dialog no longer waiting", func() bool {
				list, err := c.Sessions()
				return err == nil && len(list) == 1 && list[0].State != proto.StateWaiting
			})
		})
	}
}
