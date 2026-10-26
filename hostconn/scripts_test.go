package hostconn

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/enowx/enowx/internal/quietcmd"
)

func TestRemoteScriptsParseFromWindowsCheckout(t *testing.T) {
	bash, err := exec.LookPath("bash")
	for name, source := range map[string]string{"prepare": prepareSource, "linger": lingerSource} {
		t.Run(name, func(t *testing.T) {
			// Reproduce Git's Windows checkout even when the test runs on Unix.
			windows := strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\n", "\r\n")
			script := normalizeShellScript(windows)
			if strings.ContainsRune(script, '\r') {
				t.Fatal("remote script retains Windows line endings")
			}
			if err != nil {
				t.Skip("bash is unavailable for syntax checking")
			}
			cmd := quietcmd.Command(bash, "-n")
			cmd.Stdin = strings.NewReader(script)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Fatalf("remote script syntax: %v\n%s", err, out)
			}
		})
	}
}
