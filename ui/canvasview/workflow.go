package canvasview

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
	"github.com/enowx/enowx/ui/widget"
)

type workflowState struct {
	tab                        int
	library                    []proto.CanvasLibraryComponent
	name                       string
	loading                    bool
	message                    string
	transition                 proto.CanvasTransition
	duration, transitionSource string
	issues                     []canvas.CheckIssue
	checking                   bool
}

func (v *View) openWorkflow(o Options) {
	if v.workflow == nil {
		v.workflow = &workflowState{duration: "220", transition: proto.CanvasTransition{Trigger: "click", Effect: "fade", Easing: "ease-out", Duration: 220}}
	}
	if o.Library == nil {
		return
	}
	w := v.workflow
	w.loading = true
	o.Library(proto.CanvasLibraryParams{Op: "list"}, func(result proto.CanvasLibraryResult, err error) {
		w.loading = false
		if err != nil {
			v.err = err.Error()
			return
		}
		w.library = result.Components
	})
}

func (v *View) workflowPanel(c *ui.Context, o Options) {
	if v.workflow == nil {
		v.openWorkflow(o)
	}
	w := v.workflow
	ui.Text(c, "Canvas workflow").FontSize(13).FontWeight(600)
	names := []string{"Library", "Review", "Checks", "Arrange", "Motion", "Offline"}
	for row := 0; row < 2; row++ {
		ui.Row(c.Key(row)).FillWidth().Gap(4).Children(func() {
			for col := 0; col < 3; col++ {
				i := row*3 + col
				name := names[i]
				kind := widget.Secondary
				if w.tab == i {
					kind = widget.Primary
				}
				if widget.Button(c.Key(name), kind, nil, name, "").Width(0).Grow(1).Shrink(1).MinWidth(0).PaddingX(7).Height(28).Clicked() {
					w.tab = i
				}
			}
		})
	}
	widget.Divider(c)
	if w.message != "" {
		ui.Text(c, w.message).FontSize(11.5).TextColor(theme.Done)
	}
	switch w.tab {
	case 0:
		v.libraryPanel(c, o)
	case 1:
		v.reviewPanel(c, o)
	case 2:
		v.checksPanel(c, o)
	case 3:
		v.arrangePanel(c, o)
	case 4:
		v.motionPanel(c, o)
	case 5:
		ui.Text(c, "One offline HTML file").FontSize(13).FontWeight(600)
		ui.Text(c, "Embed images, fonts, CSS and classic scripts. Public web assets are downloaded by this host; private URLs must first be imported as canvas assets. Runtime API calls and JavaScript modules need a project build.").FontSize(12).TextColor(theme.Muted)
		if widget.GhostButton(c, "Export this design offline").FillWidth().Clicked() {
			v.export(o, *v.canvas, v.frameOf().ID, "offline-html")
		}
		if widget.GhostButton(c, "Export all designs offline").FillWidth().Clicked() {
			v.export(o, *v.canvas, "", "offline-html")
		}
	}
}

