package canvasview

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"github.com/enowx/enowx/core/canvas"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/icon"
	"github.com/enowx/enowx/ui/theme"
)

// The stage shows the frame open, scaled to fit: its mockup in a web view,
// its shapes drawn over it, each one an element the user selects, moves,
// resizes and edits in Design mode.

// dragState is a shape being moved or resized: where it shows meanwhile.
type dragState struct {
	id         string
	resize     bool
	x, y, w, h float64
	moved      bool
}

// textEdit is a shape's text being edited in place.
type textEdit struct {
	id   string
	text string
}

// highlightFor is how long an agent's change stays outlined.
const highlightFor = 1600 * time.Millisecond

func color(s string, fallback ui.Color) ui.Color {
	if c, ok := canvas.ParseColor(s); ok {
		return ui.RGB(c.R, c.G, c.B)
	}
	return fallback
}

// stage draws the frame open: on the endless board in Design mode, as a
// page to use in Interact mode.
func (v *View) stage(c *ui.Context, o Options) {
	f := v.frameOf()
	if v.chosenFrame != f.ID {
		v.chosen, v.chosenFrame = nil, f.ID
	}
	width, height := float64(f.Width), float64(f.Height)
	hasHTML := strings.TrimSpace(f.HTML) != ""
	if v.preview > 0 && hasHTML {
		height = math.Max(height*float64(v.preview)/width, height)
		width = float64(v.preview)
	}
	if o.Web != nil {
		o.Web.setDesign(v.mode == modeDesign)
		if v.canvas != nil {
			shown := f
			if v.drag != nil {
				shown.Elements = make([]proto.CanvasElement, len(f.Elements))
				for i, el := range f.Elements {
					shown.Elements[i] = v.placed(el)
				}
			}
			o.Web.show(*v.canvas, shown, nil)
		}
	}
	if v.mode == modeInteract && hasHTML && o.Web != nil {
		v.interactStage(c, o, float32(width))
		return
	}
	// The zoom's controls sit beside the board, over it, so its pointer
	// handling never takes their clicks.
	opacity, offset := v.prototypeAppearance(c)
	wrap := ui.Box(c.Key("stage-wrap")).Grow(1).FillWidth().MinHeight(0).Opacity(opacity).Translate(offset, 0)
	var area ui.Element
	wrap.Children(func() { area = ui.Box(c.Key("stage")).FillWidth().FillHeight().Background(theme.Sunken).Clip() })
	b := area.Bounds()
	if key := fmt.Sprintf("%s/%v", f.ID, width); v.cam.key != key {
		if b.W <= 0 {
			// Placed once the stage has a size: the next frame.
			c.After(time.Millisecond)
			return
		}
		v.cam.fitWidth(b.W, float32(width))
		v.cam.key = key
	}
	v.scale = v.cam.z
	s := float64(v.cam.z)
	design := ui.Rect{X: v.cam.x, Y: v.cam.y, W: float32(width * s), H: float32(height * s)}
	// A press on the page, or on a shape, is theirs; elsewhere the board's.
	onDesign := func(x, y float32) bool {
		if hasHTML || v.panel == "grid" && v.gridF != nil && v.gridF.drawing {
			return x >= design.X && y >= design.Y && x < design.X+design.W && y < design.Y+design.H
		}
		px, py := float64((x-design.X)/v.cam.z), float64((y-design.Y)/v.cam.z)
		for _, el := range f.Elements {
			el = v.placed(el)
			if el.Type == proto.ShapeLine || el.Type == proto.ShapeArrow {
				continue
			}
			if px >= el.X-6 && py >= el.Y-6 && px < el.X+el.W+6 && py < el.Y+el.H+6 {
				return true
			}
		}
		return false
	}
	area.HandleInput(v.cameraInput(0, 0, onDesign, func(float32, float32) {
		if v.mode == modeDesign {
			v.deselect(o)
		}
	}))
	if v.cam.panning {
		area.Cursor(ui.CursorGrabbing)
	}
	now := time.Now()
	area.Children(func() {
		v.dots(c)
		// A mockup's page shows as a sheet; shapes alone lie on the board
		// itself, free form, with nothing drawn under them.
		board := ui.Box(c.Key("board:"+f.ID)).Absolute().Left(design.X).Top(design.Y).Size(design.W, design.H).Label("Frame " + f.Name)
		if hasHTML {
			board.Background(color(f.Background, ui.RGB(255, 255, 255))).Clip().Shadow(0, 2, 12, 0, theme.Fade(theme.TextStrong, 0.12))
		}
		if v.mode == modeDesign && board.Clicked() {
			v.deselect(o)
		}
		board.Children(func() {
			if hasHTML {
				v.mockup(c, o, f, design, s)
			}
			drawing := v.panel == "grid" && v.gridF != nil && v.gridF.drawing
			if !drawing {
				v.gridGuides(c, o, f, s)
			}
			v.arrows(c, f, s)
			for _, el := range f.Elements {
				if el.Type == proto.ShapeLine || el.Type == proto.ShapeArrow {
					v.arrowHandle(c.Key(el.ID), o, f, el, s)
					continue
				}
				v.shape(c.Key(el.ID), o, el, s, now)
			}
			if drawing {
				v.gridGuides(c, o, f, s)
			}
		})
	})
	wrap.Children(func() {
		v.zoomBar(c, b.W, b.H, float32(width), float32(height))
		if hasHTML && o.Web != nil {
			o.Web.measureWidth(c)
			if n := o.Web.Wider(); n > 0 {
				ui.Row(c.Key("wider")).Absolute().Left(12).Top(12).Height(28).PaddingX(10).Gap(6).AlignItems(ui.Center).
					Radius(8).Background(theme.Fade(theme.Warning, 0.15)).Border(1, theme.Fade(theme.Warning, 0.4)).
					Tooltip("Something sticks out past the design's width, so its page scrolls sideways: fix it, or ask the agent to").
					Children(func() {
						ui.Icon(c, icon.Alert).FontSize(13).TextColor(theme.Warning)
						ui.Text(c, fmt.Sprintf("Wider than its frame by %d px: it scrolls sideways", n)).FontSize(12).TextColor(theme.Warning).SingleLine()
					})
			}
		}
	})
	if v.lastAt.Add(highlightFor).After(now) {
		c.After(highlightFor - now.Sub(v.lastAt) + 10*time.Millisecond)
	}
}

