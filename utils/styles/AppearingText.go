package styles

import (
	"image/color"

	"github.com/elweday/go-subtitles/utils"
	"github.com/elweday/go-subtitles/utils/interpolation"
)

type AppearingWords struct{}

func (AppearingWords) Update(opts *utils.SubtitlesOptions, perc float64) {
	opacity :=  uint8(interpolation.EaseIn(150, 255, 2)(perc))
	offset := interpolation.EaseIn(50, 0, 2)(perc)

	opts.TextOffsetY = offset
	opts.HighlightColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	opts.FontSelectedColor = color.RGBA{R: opts.FontSelectedColor.R, G: opts.FontSelectedColor.G, B: opts.FontSelectedColor.B, A: opacity}
}
func (AppearingWords) Check(words []utils.Word, index int, i int) bool {
	return i <= index
}
