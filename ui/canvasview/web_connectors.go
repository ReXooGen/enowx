package canvasview

import "strings"

// Native browser surfaces can cover GPU overlays. Paint connectors in the
// document too, so the live canvas and its exported thumbnail agree.
func connectorDocument(doc, svg string) string {
	overlay := `<div data-enowx-connectors style="position:absolute;left:0;top:0;pointer-events:none;z-index:2147483647">` + svg + `</div>`
	if i := strings.LastIndex(strings.ToLower(doc), "</body>"); i >= 0 {
		return doc[:i] + overlay + doc[i:]
	}
	return doc + overlay
}

func (w *Web) updateConnectors(svg string) {
	w.connectors = svg
	// Keep a page that is still being created current as well.
	if i := strings.Index(w.pending, `<div data-enowx-connectors`); i >= 0 {
		if end := strings.Index(w.pending[i:], "</div>"); end >= 0 {
			w.pending = w.pending[:i] + `<div data-enowx-connectors style="position:absolute;left:0;top:0;pointer-events:none;z-index:2147483647">` + svg + "</div>" + w.pending[i+end+6:]
		}
	}
	if w.page == nil {
		return
	}
	page, shown := w.page, w.shown
	go func() {
		value, err := page.Eval(`(function(){var n=document.querySelector('[data-enowx-connectors]');if(!n)return false;n.innerHTML=` + js(svg) + `;return true})()`)
		if ok, _ := value.(bool); err != nil || !ok {
			if w.Update == nil {
				return
			}
			w.Update(func() {
				if w.shown == shown && w.connectors == svg {
					w.connectors = ""
				}
			})
		}
	}()
}
