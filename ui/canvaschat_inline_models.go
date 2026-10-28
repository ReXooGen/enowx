package ui

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

func (a *App) canvasInlineModels(canvasID, agentID string, session proto.SessionInfo, active bool) func(*ui.Context) {
	if agentID != "omp" {
		return nil
	}
	return func(c *ui.Context) {
		hc := a.canvasChatOf(canvasID)
		if hc.modelRequest == 0 || hc.modelsGeneration != a.gen {
			if selector := a.notesModel[agentID]; selector != "" {
				hc.providerDraft, hc.modelDraft, _ = strings.Cut(selector, "/")
			}
			a.loadCanvasModels(hc, agentID)
		}
		busy := active && isWorking(session)
		providers := []string{}
		for _, m := range hc.models {
			if !slices.Contains(providers, m.Provider) {
				providers = append(providers, m.Provider)
			}
		}
		slices.Sort(providers)
		changed := false
		ui.Column(c.Key("provider")).Width(0).Grow(1).Shrink(1).MinWidth(0).Gap(3).Disabled(busy || hc.modelsLoading || hc.modelsError != "").Children(func() {
			ui.Text(c, "Provider").FontSize(10.5).TextColor(theme.Muted)
			if widget.Choice(c, "Canvas provider", &hc.providerDraft, providers) {
				selectCanvasRegisteredModel(hc)
				changed = true
			}
		})
		// Recompute after changing provider in this frame.
		models := []string{}
		for _, m := range hc.models {
			if m.Provider == hc.providerDraft {
				models = append(models, m.ID)
			}
		}
		ui.Column(c.Key("model")).Width(0).Grow(1).Shrink(1).MinWidth(0).Gap(3).Disabled(busy || hc.modelsLoading || hc.modelsError != "" || hc.providerDraft == "").Children(func() {
			ui.Text(c, "Model").FontSize(10.5).TextColor(theme.Muted)
			if widget.Choice(c, "Canvas model", &hc.modelDraft, models) {
				changed = true
			}
		})
		if hc.modelsLoading {
			widget.Spinner(c, 12)
		} else {
			refresh := widget.IconButton(c, icon.Sync, "Refresh provider/model list", false)
			if hc.modelsError != "" {
				refresh.Tooltip(hc.modelsError)
			}
			if refresh.Clicked() {
				a.loadCanvasModels(hc, agentID)
			}
		}
		selector, valid := canvasModelID(hc.providerDraft, hc.modelDraft)
		if changed && valid && !busy && canvasModelRegistered(hc.models, selector) {
			a.notesModel = setPick(a.notesModel, agentID, selector)
			if active {
				a.setAgentSetting(session, "model", selector)
			}
		}
	}
}
