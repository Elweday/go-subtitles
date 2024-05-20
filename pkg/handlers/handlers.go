package handlers

import (
	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/types"
)

type IOHandler interface {
	Read() (vid *renderer.VidoePayload, err error)
	SaveVideo(b []byte) error
}

var DefaultOptions = types.SubtitlesOptions{
	FontFamily:            "nunito",
	FontSize:              40,
	FontColor:             "08cded",
	FontSelectedColor:     "05fdf9",
	StrokeColor:           "ff00ff",
	StrokeWidth:           15,
	HighlightColor:        "0c7787",
	HighlightBorderRadius: 15,
	HighlightPadding:      15,
	Padding:               40,
	WordSpacing:           3,
	LineSpacing:           1.6,
	TextOffsetX:           0,
	TextOffsetY:           0,
	HighlightScale:        1,
	RTL:                   false,
	MaxLines:              2,
	FPS:                   30,
	Center:                true,
	Alignment:             "center",
}