func (v *View) libraryPanel(c *ui.Context, o Options) {
	w := v.workflow
	ui.Text(c, "Reusable HTML components").FontSize(13).FontWeight(600)
	ui.Text(c, "Select a mockup element, name it, then save. Linked instances follow updates; styles should be inline or use the destination's design tokens.").FontSize(12).TextColor(theme.Muted)
	widget.Input(c, &w.name, "Component name")
	if widget.GhostButton(c, "Save selected component").FillWidth().Disabled(v.sel == "" || o.Library == nil || w.name == "").Clicked() {
		v.libraryCall(o, proto.CanvasLibraryParams{Op: "save", Canvas: v.canvas.ID, Frame: v.frameOf().ID, Target: v.sel, Component: proto.CanvasLibraryComponent{Name: w.name}})
	}
	if w.loading {
		ui.Text(c, "Loading library…").FontSize(12).TextColor(theme.Muted)
	}
	for _, comp := range w.library {
		ui.Column(c.Key(comp.ID)).FillWidth().Gap(5).Padding(8).Background(theme.Sunken).Radius(7).Children(func() {
			ui.Text(c, comp.Name).FontSize(12.5).FontWeight(600)
			ui.Row(c).FillWidth().Gap(5).Children(func() {
				if widget.GhostButton(c, "Insert").Height(27).PaddingX(8).Label("Insert component " + comp.Name).Clicked() {
					v.libraryCall(o, proto.CanvasLibraryParams{Op: "insert", Canvas: v.canvas.ID, Frame: v.frameOf().ID, Component: comp})
				}
				if widget.GhostButton(c, "Update").Height(27).PaddingX(8).Label("Update linked " + comp.Name).Disabled(v.sel == "").Tooltip("Replace this definition from the selected element, updating linked instances on this host").Clicked() {
					v.libraryCall(o, proto.CanvasLibraryParams{Op: "save", Canvas: v.canvas.ID, Frame: v.frameOf().ID, Target: v.sel, Component: comp})
				}
			})
			if widget.GhostButton(c, "Remove from library").Height(25).Label("Remove component " + comp.Name).Clicked() {
				v.libraryCall(o, proto.CanvasLibraryParams{Op: "remove", Component: comp})
			}
		})
	}
}

func (v *View) libraryCall(o Options, p proto.CanvasLibraryParams) {
	if o.Library == nil {
		return
	}
	w := v.workflow
	w.loading = true
	w.message = ""
	o.Library(p, func(result proto.CanvasLibraryResult, err error) {
		w.loading = false
		if err != nil {
			v.err = err.Error()
			return
		}
		w.library = result.Components
		w.message = "Library updated"
		for _, cv := range result.Canvases {
			if cv.ID == v.selected {
				v.show(o, cv)
			}
		}
	})
}

func (v *View) reviewPanel(c *ui.Context, o Options) {
	ui.Text(c, "Review one element at a time").FontSize(13).FontWeight(600)
	ui.Text(c, "Choose Compare on an older version first. Keeping or rejecting one element preserves changes to the rest of the design.").FontSize(12).TextColor(theme.Muted)
	if widget.GhostButton(c, "Choose comparison version").FillWidth().Clicked() {
		v.panel = "versions"
		v.loadHistory(o)
	}
	cmp := v.compare
	if cmp == nil {
		return
	}
	for i, ch := range cmp.changes {
		if ch.Element == "" {
			continue
		}
		ui.Column(c.Key("review-" + strconv.Itoa(i))).FillWidth().Gap(5).Padding(7).Background(theme.Sunken).Radius(6).Children(func() {
			ui.Text(c, ch.Kind+" · "+ch.Element).FontSize(12).SingleLine()
			for _, review := range v.canvas.Reviews {
				if review.Base == cmp.n && review.Head == cmp.head && review.Frame == ch.FrameID && review.Element == strings.TrimPrefix(ch.Element, "#") {
					ui.Text(c, review.Decision).FontSize(11).TextColor(theme.Done)
				}
			}
			ui.Row(c).FillWidth().Gap(6).Children(func() {
				for _, decision := range []string{"keep", "reject"} {
					if widget.GhostButton(c, decision).Label(decision + " " + ch.Element).Height(27).Clicked() {
						frame := ch.FrameID
						if frame == "" {
							frame = ch.Frame
						}
						v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpReview, Review: &proto.CanvasElementReview{Frame: frame, Element: strings.TrimPrefix(ch.Element, "#"), Base: cmp.n, Head: cmp.head, Decision: decision}})
					}
				}
			})
		})
	}
}

