package canvas

import (
	"encoding/json"

	"github.com/enowx/enowx/proto"
)

// PrototypeScript gives persisted element transitions the same behavior in a
// mockup's web view and an exported HTML preview.
func PrototypeScript(f proto.CanvasFrame) string {
	if len(f.Transitions) == 0 {
		return ""
	}
	data, _ := json.Marshal(f.Transitions)
	return `<script>(function(){var rules=` + string(data) + `;function run(e,trigger){var n=e.target;while(n&&n!==document){var t=rules.find(function(x){return x.element===n.id&&x.trigger===trigger});if(t){if(trigger==='hover'&&e.relatedTarget&&n.contains(e.relatedTarget))return;e.preventDefault();e.stopImmediatePropagation();var duration=matchMedia('(prefers-reduced-motion: reduce)').matches?0:t.duration;var frames=t.effect==='slide'?[{opacity:1,transform:'translateX(0)'},{opacity:0,transform:'translateX(-18px)'}]:[{opacity:1},{opacity:0}];var done=function(){if(window.__enowxNavigate)window.__enowxNavigate(t.target,t);else window.__enowxGoto=t.target;document.body.style.opacity='1'};if(t.effect==='none'||!duration){done()}else{document.body.animate(frames,{duration:duration,easing:t.easing}).finished.then(done)}return}n=n.parentElement}}document.addEventListener('click',function(e){run(e,'click')},true);document.addEventListener('mouseover',function(e){run(e,'hover')},true)})();</script>`
}
