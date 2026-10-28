(function () {
  window.__enowxCheckStart = async function () {
    window.__enowxCheckResult = null;
    const issues = [], source = document.documentElement.cloneNode(true);
    source.classList.remove('enowx-design');
    source.querySelectorAll('script,iframe,[data-enowx-frame-base]').forEach(n => n.remove());
    const label = n => n.id || (n.closest('[id]') || {}).id || n.tagName.toLowerCase();
    const add = (element, message, width) => {
      if (issues.length < 150 && !issues.some(i => i.element === element && i.message === message && i.width === width)) issues.push({element, message, width});
    };
    const rgb = s => {const m = s.match(/[\d.]+/g); return m && m.length >= 3 ? m.map(Number) : null;};
    const lum = c => c.slice(0, 3).reduce((sum, v, i) => {v /= 255; return sum + [.2126,.7152,.0722][i] * (v <= .04045 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4);}, 0);
    try {
      for (const width of [1440, 768, 390]) {
        const frame = document.createElement('iframe');
        frame.setAttribute('sandbox', 'allow-same-origin');
        frame.style.cssText = `position:fixed;left:-20000px;top:0;width:${width}px;height:900px;border:0;visibility:hidden`;
        const loaded = new Promise(resolve => {frame.onload = resolve; setTimeout(resolve, 3000);});
        frame.srcdoc = '<!doctype html>' + source.outerHTML;
        document.body.append(frame);
        try {
          await loaded;
          const doc = frame.contentDocument, win = frame.contentWindow;
          if (!doc || !win) {add('frame', 'Could not inspect this preview', width); continue;}
          if (doc.fonts) await Promise.race([doc.fonts.ready, new Promise(resolve => setTimeout(resolve, 1500))]);
          const controls = [];
          for (const node of Array.from(doc.body.querySelectorAll('*')).slice(0, 2000)) {
            const r = node.getBoundingClientRect(), st = win.getComputedStyle(node);
            if (!r.width || !r.height || st.visibility === 'hidden' || st.display === 'none' || Number(st.opacity) === 0) continue;
            const id = label(node);
            if (r.left < -2 || r.right > width + 2) add(id, 'Extends beyond the preview width', width);
            if (node.childElementCount === 0 && node.textContent.trim()) {
              if ((node.scrollWidth > node.clientWidth + 2 && ['hidden','clip'].includes(st.overflowX)) || (node.scrollHeight > node.clientHeight + 2 && ['hidden','clip'].includes(st.overflowY))) add(id, 'Text is clipped', width);
              let bg = node, color = null, unknown = false;
              while (bg) {const b = win.getComputedStyle(bg); if (b.backgroundImage !== 'none') {unknown = true; break;} const c = rgb(b.backgroundColor); if (c && (c.length < 4 || c[3] === 1)) {color = c; break;} bg = bg.parentElement;}
              const fg = rgb(st.color);
              if (!unknown && fg && (fg.length < 4 || fg[3] === 1)) {
                const a = lum(fg), b = lum(color || [255,255,255]);
                const ratio = (Math.max(a,b) + .05) / (Math.min(a,b) + .05);
                const large = parseFloat(st.fontSize) >= 24 || (parseFloat(st.fontSize) >= 18.66 && Number(st.fontWeight) >= 700);
                if (ratio < (large ? 3 : 4.5)) add(id, `Text contrast ${ratio.toFixed(1)}:1`, width);
              }
            }
            if (node.matches('button,input,select,textarea,a[href],[role=button]')) {
              if (r.width < 24 || r.height < 24) add(id, 'Interactive target is smaller than 24 px', width);
              controls.push({node,r,id});
            }
          }
          for (let i = 0; i < controls.length; i++) for (let j = i + 1; j < controls.length; j++) {
            const a = controls[i], b = controls[j];
            if (a.node.contains(b.node) || b.node.contains(a.node)) continue;
            const x = Math.min(a.r.right,b.r.right) - Math.max(a.r.left,b.r.left), y = Math.min(a.r.bottom,b.r.bottom) - Math.max(a.r.top,b.r.top);
            if (x > 4 && y > 4) add(a.id, 'Interactive control overlaps ' + b.id, width);
          }
        } finally {frame.remove();}
      }
      window.__enowxCheckResult = {issues};
    } catch (error) {window.__enowxCheckResult = {error: String(error)};}
  };
})();
