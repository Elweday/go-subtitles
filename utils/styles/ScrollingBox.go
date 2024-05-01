package styles

import (
	types "github.com/elweday/go-subtitles/utils"
	"github.com/elweday/go-subtitles/utils/interpolation"
)

type ScrollingBox types.SubtitlesOptions

var f = interpolation.Spring(0.9, 1, types.SpringOptions{ Stiffness: 3, Damping: 0.1, Mass: 0.5 })

func (ScrollingBox) Update(opts *types.SubtitlesOptions, perc float64) {
	opts.HighlightScale = f(perc)
}
func (ScrollingBox) Check(words []types.Word, index int, i int) bool {
	return true
}


