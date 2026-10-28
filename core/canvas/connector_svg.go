package canvas

import (
	"fmt"
	"strings"

	"github.com/enowx/enowx/proto"
)

// ConnectorSVG keeps connected shapes as geometry anchors while drawing only
// their lines, for overlays above a mockup's own browser surface.
func ConnectorSVG(f proto.CanvasFrame) string {
	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d">`, f.Width, f.Height, f.Width, f.Height)
	b.WriteString(`<defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse"><path d="M0,0 L10,5 L0,10 z" fill="context-stroke"/></marker></defs>`)
	for _, el := range f.Elements {
		if el.Type == proto.ShapeLine || el.Type == proto.ShapeArrow {
			svgElement(&b, f, el, nil)
		}
	}
	b.WriteString(`</svg>`)
	return b.String()
}
