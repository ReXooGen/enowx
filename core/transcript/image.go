package transcript

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"strings"

	"github.com/enowx/enowx/proto"
	xdraw "golang.org/x/image/draw"
)

// Thumbnails bound transcript payloads and never download private image URLs.
func transcriptThumbnail(encoded string) proto.TranscriptImage {
	out := proto.TranscriptImage{Name: "Image preview unavailable"}
	if strings.HasPrefix(encoded, "data:") {
		_, encoded, _ = strings.Cut(encoded, ",")
	}
	if len(encoded) > 8<<20 {
		return out
	}
	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return out
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width < 1 || config.Height < 1 || int64(config.Width)*int64(config.Height) > 32<<20 {
		return out
	}
	src, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return out
	}
	w, h := config.Width, config.Height
	if w > 160 || h > 160 {
		if w >= h {
			h = max(1, h*160/w)
			w = 160
		} else {
			w = max(1, w*160/h)
			h = 160
		}
	}
	thumb := image.NewRGBA(image.Rect(0, 0, w, h))
	draw.Draw(thumb, thumb.Bounds(), image.NewUniform(color.White), image.Point{}, draw.Src)
	xdraw.ApproxBiLinear.Scale(thumb, thumb.Bounds(), src, src.Bounds(), draw.Over, nil)
	var buf bytes.Buffer
	if err = jpeg.Encode(&buf, thumb, &jpeg.Options{Quality: 60}); err != nil || buf.Len() > 8<<10 {
		return out
	}
	out.Name, out.Thumbnail = "Image", buf.Bytes()
	return out
}
