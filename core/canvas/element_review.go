package canvas

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/enowx/enowx/proto"
)

func (s *Store) reviewElement(c *proto.Canvas, e proto.CanvasEdit) (string, string, error) {
	if e.Review == nil {
		return "", "", fmt.Errorf("canvas: element.review needs a review")
	}
	r := *e.Review
	if r.Element == "" || r.Base <= 0 || r.Head < r.Base || r.Head > c.Version {
		return "", "", fmt.Errorf("canvas: invalid review versions or element")
	}
	if r.Decision != "keep" && r.Decision != "reject" {
		return "", "", fmt.Errorf("canvas: review decision must be keep or reject")
	}
	base, err := s.Version(c.ID, r.Base)
	if err != nil {
		return "", "", err
	}
	head, err := s.Version(c.ID, r.Head)
	if err != nil {
		return "", "", err
	}
	ci, err := FrameIndex(*c, r.Frame)
	if err != nil {
		return "", "", err
	}
	bi, err := FrameIndex(base, r.Frame)
	if err != nil {
		return "", "", err
	}
	hi, err := FrameIndex(head, r.Frame)
	if err != nil {
		return "", "", err
	}
	f := &c.Frames[ci]
	shape := func(f proto.CanvasFrame) *proto.CanvasElement {
		for _, el := range f.Elements {
			if el.ID == r.Element {
				x := el
				return &x
			}
		}
		return nil
	}
	before, after, current := shape(base.Frames[bi]), shape(head.Frames[hi]), shape(*f)
	if before != nil || after != nil {
		if !reflect.DeepEqual(after, current) {
			return "", "", fmt.Errorf("canvas: element changed since review; compare again")
		}
		if current != nil && current.Locked {
			return "", "", fmt.Errorf("canvas: element is locked")
		}
		if r.Decision == "reject" {
			index := slices.IndexFunc(f.Elements, func(el proto.CanvasElement) bool { return el.ID == r.Element })
			if before == nil {
				for _, el := range f.Elements {
					if el.From == r.Element || el.To == r.Element {
						return "", "", fmt.Errorf("canvas: review connected arrows first")
					}
				}
				f.Elements = slices.Delete(f.Elements, index, index+1)
			} else if index >= 0 {
				f.Elements[index] = *before
			} else {
				at := len(f.Elements)
				original := base.Frames[bi].Elements
				found := false
				for _, el := range original {
					if el.ID == r.Element {
						found = true
						continue
					}
					if found {
						if next := slices.IndexFunc(f.Elements, func(x proto.CanvasElement) bool { return x.ID == el.ID }); next >= 0 {
							at = next
							break
						}
					}
				}
				f.Elements = slices.Insert(f.Elements, at, *before)
			}
		}
	} else {
		oldHTML, _ := ElementHTML(base.Frames[bi].HTML, r.Element)
		newHTML, _ := ElementHTML(head.Frames[hi].HTML, r.Element)
		curHTML, _ := ElementHTML(f.HTML, r.Element)
		if oldHTML == "" && newHTML == "" {
			return "", "", fmt.Errorf("canvas: no reviewed element %s", r.Element)
		}
		if curHTML != newHTML {
			return "", "", fmt.Errorf("canvas: element changed since review; compare again")
		}
		if r.Decision == "reject" {
			if oldHTML == "" {
				f.HTML, err = removeHTML(f.HTML, r.Element)
			} else if newHTML != "" {
				f.HTML, err = patchHTML(f.HTML, r.Element, "", oldHTML)
			} else {
				f.HTML, err = restoreHTMLPosition(f.HTML, base.Frames[bi].HTML, r.Element)
			}
			if err != nil {
				return "", "", err
			}
		}
	}
	r.Frame = f.ID
	c.Reviews = append(c.Reviews, r)
	return f.ID, r.Element, nil
}

func restoreHTMLPosition(page, old, id string) (string, error) {
	d, err := parseDoc(page)
	if err != nil {
		return "", err
	}
	b, err := parseDoc(old)
	if err != nil {
		return "", err
	}
	n := b.byID(id)
	if n == nil {
		return "", fmt.Errorf("canvas: missing original element")
	}
	parent := d.body()
	if parentID := attr(n.Parent, "id"); parentID != "" {
		parent = d.byID(parentID)
		if parent == nil {
			return "", fmt.Errorf("canvas: original parent changed; restore it first")
		}
	}
	source, err := ElementHTML(old, id)
	if err != nil {
		return "", err
	}
	nodes, err := fragment(source, parent)
	if err != nil {
		return "", err
	}
	var nextID string
	for next := n.NextSibling; next != nil; next = next.NextSibling {
		if nextID = attr(next, "id"); nextID != "" {
			break
		}
	}
	for _, node := range nodes {
		if next := d.byID(nextID); next != nil && next.Parent == parent {
			parent.InsertBefore(node, next)
		} else {
			parent.AppendChild(node)
		}
	}
	return d.render(), nil
}
