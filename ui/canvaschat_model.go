package ui

import (
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/agent"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

func (a *App) canvasChooseModel(canvasID, agentID string) func() {
	if len(agent.SettingsOf(agentID).ModelFlag) == 0 {
		return nil
	}
	return func() { a.openCanvasModelPicker(canvasID, agentID) }
}

func (a *App) openCanvasModelPicker(canvasID, agentID string) {
	hc := a.canvasChatOf(canvasID)
	hc.modelDraft = a.notesModel[agentID]
	hc.providerDraft = ""
	if agentID == "omp" {
		if provider, model, ok := strings.Cut(hc.modelDraft, "/"); ok {
			hc.providerDraft, hc.modelDraft = provider, model
		}
	}
	hc.modelOpen = true
	if agentID == "omp" {
		a.loadCanvasModels(hc, agentID)
	}
}

func (a *App) canvasModelDialog(c *ui.Context, hc *hostCanvasChat, agentID string, s proto.SessionInfo, running bool) {
	if !hc.modelOpen {
		return
	}
	if agentID == "omp" && hc.modelsGeneration != a.gen {
		a.loadCanvasModels(hc, agentID)
	}
	settings := agent.SettingsOf(agentID)
	widget.Dialog(c, &hc.modelOpen, 420, 12, func() {
		ui.Text(c, "Model / provider").FontSize(16).FontWeight(600).TextColor(theme.TextStrong)
		ui.Text(c, "Use a model available in this agent's configuration on this host.").FontSize(12).TextColor(theme.Muted)
		if agentID == "omp" {
			a.canvasRegisteredModels(c, hc, agentID)
		} else {
			widget.Field(c, "Model", func() {
				widget.Input(c, &hc.modelDraft, "Model ID").Label("Model ID").TextColor(theme.Text).AutoFocus()
			})
		}
		if agentID != "omp" && len(settings.Models) > 0 {
			labels := []string{}
			for _, model := range settings.Models {
				labels = append(labels, model.ID)
			}
			widget.Choice(c, "Known models", &hc.modelDraft, labels)
		}
		busy := running && isWorking(s)
		if busy {
			ui.Text(c, "The current task is still running. Apply the change after it finishes.").FontSize(12).TextColor(theme.Muted)
		}
		widget.DialogFooter(c, "", func() {
			if widget.GhostButton(c, "Cancel").Clicked() {
				hc.modelOpen = false
			}
			model, valid := canvasModelID(hc.providerDraft, hc.modelDraft)
			if agentID == "omp" {
				valid = valid && !hc.modelsLoading && hc.modelsError == "" && canvasModelRegistered(hc.models, model)
			}
			unsupported := len(settings.ModelFlag) == 0 || running && settings.ModelCommand == ""
			if widget.PrimaryButton(c, nil, "Apply model", "").Disabled(!valid || busy || unsupported).Clicked() {
				a.notesModel = setPick(a.notesModel, agentID, model)
				if running {
					a.setAgentSetting(s, "model", model)
				}
				hc.modelOpen = false
			}
			if agentID == "omp" && running && widget.GhostButton(c, "Browse in Terminal").Disabled(busy).Clicked() {
				hc.modelOpen = false
				hc.terminal = true
				a.sendToSession(s.ID, "/model")
			}
		})
	})
}

func canvasModelID(provider, model string) (string, bool) {
	provider, model = strings.TrimSpace(provider), strings.TrimSpace(model)
	if model == "" || strings.ContainsAny(provider+model, " \t\r\n\x00") {
		return "", false
	}
	if provider != "" {
		model = provider + "/" + model
	}
	return model, true
}
