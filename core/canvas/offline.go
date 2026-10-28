package canvas

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

// FetchAsset fetches a public resource without credentials.
type FetchAsset func(rawURL string) (data []byte, mediaType string, err error)

// OfflineDocument embeds external styles, scripts, fonts and images. It fails
// instead of quietly exporting a page that still depends on remote resources.
func OfflineDocument(page string, fetch FetchAsset) (string, error) {
	b := offlineBundler{fetch: fetch, cache: map[string]string{}}
	return b.document(page, 0)
}

func (b *offlineBundler) document(page string, depth int) (string, error) {
	if depth > 8 {
		return "", fmt.Errorf("canvas: nested documents exceed eight levels")
	}
	d, err := parseDoc(page)
	if err != nil {
		return "", err
	}
	var walk func(*html.Node) error
	walk = func(n *html.Node) error {
		if n.Type == html.ElementNode {
			if n.Data == "iframe" && attr(n, "srcdoc") != "" {
				src, err := b.document(attr(n, "srcdoc"), depth+1)
				if err != nil {
					return err
				}
				for i := range n.Attr {
					if n.Attr[i].Key == "srcdoc" {
						n.Attr[i].Val = src
					}
				}
			}
			if n.Data == "base" {
				return fmt.Errorf("canvas: remove the base URL before offline export")
			}
			if n.Data == "script" && attr(n, "type") == "module" {
				return fmt.Errorf("canvas: bundle JavaScript modules in the project before exporting offline")
			}
			if n.Data == "link" && strings.EqualFold(attr(n, "rel"), "stylesheet") {
				u := attr(n, "href")
				if u != "" {
					css, err := b.stylesheet(u, 0)
					if err != nil {
						return err
					}
					n.Data = "style"
					n.DataAtom = 0
					n.Attr = nil
					for n.FirstChild != nil {
						n.RemoveChild(n.FirstChild)
					}
					n.AppendChild(&html.Node{Type: html.TextNode, Data: css})
				}
			}
			if n.Data == "script" && attr(n, "src") != "" {
				data, _, err := b.load(attr(n, "src"))
				if err != nil {
					return err
				}
				if strings.Contains(strings.ToLower(string(data)), "</script") {
					return fmt.Errorf("canvas: script contains an unsafe closing tag")
				}
				n.Attr = filterAttributes(n.Attr, "src", "integrity", "crossorigin")
				n.AppendChild(&html.Node{Type: html.TextNode, Data: string(data)})
			}
			for i := range n.Attr {
				a := &n.Attr[i]
				if a.Key == "srcset" {
					return fmt.Errorf("canvas: use one image source instead of srcset for offline export")
				}
				if a.Key == "src" || a.Key == "poster" || (n.Data == "link" && a.Key == "href") {
					if externalResource(a.Val) {
						v, err := b.resource(a.Val)
						if err != nil {
							return err
						}
						a.Val = v
					}
				}
				if a.Key == "style" {
					v, err := b.css(a.Val, "", 0)
					if err != nil {
						return err
					}
					a.Val = v
				}
			}
			if n.Data == "style" {
				for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
					if ch.Type == html.TextNode {
						css, err := b.css(ch.Data, "", 0)
						if err != nil {
							return err
						}
						ch.Data = css
					}
				}
			}
		}
		for ch := n.FirstChild; ch != nil; ch = ch.NextSibling {
			if err := walk(ch); err != nil {
				return err
			}
		}
		return nil
	}
	if err := walk(d.root); err != nil {
		return "", err
	}
	if head := find(d.root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == "head" }); head != nil {
		policy := &html.Node{Type: html.ElementNode, Data: "meta", Attr: []html.Attribute{{Key: "http-equiv", Val: "Content-Security-Policy"}, {Key: "content", Val: "default-src data: blob:; script-src 'unsafe-inline' 'unsafe-eval' data: blob:; style-src 'unsafe-inline' data:; font-src data:; img-src data: blob:; frame-src 'self' data: blob:; connect-src 'none'"}}}
		head.InsertBefore(policy, head.FirstChild)
	}
	result := d.render()
	if len(result) > 8<<20 {
		return "", fmt.Errorf("canvas: offline HTML exceeds the 8 MB protocol-safe limit")
	}
	return result, nil
}

