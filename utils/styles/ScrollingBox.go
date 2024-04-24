package styles

import (
	types "github.com/elweday/go-subtitles/utils"
	// "github.com/elweday/go-subtitles/utils/interpolation"
)

type ScrollingBox types.SubtitlesOptions


// var f = interpolation.Spring(30, 20, types.SpringOptions{ Stiffness: 100, Damping: 10, Mass: 10 })


func (ScrollingBox) Update(opts *types.SubtitlesOptions, perc float64) {
	// opts.HighlightPadding = f(perc)
}
func (ScrollingBox) Check(words []types.Word, index int, i int) bool {
	return true
}
