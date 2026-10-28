package canvasview

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

//go:embed checks.js
var designChecksScript string

func (v *View) checksPanel(c *ui.Context, o Options) {
	w := v.workflow
	ui.Text(c, "Layout and readability").FontSize(13).FontWeight(600)
	ui.Text(c, "Inspect shape geometry and HTML at 1440, 768 and 390 px. Findings need review; decorative overlaps and image backgrounds may need a manual check. Preview copies do not execute scripts.").FontSize(12).TextColor(theme.Muted)
	if widget.GhostButton(c, "Check this design").FillWidth().Disabled(w.checking).Clicked() {
		w.issues = canvas.CheckShapes(v.frameOf())
		w.message = ""
		if o.Web != nil && v.frameOf().HTML != "" {
			w.checking = true
			o.Web.checkDesign(func(issues []canvas.CheckIssue, err error) {
				w.checking = false
				if err != nil {
					v.err = err.Error()
					return
				}
				w.issues = append(w.issues, issues...)
				w.message = fmt.Sprintf("%d findings", len(w.issues))
			})
		} else {
			w.message = fmt.Sprintf("%d findings", len(w.issues))
		}
	}
	if w.checking {
		ui.Text(c, "Checking three preview widths…").FontSize(12).TextColor(theme.Muted)
	}
	for i, issue := range w.issues {
		ui.Column(c.Key(fmt.Sprintf("finding-%d", i))).FillWidth().Padding(8).Gap(4).Radius(6).Background(theme.Sunken).Children(func() {
			ui.Text(c, fmt.Sprintf("%s · %d px", issue.Element, issue.Width)).FontSize(11.5).TextColor(theme.Warning)
			ui.Text(c, issue.Message).FontSize(12)
			if widget.GhostButton(c, "Select element").Height(25).Label("Select finding " + issue.Element).Clicked() {
				v.sel = issue.Element
				if o.Web != nil {
					o.Web.outline(issue.Element)
				}
			}
		})
	}
}

func (w *Web) checkDesign(done func([]canvas.CheckIssue, error)) {
	page, shown := w.page, w.shown
	if page == nil {
		done(nil, fmt.Errorf("open the mockup preview first"))
		return
	}
	go func() {
		var result struct {
			Issues []canvas.CheckIssue `json:"issues"`
			Error  string              `json:"error"`
		}
		_, err := page.Eval(designChecksScript + ";window.__enowxCheckStart();true")
		deadline := time.Now().Add(25 * time.Second)
		finished := false
		for err == nil && time.Now().Before(deadline) {
			var value any
			value, err = page.Eval("window.__enowxCheckResult")
			if err != nil {
				break
			}
			if value != nil {
				data, _ := json.Marshal(value)
				err = json.Unmarshal(data, &result)
				finished = true
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		if err == nil && !finished {
			err = fmt.Errorf("design checks timed out")
		}
		if err == nil && result.Error != "" {
			err = fmt.Errorf("design checks: %s", result.Error)
		}
		w.Update(func() {
			if w.shown != shown {
				done(nil, fmt.Errorf("design changed while checking; run checks again"))
				return
			}
			done(result.Issues, err)
		})
	}()
}