type offlineBundler struct {
	fetch        FetchAsset
	cache        map[string]string
	total, count int
}

func externalResource(s string) bool {
	return s != "" && !strings.HasPrefix(s, "data:") && !strings.HasPrefix(s, "#") && s != "about:blank"
}

func (b *offlineBundler) load(raw string) ([]byte, string, error) {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil {
		return nil, "", fmt.Errorf("canvas: offline resources need public HTTP(S) URLs or imported assets")
	}
	if b.count >= 96 {
		return nil, "", fmt.Errorf("canvas: offline export exceeds 96 resources")
	}
	b.count++
	data, typ, err := b.fetch(raw)
	if err != nil {
		return nil, "", fmt.Errorf("canvas: cannot bundle resource from %s: %w", u.Hostname(), err)
	}
	b.total += len(data)
	if len(data) > 4<<20 || b.total > 6<<20 {
		return nil, "", fmt.Errorf("canvas: offline resource exceeds the 4 MB or 6 MB total asset limit")
	}
	return data, typ, nil
}

func (b *offlineBundler) resource(raw string) (string, error) {
	if !externalResource(raw) {
		return raw, nil
	}
	if v, ok := b.cache[raw]; ok {
		return v, nil
	}
	data, typ, err := b.load(raw)
	if err != nil {
		return "", err
	}
	if typ == "" {
		typ = "application/octet-stream"
	}
	if t, _, err := mime.ParseMediaType(typ); err == nil {
		typ = t
	}
	v := "data:" + typ + ";base64," + base64.StdEncoding.EncodeToString(data)
	b.cache[raw] = v
	return v, nil
}

var cssURL = regexp.MustCompile(`(?i)url\(\s*["']?([^\s"')]+)["']?\s*\)`)
var cssImport = regexp.MustCompile(`(?i)@import\s+(?:url\(\s*["']?([^\s"')]+)["']?\s*\)|["']([^"']+)["'])\s*([^;]*);`)

func resolveResource(base, ref string) string {
	u, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	if base != "" {
		b, err := url.Parse(base)
		if err == nil {
			return b.ResolveReference(u).String()
		}
	}
	return ref
}
func (b *offlineBundler) stylesheet(raw string, depth int) (string, error) {
	if depth > 5 {
		return "", fmt.Errorf("canvas: nested stylesheets exceed five levels")
	}
	data, _, err := b.load(raw)
	if err != nil {
		return "", err
	}
	return b.css(string(data), raw, depth+1)
}
func (b *offlineBundler) css(source, base string, depth int) (string, error) {
	var failed error
	source = cssImport.ReplaceAllStringFunc(source, func(s string) string {
		if failed != nil {
			return s
		}
		m := cssImport.FindStringSubmatch(s)
		ref := m[1]
		if ref == "" {
			ref = m[2]
		}
		if strings.Contains(m[3], "layer") || strings.Contains(m[3], "supports") {
			failed = fmt.Errorf("canvas: bundle layered or conditional CSS imports in the project first")
			return s
		}
		css, err := b.stylesheet(resolveResource(base, ref), depth)
		if err != nil {
			failed = err
			return s
		}
		if strings.TrimSpace(m[3]) != "" {
			css = "@media " + m[3] + "{" + css + "}"
		}
		return css
	})
	if failed != nil {
		return "", failed
	}
	source = cssURL.ReplaceAllStringFunc(source, func(s string) string {
		if failed != nil {
			return s
		}
		ref := cssURL.FindStringSubmatch(s)[1]
		if !externalResource(ref) {
			return s
		}
		v, err := b.resource(resolveResource(base, ref))
		if err != nil {
			failed = err
			return s
		}
		return `url("` + v + `")`
	})
	return source, failed
}

func filterAttributes(attrs []html.Attribute, remove ...string) []html.Attribute {
	var result []html.Attribute
	for _, a := range attrs {
		found := false
		for _, key := range remove {
			if key == a.Key {
				found = true
			}
		}
		if !found {
			result = append(result, a)
		}
	}
	return result
}

// ReadAsset limits one response before it is retained in memory.
func ReadAsset(r io.Reader) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(r, (4<<20)+1))
	if err == nil && len(data) > 4<<20 {
		err = fmt.Errorf("resource exceeds 4 MB")
	}
	return data, err
}
