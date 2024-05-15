package styles

import (
	"image/color"

	interpolation "github.com/elweday/go-subtitles/src/interpolation"
	"github.com/elweday/go-subtitles/src/types"
)

type AppearingWords struct{}

func (AppearingWords) Update(opts *types.SubtitlesOptions, perc float64) {
	opacity := uint8(255 * (perc))
	offset := interpolation.EaseIn(25, 0, 1)(perc)

	opts.TextOffsetY = offset
	opts.HighlightColor = color.RGBA{R: 0, G: 0, B: 0, A: 0}
	opts.FontSelectedColor = color.RGBA{R: 255, G: 0, B: 0, A: opacity}
}

func (AppearingWords) Check(words []types.Word, index int, i int) bool {
	return i <= index
}
