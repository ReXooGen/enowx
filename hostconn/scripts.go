package hostconn

import (
	_ "embed"
	"strings"
)

//go:embed prepare.sh
var prepareSource string

//go:embed linger.sh
var lingerSource string

var prepareScript = normalizeShellScript(prepareSource)
var lingerScript = normalizeShellScript(lingerSource)

// Windows checkouts can embed CRLF; the remote shell needs Unix line endings.
func normalizeShellScript(script string) string {
	return strings.ReplaceAll(script, "\r\n", "\n")
}
