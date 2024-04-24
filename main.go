package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	types "github.com/elweday/go-subtitles/utils"
	styles "github.com/elweday/go-subtitles/utils/styles"
	"github.com/fogleman/gg"
)

// Item represents a single item in the JSON array

// ReadAndConvertToFrames reads the JSON file and converts time to frames
func ReadAndConvertToFrames(jsonString []byte, frameRate int) ([]types.Word, error) {
	var items []types.Word
	err := json.Unmarshal(jsonString, &items)
	if err != nil {
		return nil, err
	}

	// Convert time to frames
	for i := range items {
		items[i].Frames = int(math.Round(items[i].Time * float64(frameRate)))
	}

	items = append([]types.Word{
		{Time: 0, Value: "", Frames: 0},
	}, items...)


	return items, nil
}

const (
	Width     = 1080
	Height    = 1920
)

func DrawFrame(text []types.Word, idx int, perc float64, opts types.SubtitlesOptions, u  types.Updater) *gg.Context {
	u.Update(&opts, perc)
	var words = []types.Word{}
	for i:=0; i<len(text); i++ {
		if u.Check(text, idx, i){
			words = append(words, text[i])
		}
	}

	dc := gg.NewContext(Width, Height)

	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	fontPath := "font.ttf" 
	if err := dc.LoadFontFace(fontPath, opts.FontSize); err != nil {
		panic(err)
	}


	maxWidth := float64(Width - 2*opts.Padding)

	startX := opts.Padding
	startY := opts.Padding

	currWidth := 0.0
	currHeight := 0.0
	spaceWidth, lineHeight := dc.MeasureString(strings.Repeat(" ", opts.NSpaces))

	dc.SetColor(opts.FontColor)
	for i, word := range words {
		wordWidth, _ := dc.MeasureString(word.Value)

		if currWidth+wordWidth > maxWidth {
			currWidth = 0
			if lineHeight > opts.LineHeight {
				currHeight += lineHeight
			} else {
				currHeight += opts.LineHeight
			}
		}
		
		wordX := float64(startX)+float64(currWidth)
		wordY := float64(startY) + float64(currHeight)
		x := wordX + opts.TextOffsetX;
		y := wordY - lineHeight + opts.TextOffsetY;
		w := wordWidth
		h := lineHeight
		cx := x + w/2
		cy := y + h/2

		if i == idx {
			dc.Push()
			dc.SetColor(opts.HighlightColor)
			dc.ScaleAbout(1.2, 1.6, cx, cy)
			dc.DrawRoundedRectangle(x, y, w, h, float64(opts.HighlightBorderRadius))
			dc.Fill()
			dc.SetColor(opts.FontSelectedColor)
			dc.Pop()
			dc.DrawString(word.Value, wordX + opts.TextOffsetX, wordY + opts.TextOffsetY)
		} else {
			dc.SetColor(opts.FontColor)
			dc.DrawString(word.Value, wordX, wordY)
		}
		if i < len(text)-1 {
			currWidth += wordWidth+spaceWidth
		}

	}

	return dc
}

func Render(){
	start := time.Now()
	defer func(){
		fmt.Printf("Video created successfully! %s", time.Since(start))
	}()
	newpath := filepath.Join(".", "images")
	os.MkdirAll(newpath, os.ModePerm)
	fileData, _ := os.ReadFile("time_stamps.json")
	words, err := ReadAndConvertToFrames(fileData, 30)

	if err != nil {
		panic(err)
	}

	opts := types.SubtitlesOptions{
		FontPath: "font.ttf",
		FontSize: 60,
		FontColor: color.RGBA{ R: 255, G: 255, B: 255, A: 255 },
		FontSelectedColor: color.RGBA{ R: 255, G: 255, B: 0, A: 255 },
		StrokeColor: color.RGBA{ R: 255, G: 0, B: 255, A: 255 },
		StrokeWidth: 15,
		HighlightColor: color.RGBA{ R: 255, G: 0, B: 0, A: 255 },
		HighlightBorderRadius: 15,
		HighlightPadding: 20,
		Padding: 50,
		LineWidth: 5,
		NSpaces: 2,
		LineHeight: 80,
		TextOffsetX: 0,
		TextOffsetY: 0,
	}

	var wg sync.WaitGroup

	updater := styles.ScrollingBox{}


	count := 0
	for i:=1; i<len(words); i++ {
		current := words[i]
		prev := words[i-1]
		frames :=  current.Frames - prev.Frames
		for j := 0; j < frames; j++ {
			count++;
			perc := float64(j) / float64(frames)

			wg.Add(1)
			go func( _text []types.Word, _id int, _perc float64, _opts types.SubtitlesOptions, _count int, _updater types.Updater) {
				defer wg.Done()
				frameCtx := DrawFrame(_text, _id, _perc, _opts, _updater)
				output := fmt.Sprintf("images/img_%0*d.png", 5, _count)
				frameCtx.SavePNG(output)
			}(words, i, perc, opts, count, updater)

		}
	}


	wg.Wait()

	cmd := exec.Command("ffmpeg", "-y", "-framerate", fmt.Sprint(30), "-i", "./images/img_%05d.png", "-c:v", "libx264", "-r", "30", "-pix_fmt", "yuv420p", "output.mp4")

    err = cmd.Run()

    if err != nil {
        fmt.Println("Error:", err)
        return
    }

}


func main(){
	Render()
}


