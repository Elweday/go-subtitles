package handlers

import (
	"os"

	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/types"
)

type IOHandler interface {
	Read() (vid *renderer.VidoePayload, err error)
	SaveVideo(b []byte) error
}

func GetIOHnadler() IOHandler {
	runEnv := os.Getenv("SUBTITLES_RUN_ENVIRONMENT")
	if runEnv == "LOCAL" {
		return NewLocalHandler()

	} else if runEnv == "GCP" {
		return NewGcpHandler()
	}
	return nil
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
	LineWidth:             5,
	WordSpacing:           3,
	LineSpacing:           1.6,
	TextOffsetX:           0,
	TextOffsetY:           0,
	HighlightScale:        1,
	RTL:                   false,
	MaxLines:              3,
	FPS:                   30,
}
