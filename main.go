package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"math"

	"os"
	"time"

	"github.com/abdullahdiaa/garabic"
	"github.com/elweday/go-subtitles/pkg/renderer"
	"github.com/elweday/go-subtitles/pkg/types"
)

func ReadAndConvertToFrames(jsonString []byte, frameRate int) ([]types.Word, error) {
	var items = []types.Word{
		{Time: 0, Value: "", Frames: 0},
	}
	reader := bytes.NewReader(jsonString)

	decoder := json.NewDecoder(reader)

	decoder.Decode(&items)

	// Convert time to frames

	for i := range items {
		items[i].Frames = int64(math.Round(items[i].Time * float64(frameRate)))
		items[i].Value = garabic.Shape(items[i].Value)
	}

	return items, nil
}

func main() {
	start := time.Now()
	defer since(start)

	backgroundVideo, err := os.ReadFile("temp/inputVideo.mp4")
	if err != nil {
		fmt.Println(err)
		return
	}

	w, h, err := renderer.FFmpegGetVideoDimensions(backgroundVideo)

	if err != nil {
		fmt.Println(err)
	}

	opts := types.SubtitlesOptions{
		FontFamily:            "nunito",
		FontSize:              40,
		FontColor:             color.RGBA{R: 8, G: 205, B: 237, A: 255},
		FontSelectedColor:     color.RGBA{R: 5, G: 253, B: 249, A: 255},
		StrokeColor:           color.RGBA{R: 255, G: 0, B: 255, A: 255},
		StrokeWidth:           15,
		HighlightColor:        color.RGBA{R: 12, G: 119, B: 135, A: 255},
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
		Width:                 w,
		Height:                h,
	}

	fileData, _ := os.ReadFile("temp/time_stamps_en.json")
	words, err := ReadAndConvertToFrames(fileData, opts.FPS)

	if err != nil {
		panic(err)
	}

	vid := &renderer.VidoePayload{
		InputVideo: backgroundVideo,
		Words:      words,
		Opts:       opts,
	}
	err = vid.Render()
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.WriteFile("temp/output.mkv", vid.SubtitledVideo, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func since(start time.Time) {
	fmt.Printf("\nElapsed time: %s\n", time.Since(start))
}
