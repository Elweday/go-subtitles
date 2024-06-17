package handlers

import (
	"fmt"

	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/utils"
)

type EndPointHandler struct {
	InputVideo []byte `json:"inputVideo"`
	Transcript []byte `json:"transcript"`
	Config     []byte `json:"config"`
	Out        []byte
}

func (handler *EndPointHandler) Read() (vid *renderer.VidoePayload, err error) {
	w, h, err := renderer.FFmpegGetVideoDimensions(handler.InputVideo)
	if err != nil {
		return nil, fmt.Errorf("failed to get video dimensions: %v", err)
	}
	opts := DefaultOptions
	opts.Width = w
	opts.Height = h

	words, err := utils.ReadAndConvertToFrames(handler.Transcript, opts.FPS)
	if err != nil {
		return nil, fmt.Errorf("cannot parse json object from body")
	}

	vid = &renderer.VidoePayload{
		InputVideo: handler.InputVideo,
		Words:      words,
		Opts:       opts,
	}

	return vid, nil

}

func (handler *EndPointHandler) SaveVideo(b []byte) error {
	handler.Out = b
	return nil
}
