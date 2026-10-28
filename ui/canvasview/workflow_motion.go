package canvasview

import (
	"time"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
)

type prototypeRun struct {
	rule     proto.CanvasTransition
	at       time.Time
	entering bool
}

func (v *View) transitionPointer(c *ui.Context, o Options, box ui.Element, id string) {
	was := ui.Local[bool](box, "prototype-hover", func() bool { return false })
	over := box.Hovered()
	entered := over && !*was
	*was = over
	for _, rule := range v.frameOf().Transitions {
		if rule.Element == id && ((rule.Trigger == "click" && box.Clicked()) || (rule.Trigger == "hover" && entered)) {
			if v.prototype != nil {
				return
			}
			if rule.Effect == "none" || rule.Duration == 0 {
				v.followLink(rule.Target)
				c.After(0)
			} else {
				v.prototype = &prototypeRun{rule: rule, at: c.Now()}
				c.After(time.Millisecond)
			}
			return
		}
	}
}

func (v *View) prototypeAppearance(c *ui.Context) (opacity, x float32) {
	p := v.prototype
	if p == nil {
		return 1, 0
	}
	d := time.Duration(p.rule.Duration) * time.Millisecond
	f := float32(c.Now().Sub(p.at)) / float32(d)
	if f >= 1 {
		if !p.entering {
			v.followLink(p.rule.Target)
			p.entering = true
			p.at = c.Now()
			c.After(time.Millisecond)
			return 0, 0
		}
		v.prototype = nil
		return 1, 0
	}
	c.After(16 * time.Millisecond)
	f = max(0, f)
	switch p.rule.Easing {
	case "ease-in":
		f = f * f
	case "ease-out":
		f = 1 - (1-f)*(1-f)
	case "ease", "ease-in-out":
		f = f * f * (3 - 2*f)
	}
	if p.rule.Effect == "slide" {
		x = -18 * f
		if p.entering {
			x = 18 * (1 - f)
		}
	}
	if p.entering {
		return f, x
	}
	return 1 - f, x
}