// deselect selects nothing.
func (v *View) deselect(o Options) {
	v.chosen = nil
	v.sel, v.connect = "", ""
	if o.Web != nil {
		o.Web.outline("")
	}
}

// interactStage shows the mockup as a page to use: as wide as the stage
// lets it (its own width at most), scrolled as a browser scrolls.
func (v *View) interactStage(c *ui.Context, o Options, width float32) {
	area := ui.Box(c.Key("stage")).Grow(1).FillWidth().MinHeight(0).Background(theme.Sunken).Clip()
	b := area.Bounds()
	z := float32(1)
	if b.W > 0 {
		z = min(max((b.W-48)/width, minZoom), 1)
	}
	v.scale = z
	area.Children(func() {
		if b.W > 0 {
			o.Web.view(c.Key("web"), width*z, b.H, float64(z)).Absolute().Left((b.W - width*z) / 2).Top(0)
		}
	})
	o.Web.pollGoto(c, v.followLink)
}

// mockup is the frame's HTML: in a web view, or, where there is none,
// its elements listed as they come. In Design mode a cover over the page
// takes its pointer, so the board moves and zooms over it too, and a click
// on it selects nothing.
func (v *View) mockup(c *ui.Context, o Options, f proto.CanvasFrame, design ui.Rect, s float64) {
	w, h := design.W, design.H
	if o.Web != nil {
		// Small, the design shows as its picture, rendered at its own
		// width: a web view zoomed out that far lays its page out anew,
		// narrower, rather than shrinking it. The live page shows from
		// liveZoom up, and until the picture comes.
		var img ui.ImageSource
		if s < liveZoom && v.preview == 0 {
			img = v.picture(c, o, f)
		}
		if img != nil {
			ui.Image(c.Key("design-picture"), img).Absolute().Left(0).Top(0).Size(w, h).Label("Picture of " + f.Name)
		} else {
			o.Web.view(c.Key("web"), w, h, s).Absolute().Left(0).Top(0)
		}
		// A click picks the component under it, for the user to ask about,
		// change or note: on the live page.
		cover := ui.Box(c.Key("web-cover")).Absolute().Left(0).Top(0).Size(w, h).Label("Design " + f.Name).Cursor(ui.CursorPointer).
			HandleInput(v.cameraInput(design.X, design.Y, nil, func(x, y float32) {
				if img != nil {
					// Too small to pick from: zoom in there, to the page.
					v.cam.zoomAt(design.X+x, design.Y+y, liveZoom/float32(s))
					return
				}
				v.pickAt(o, x/float32(s), y/float32(s))
			}))
		cover.Hovered() // takes the pointer, over the page
		if img != nil {
			cover.Tooltip("Click to zoom in and pick its components")
		}
		v.componentBar(c, o, s, w, h)
		return
	}
	ui.Column(c.Key("outline")).Absolute().Left(0).Top(0).Size(w, h).Padding(16).Gap(4).Children(func() {
		ui.Text(c, "This device shows mockups in a web view: here, their elements.").FontSize(11).TextColor(ui.RGB(100, 116, 139))
		for _, el := range canvas.Outline(f.HTML) {
			row := ui.Row(c.Key(el.ID)).Padding(2, 6, 2, 6+float32(el.Depth)*12).Radius(4).Gap(6)
			if v.recent(el.ID) {
				row.Background(ui.RGBA(99, 102, 241, 0.18))
			}
			row.Children(func() {
				ui.Text(c, "#"+el.ID).FontSize(11).TextColor(ui.RGB(79, 70, 229))
				ui.Text(c, el.Text).FontSize(11).TextColor(ui.RGB(15, 23, 42)).SingleLine().Shrink(1)
			})
		}
	})
}

// followLink shows the design a mockup's link goes to: its number
// ("design-2"), name or ID.
func (v *View) followLink(target string) {
	if v.canvas == nil {
		return
	}
	if i, err := canvas.FrameIndex(*v.canvas, strings.TrimPrefix(target, "#")); err == nil {
		v.frame, v.sel = v.canvas.Frames[i].ID, ""
	}
}
