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
	w.flushConnectors()
}

// At most one browser update is in flight; pointer motion replaces the queued
// geometry rather than building an out-of-order backlog of old positions.
func (w *Web) flushConnectors() {
	if w.page == nil || w.connectorBusy || w.Update == nil {
		return
	}
	page, shown, svg := w.page, w.shown, w.connectors
	// No new work is needed after a failed update until the next shown frame.
	if svg == "" {
		return
	}
	w.connectorBusy = true
	go func() {
		value, err := page.Eval(`(function(){var n=document.querySelector('[data-enowx-connectors]');if(!n)return false;n.innerHTML=` + js(svg) + `;return true})()`)
		w.Update(func() {
			w.connectorBusy = false
			if w.shown != shown || w.connectors != svg {
				w.flushConnectors()
				return
			}
			if ok, _ := value.(bool); err != nil || !ok {
				w.connectors = ""
			}
		})
	}()
}