func (v *View) motionPanel(c *ui.Context, o Options) {
	w := v.workflow
	f := v.frameOf()
	ui.Text(c, "Prototype transition").FontSize(13).FontWeight(600)
	ui.Text(c, "Select a shape or mockup element. The transition runs in Interact mode and in exported previews.").FontSize(12).TextColor(theme.Muted)
	ui.Text(c, "Source: "+v.sel).FontSize(12).TextColor(theme.Accent)
	if w.transitionSource != v.sel {
		w.transitionSource = v.sel
		w.transition = proto.CanvasTransition{Element: v.sel, Trigger: "click", Effect: "fade", Easing: "ease-out", Duration: 220}
		w.duration = "220"
		for _, t := range f.Transitions {
			if t.Element == v.sel {
				w.transition = t
				w.duration = strconv.Itoa(t.Duration)
			}
		}
	}
	names := []string{}
	for _, fr := range v.canvas.Frames {
		names = append(names, fr.ID+" · "+fr.Name)
	}
	choice := ""
	for _, fr := range v.canvas.Frames {
		if fr.ID == w.transition.Target {
			choice = fr.ID + " · " + fr.Name
		}
	}
	ui.Text(c, "Destination").FontSize(11.5).TextColor(theme.Muted)
	if widget.Choice(c.Key("target"), "Target design", &choice, names) {
		for _, fr := range v.canvas.Frames {
			if choice == fr.ID+" · "+fr.Name {
				w.transition.Target = fr.ID
			}
		}
	}
	widget.Choice(c.Key("trigger"), "Transition trigger", &w.transition.Trigger, []string{"click", "hover"})
	widget.Choice(c.Key("effect"), "Transition effect", &w.transition.Effect, []string{"none", "fade", "slide"})
	widget.Choice(c.Key("easing"), "Transition easing", &w.transition.Easing, []string{"ease", "ease-in", "ease-out", "ease-in-out", "linear"})
	widget.Input(c, &w.duration, "Duration in milliseconds")
	if widget.GhostButton(c, "Apply transition").FillWidth().Disabled(v.sel == "" || w.transition.Target == "").Clicked() {
		n, err := strconv.Atoi(w.duration)
		if err != nil {
			v.err = "Duration must be a number"
		} else {
			t := w.transition
			t.Element = v.sel
			t.Duration = n
			v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpTransition, Transition: &t})
			w.message = "Transition saved"
		}
	}
	if widget.GhostButton(c, "Remove transition").FillWidth().Disabled(v.sel == "").Clicked() {
		v.edit(o, proto.CanvasEdit{Op: proto.CanvasOpTransition, Transition: &proto.CanvasTransition{Element: v.sel}})
	}
}

func (v *View) arrangePanel(c *ui.Context, o Options) {
	ui.Text(c, fmt.Sprintf("%d shapes selected", len(v.selectedShapes()))).FontSize(13).FontWeight(600)
	ui.Text(c, "Shift-click to add shapes; grouped shapes select together. Ctrl/Cmd+A selects all shapes; Ctrl/Cmd+G groups them.").FontSize(12).TextColor(theme.Muted)
	ui.Row(c).FillWidth().Gap(6).Children(func() {
		if widget.GhostButton(c, "Select all shapes").Clicked() {
			v.selectAllShapes()
		}
		if widget.GhostButton(c, "Clear").Clicked() {
			v.deselect(o)
		}
	})
	actions := []string{"left", "center-x", "right", "top", "center-y", "bottom", "distribute-x", "distribute-y", "group", "ungroup", "lock", "unlock"}
	for row := 0; row < len(actions)/2; row++ {
		ui.Row(c.Key("arrange-" + strconv.Itoa(row))).FillWidth().Gap(6).Children(func() {
			for col := 0; col < 2; col++ {
				action := actions[row*2+col]
				if widget.GhostButton(c.Key(action), action).Width(0).Grow(1).Shrink(1).MinWidth(0).Disabled(len(v.selectedShapes()) == 0).Clicked() {
					v.batch(o, action)
				}
			}
		})
	}
}
