package styles

import (
	"github.com/elweday/go-subtitles/pkg/types"
	interpolation "github.com/elweday/go-subtitles/pkg/utils/interpolation"
)

type AppearingWords types.SubtitlesOptions

func (opts *AppearingWords) Update(perc float64) {
	// opacity := uint8(255 * (perc))
	offset := interpolation.EaseIn(25, 0, 1)(perc)

	opts.TextOffsetY = offset
	opts.HighlightColor = "000000"
	opts.FontSelectedColor = "FF0000"
}
