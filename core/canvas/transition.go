package canvas

import (
	"fmt"
	"slices"

	"github.com/enowx/enowx/proto"
)

func setTransition(c *proto.Canvas, f *proto.CanvasFrame, e proto.CanvasEdit) error {
	if e.Transition == nil {
		return fmt.Errorf("canvas: transition settings required")
	}
	t := *e.Transition
	if t.Element == "" {
		return fmt.Errorf("canvas: select a source element")
	}
	if t.Target != "" {
		if _, err := FrameIndex(*c, t.Target); err != nil {
			return err
		}
		_, htmlErr := ElementHTML(f.HTML, t.Element)
		if htmlErr != nil && !slices.ContainsFunc(f.Elements, func(el proto.CanvasElement) bool { return el.ID == t.Element }) {
			return fmt.Errorf("canvas: no transition source %s", t.Element)
		}
		if !slices.Contains([]string{"click", "hover"}, t.Trigger) || !slices.Contains([]string{"none", "fade", "slide"}, t.Effect) || !slices.Contains([]string{"ease", "ease-in", "ease-out", "ease-in-out", "linear"}, t.Easing) || t.Duration < 0 || t.Duration > 3000 {
			return fmt.Errorf("canvas: invalid transition settings")
		}
	}
	f.Transitions = slices.DeleteFunc(f.Transitions, func(x proto.CanvasTransition) bool { return x.Element == t.Element })
	if t.Target != "" {
		f.Transitions = append(f.Transitions, t)
	}
	return nil
}
