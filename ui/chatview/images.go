package chatview

import (
	"bytes"

	"github.com/egoist/mygo/ui"
	"github.com/enowx/enowx/proto"
	"github.com/enowx/enowx/ui/theme"
)

type transcriptPreview struct {
	data   []byte
	bitmap *ui.Bitmap
}

func transcriptImages(c *ui.Context, images []proto.TranscriptImage) {
	if len(images) == 0 {
		return
	}
	ui.Column(c.Key("transcript-images")).FillWidth().Gap(6).Children(func() {
		for i, img := range images {
			box := ui.Column(c.Key(i)).Gap(4)
			cache := ui.Local[transcriptPreview](box, "preview", func() transcriptPreview { return transcriptPreview{} })
			if !bytes.Equal(cache.data, img.Thumbnail) {
				cache.data = img.Thumbnail
				cache.bitmap, _ = ui.DecodeBitmap(img.Thumbnail)
			}
			box.Children(func() {
				if cache.bitmap != nil {
					ui.Image(c, cache.bitmap).Size(160, 110).Fit(ui.Contain).Radius(6)
				}
				ui.Text(c, img.Name).FontSize(11).TextColor(theme.Muted)
			})
		}
	})
}
