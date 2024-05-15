package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	_ "net/http/pprof"

	"os"
	"time"

	"github.com/abdullahdiaa/garabic"
	"github.com/elweday/go-subtitles/src/renderer"
	"github.com/elweday/go-subtitles/src/types"
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

/*
	 func DrawFrame(text []types.Word, idx int, perc float64, opts types.SubtitlesOptions, u types.Updater, regFont, boldFont []byte) []byte {
		u.Update(&opts, perc)
		var words = []types.Word{}
		for i := 0; i < len(text); i++ {
			if u.Check(text, idx, i) {
				words = append(words, text[i])
			}
		}

		dc := gg.NewContext(Width, Height)

		dc.SetRGBA(0, 0, 0, 255)
		dc.Clear()

		// dir := Iff(opts.RTL, 1, -1)
		reg := ReadFont(regFont, opts.FontSize)
		bold := ReadFont(regFont, opts.FontSize)

		dc.SetFontFace(reg)

		maxWidth := float64(Width - 2*opts.Padding)

		currWidth := 0.0
		currHeight := float64(opts.Padding)
		sep := strings.Repeat(" ", opts.WordSpacing)
		spaceWidth, _ := dc.MeasureString(sep)
		startX := Iff(opts.RTL, Width-opts.Padding, opts.Padding)
		startY := opts.Padding
		dir := Iff(opts.RTL, -1.0, 1.0)

			arr := []string{}
		   	for _, word := range words {
		   		arr = append(arr, word.Value)
		   	}


		   	fmt.Println(len(dc.WordWrap(strings.Join(arr, sep), maxWidth)))


		lineIndex := 0
		dc.SetColor(opts.FontColor)
		for i, word := range words {

			wordWidth, _ := dc.MeasureString(word.Value)
			lineHeight := opts.FontSize

			if currWidth+wordWidth > maxWidth {
				currWidth = 0
				lineIndex += 1
				if lineIndex%opts.MaxLines == 0 {

					opts.CurrentLine += opts.MaxLines
					fmt.Printf("Current line: %d, Last word: %s \n", opts.CurrentLine, words[i-1].Value)
				}
				currHeight += lineHeight * opts.LineSpacing
			}

			wordX := float64(startX) + float64(currWidth)*dir + Iff(opts.RTL, -wordWidth, 0)
			if lineIndex == 0 {
				wordX -= spaceWidth
			}
			wordY := float64(startY) + float64(currHeight)
			x := wordX - opts.HighlightPadding + opts.TextOffsetX
			y := wordY - lineHeight - opts.HighlightPadding + opts.TextOffsetY + (opts.FontSize * 0.23)
			w := wordWidth + 2*opts.HighlightPadding
			h := lineHeight + 2*opts.HighlightPadding
			cx := x + w/2
			cy := y + h/2

			if i == idx {
				dc.SetFontFace(bold)
				dc.Push()
				dc.SetColor(opts.HighlightColor)
				dc.ScaleAbout(opts.HighlightScale, opts.HighlightScale, cx, cy)
				dc.DrawRoundedRectangle(x, y, w, h, float64(opts.HighlightBorderRadius))
				dc.Fill()
				dc.SetColor(opts.FontSelectedColor)
				dc.Stroke()
				dc.DrawString(word.Value, wordX+opts.TextOffsetX, wordY+opts.TextOffsetY)
				dc.Pop()
			} else {
				dc.SetFontFace(reg)
				dc.SetColor(opts.FontColor)
				dc.DrawString(word.Value, wordX, wordY)
			}
			if Iff(opts.RTL, i > 0, i < len(text)-1) {
				currWidth += wordWidth + spaceWidth
			}

		}

		var buf bytes.Buffer
		if err := dc.EncodePNG(&buf); err != nil {
			panic(err)
		}

		reg.Close()
		bold.Close()
		return buf.Bytes()
	}
*/
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
