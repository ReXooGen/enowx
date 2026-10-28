package ui

import (
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

func (a *App) loadCanvasModels(hc *hostCanvasChat, agentID string) {
	hc.modelRequest++
	hc.modelsGeneration = a.gen
	generation := a.gen
	request := hc.modelRequest
	hc.modelsLoading, hc.modelsError = true, ""
	cl := a.client
	if cl == nil {
		hc.modelsLoading = false
		hc.modelsError = "Connect to this host to read its configured models."
		return
	}
	p := proto.AgentModelsParams{Agent: agentID, Canvas: hc.canvas}
	if session, ok := a.canvasSession(hc.canvas, agentID); ok && !session.Exited {
		p.Session = session.ID
	}
	go func() {
		catalog, err := cl.AgentModels(p)
		a.onMain(func() {
			if hc.modelRequest != request || a.gen != generation {
				return
			}
			hc.modelsLoading = false
			if err != nil {
				hc.modelsError = "Could not read this host's CLI models: " + err.Error()
				return
			}
			if !catalog.Registered {
				hc.modelsError = "This host needs an updated daemon to list OMP's registered models."
				return
			}
			hc.models = catalog.Models
			if len(hc.models) == 0 {
				hc.modelsError = "OMP has no registered chat models on this host. Configure a provider in OMP, then refresh."
				return
			}
			selectCanvasRegisteredModel(hc)
		})
	}()
}

func selectCanvasRegisteredModel(hc *hostCanvasChat) {
	selector, _ := canvasModelID(hc.providerDraft, hc.modelDraft)
	if canvasModelRegistered(hc.models, selector) {
		return
	}
	for _, m := range hc.models {
		if m.Provider == hc.providerDraft {
			hc.modelDraft = m.ID
			return
		}
	}
	if len(hc.models) > 0 {
		hc.providerDraft, hc.modelDraft = hc.models[0].Provider, hc.models[0].ID
	}
}

func canvasModelRegistered(models []proto.AgentModel, selector string) bool {
	return slices.ContainsFunc(models, func(m proto.AgentModel) bool { return m.Selector == selector && m.Provider+"/"+m.ID == selector })
}

func (a *App) canvasRegisteredModels(c *ui.Context, hc *hostCanvasChat, agentID string) {
	if hc.modelsLoading {
		ui.Row(c).Gap(7).AlignItems(ui.Center).Children(func() {
			widget.Spinner(c, 12)
			ui.Text(c, "Reading configured models from this host…").FontSize(12).TextColor(theme.Muted)
		})
	} else if hc.modelsError != "" {
		ui.Text(c, hc.modelsError).FontSize(12).TextColor(theme.Error)
	} else {
		providers := []string{}
		for _, model := range hc.models {
			if !slices.Contains(providers, model.Provider) {
				providers = append(providers, model.Provider)
			}
		}
		slices.Sort(providers)
		widget.Field(c, "Provider", func() {
			if widget.Choice(c.Key("registered-provider"), "Registered provider", &hc.providerDraft, providers) {
				selectCanvasRegisteredModel(hc)
			}
		})
		models := []string{}
		for _, model := range hc.models {
			if model.Provider == hc.providerDraft {
				models = append(models, model.ID)
			}
		}
		widget.Field(c, "Model", func() { widget.Choice(c.Key("registered-model"), "Registered model", &hc.modelDraft, models) })
		for _, model := range hc.models {
			if model.Provider == hc.providerDraft && model.ID == hc.modelDraft && !strings.EqualFold(model.Name, model.ID) {
				ui.Text(c, model.Name).FontSize(12).TextColor(theme.Muted)
				break
			}
		}
	}
	if widget.GhostButton(c, "Refresh models").Disabled(hc.modelsLoading).Clicked() {
		a.loadCanvasModels(hc, agentID)
	}
}
