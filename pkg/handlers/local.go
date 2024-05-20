package handlers

import (
	"fmt"
	"os"

	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/utils"
)

type LocalIOHandler struct {
	InputVideoPath string
	TranscriptPath string
	ConfigPath     string
	OutputPath     string
}

func (handler *LocalIOHandler) Read() (vid *renderer.VidoePayload, err error) {

	inputVideo, err := os.ReadFile(handler.InputVideoPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s", handler.InputVideoPath)
	}

	w, h, err := renderer.FFmpegGetVideoDimensions(inputVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to get video dimensions: %v", err)
	}
	opts := DefaultOptions
	opts.Width = w
	opts.Height = h

	transcriptBytes, err := os.ReadFile(handler.TranscriptPath)
	if err != nil {
		return nil, fmt.Errorf("file %s does not exist, make sure you set SUBTITLES_TRANSCRIPT_PATH environment variable to a json file that follows the correct format", handler.TranscriptPath)
	}

	words, err := utils.ReadAndConvertToFrames(transcriptBytes, opts.FPS)
	if err != nil {
		return nil, fmt.Errorf("file %s does not follow the correct format", handler.TranscriptPath)
	}

	vid = &renderer.VidoePayload{
		InputVideo: inputVideo,
		Words:      words,
		Opts:       opts,
	}

	return vid, nil

}

func (handler *LocalIOHandler) SaveVideo(b []byte) error {
	return os.WriteFile(handler.OutputPath, b, 0644)
}
