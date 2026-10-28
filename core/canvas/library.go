package canvas

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/enowx/enowx/proto"
	"golang.org/x/net/html"
)

// Library manages reusable components on this host. Only instances explicitly
// linked with data-enowx-component follow updates; copies remain independent.
func (s *Store) Library(p proto.CanvasLibraryParams) (proto.CanvasLibraryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var result proto.CanvasLibraryResult
	file := filepath.Join(s.dir, "components.json")
	if err := readJSON(file, &result.Components); err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	if result.Components == nil {
		result.Components = []proto.CanvasLibraryComponent{}
	}
	if p.Op == "" || p.Op == "list" {
		return result, nil
	}
	if p.Op == "insert" {
		i := slices.IndexFunc(result.Components, func(x proto.CanvasLibraryComponent) bool { return x.ID == p.Component.ID })
		if i < 0 {
			return result, fmt.Errorf("canvas: component no longer exists")
		}
		cv, err := s.Get(p.Canvas)
		if err != nil {
			return result, err
		}
		fi, err := FrameIndex(cv, p.Frame)
		if err != nil {
			return result, err
		}
		f := &cv.Frames[fi]
		id, src := ComponentHTML(Component{Name: "component-" + result.Components[i].ID, HTML: result.Components[i].HTML}, f.HTML)
		src, err = componentInstance(src, id, result.Components[i].ID)
		if err != nil {
			return result, err
		}
		f.HTML, err = patchHTML(f.HTML, "", p.Into, src)
		if err != nil {
			return result, err
		}
		if err = s.commit(&cv, proto.CanvasVersion{Op: "component.insert", Frame: f.ID, Element: id}); err != nil {
			return result, err
		}
		result.Canvases = append(result.Canvases, cv)
		return result, nil
	}
	i := slices.IndexFunc(result.Components, func(x proto.CanvasLibraryComponent) bool { return x.ID == p.Component.ID })
	if p.Op == "remove" {
		if i < 0 {
			return result, fmt.Errorf("canvas: no component to remove")
		}
		result.Components = slices.Delete(result.Components, i, i+1)
		return result, writeJSON(file, result.Components)
	}
	if p.Op != "save" {
		return result, fmt.Errorf("canvas: unknown library action %q", p.Op)
	}
	comp := p.Component
	comp.Name = strings.TrimSpace(comp.Name)
	if comp.Name == "" || len(comp.Name) > 100 {
		return result, fmt.Errorf("canvas: component name must have 1–100 characters")
	}
	if p.Canvas != "" {
		cv, err := s.Get(p.Canvas)
		if err != nil {
			return result, err
		}
		fi, err := FrameIndex(cv, p.Frame)
		if err != nil {
			return result, err
		}
		comp.HTML, err = ElementHTML(cv.Frames[fi].HTML, p.Target)
		if err != nil {
			return result, err
		}
		comp.HTML = inlineAssets(comp.HTML, func(name string) ([]byte, error) { return s.Asset(p.Canvas, name) })
	}
	if strings.TrimSpace(comp.HTML) == "" || len(comp.HTML) > 2<<20 {
		return result, fmt.Errorf("canvas: component must contain HTML up to 2 MB")
	}
	src, err := componentInstance(comp.HTML, "ID", "")
	if err != nil {
		return result, err
	}
	comp.HTML = src
	if comp.ID == "" {
		comp.ID = newID(6)
	} else if i < 0 {
		return result, fmt.Errorf("canvas: no component to update")
	}
	comp.Updated = time.Now().UTC()
	if i >= 0 {
		result.Components[i] = comp
	} else {
		result.Components = append(result.Components, comp)
	}
	encoded, err := json.Marshal(result.Components)
	if err != nil || len(encoded) > 8<<20 {
		return result, fmt.Errorf("canvas: component library exceeds the 8 MB limit")
	}
	if err := writeJSON(file, result.Components); err != nil {
		return result, err
	}
	if i >= 0 {
		list, err := s.List()
		if err != nil {
			return result, err
		}
		for _, info := range list {
			cv, err := s.Get(info.ID)
			if err != nil {
				return result, err
			}
			changed := false
			for fi := range cv.Frames {
				f := &cv.Frames[fi]
				doc, err := parseDoc(f.HTML)
				if err != nil {
					return result, err
				}
				var ids []string
				var walk func(*html.Node)
				walk = func(n *html.Node) {
					if attr(n, "data-enowx-component") == comp.ID {
						ids = append(ids, attr(n, "id"))
						return
					}
					for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
						walk(ch)
					}
				}
				walk(doc.root)
				for _, id := range ids {
					src, err := componentInstance(comp.HTML, id, comp.ID)
					if err != nil {
						return result, err
					}
					f.HTML, err = patchHTML(f.HTML, id, "", src)
					if err != nil {
						return result, err
					}
					changed = true
				}
			}
			if changed {
				if err := s.commit(&cv, proto.CanvasVersion{Op: "component.update"}); err != nil {
					return result, err
				}
				result.Canvases = append(result.Canvases, cv)
			}
		}
	}
	return result, nil
}

func componentInstance(src, id, component string) (string, error) {
	d, err := parseDoc(src)
	if err != nil {
		return "", err
	}
	var root *html.Node
	for n := d.body().FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode {
			if root != nil {
				return "", fmt.Errorf("canvas: a component needs one root element")
			}
			root = n
		}
	}
	if root == nil {
		return "", fmt.Errorf("canvas: component has no root element")
	}
	ids := map[string]string{}
	rootID := attr(root, "id")
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if old := attr(n, "id"); old != "" {
			ids[old] = id + "-" + strings.TrimPrefix(old, rootID+"-")
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(root)
	if old := attr(root, "id"); old != "" {
		ids[old] = id
	}
	walk = func(n *html.Node) {
		var attrs []html.Attribute
		for _, a := range n.Attr {
			if a.Key == "data-enowx-component" {
				continue
			}
			if a.Key == "id" {
				a.Val = ids[a.Val]
			}
			if a.Key == "for" || a.Key == "aria-labelledby" || a.Key == "aria-describedby" {
				parts := strings.Fields(a.Val)
				for i, v := range parts {
					if ids[v] != "" {
						parts[i] = ids[v]
					}
				}
				a.Val = strings.Join(parts, " ")
			}
			if a.Key == "href" && strings.HasPrefix(a.Val, "#") {
				if ref := ids[strings.TrimPrefix(a.Val, "#")]; ref != "" {
					a.Val = "#" + ref
				}
			}
			attrs = append(attrs, a)
		}
		n.Attr = attrs
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			walk(ch)
		}
	}
	walk(root)
	if attr(root, "id") == "" {
		root.Attr = append(root.Attr, html.Attribute{Key: "id", Val: id})
	}
	if component != "" {
		root.Attr = append(root.Attr, html.Attribute{Key: "data-enowx-component", Val: component})
	}
	var b bytes.Buffer
	if err := html.Render(&b, root); err != nil {
		return "", err
	}
	return b.String(), nil
}
