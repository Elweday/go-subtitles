package styles

import (
	"github.com/elweday/go-subtitles/utils"
	"github.com/elweday/go-subtitles/utils/interpolation"
)

type AppearingWords struct{}

func (AppearingWords) Update(opts *utils.SubtitlesOptions, perc float64) {
	opacity :=  uint8(interpolation.Linear(0.5, 1)(perc))
	offset := interpolation.Linear(50, 0)(perc)

	opts.TextOffsetY = offset
	opts.HighlightColor.A = opacity
	opts.FontColor.A = opacity
}
func (AppearingWords) Check(words []utils.Word, index int, i int) bool {
	return i <= index
}
