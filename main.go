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

	"golang.org/x/exp/constraints"
	"golang.org/x/image/font"

	types "github.com/elweday/go-subtitles/utils"
	styles "github.com/elweday/go-subtitles/utils/styles"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
)


type Number interface {
    constraints.Integer | constraints.Float
}

func chunkBy[T any](items []T, chunkSize int) (chunks [][]T) {
    for chunkSize < len(items) {
        items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
    }
    return append(chunks, items)
}

/* func WrapLines(dc *gg.Context, text []types.Word, w float64, maxLines int, sep string) [][]types.Word {
	arr := []string{}

	for i:=0; i<len(text); i++ {
		arr = append(arr, text[i].Value)
	}
	
	start := 0;
	lines := [][]types.Word{}
	for _, chunk := range dc.WordWrap(strings.Join(arr, sep), w) {
		size := len(strings.Split(chunk, sep))
		lines = append(lines, text[start:start+size])
		start += size
	}
} */


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

func ReadFont(fontBytes []byte, size float64) font.Face {
	f, _ := truetype.Parse(fontBytes)
	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
	})
	return face;
}

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
	

	fontBytes, _ := os.ReadFile(opts.FontPathRegular)
	faceRegular := ReadFont(fontBytes, opts.FontSize)
	faceBold := ReadFont(fontBytes, opts.FontSize)
	
	dc.SetFontFace(faceRegular)

	maxWidth := float64(Width - 2*opts.Padding)

	currWidth := 0.0
	currHeight := float64(opts.Padding)
	sep := strings.Repeat(" ", opts.NSpaces)
	spaceWidth, lineHeight := dc.MeasureString(sep)
	startX := opts.Padding
	startY := opts.Padding

	
/* 	arr := []string{}
	for _, word := range words {
		arr = append(arr, word.Value)
	}
	

	fmt.Println(len(dc.WordWrap(strings.Join(arr, sep), maxWidth)))
 */	
	
	lineIndex := 0;
	dc.SetColor(opts.FontColor)
	for i, word := range words {
		wordWidth, _ := dc.MeasureString(word.Value)

		if currWidth+wordWidth > maxWidth {
			currWidth = 0
			lineIndex += 1
			currHeight += lineHeight * opts.LineHeight
		}
		
		wordX := float64(startX)+float64(currWidth)
		if lineIndex == 0 {
			wordX -= spaceWidth
		}
		wordY := float64(startY) + float64(currHeight)
		x := wordX - opts.HighlightPadding + opts.TextOffsetX;
		y := wordY - lineHeight - opts.HighlightPadding + opts.TextOffsetY;
		w := wordWidth + 2*opts.HighlightPadding
		h := lineHeight + 2*opts.HighlightPadding
		cx := x + w/2
		cy := y + h/2

		if i == idx {
			dc.SetFontFace(faceBold)
			dc.Push()
			dc.SetColor(opts.HighlightColor)
			dc.ScaleAbout(opts.HighlightScale, opts.HighlightScale, cx, cy)
			dc.DrawRoundedRectangle(x, y, w, h, float64(opts.HighlightBorderRadius))
			dc.Fill()
			dc.SetColor(opts.FontSelectedColor)
			dc.Stroke()
			dc.DrawString(word.Value, wordX + opts.TextOffsetX, wordY + opts.TextOffsetY)
			dc.Pop()
		} else {
			dc.SetFontFace(faceBold)
			dc.SetColor(opts.FontColor)
			dc.DrawString(word.Value, wordX, wordY)
		}
		if i < len(text)-1 {
			currWidth += wordWidth+spaceWidth
		}
	}

	return dc
}


func min[T Number](a, b T) T {
    if a < b {
        return a
    }
    return b
}

func Render(){
	start := time.Now()
	defer func(){
		fmt.Printf("Video created successfully! %s", time.Since(start))
	}()
	fps := 30
	newpath := filepath.Join(".", "images")
	os.MkdirAll(newpath, os.ModePerm)
	fileData, _ := os.ReadFile("time_stamps.json")
	words, err := ReadAndConvertToFrames(fileData, fps)

	if err != nil {
		panic(err)
	}

	opts := types.SubtitlesOptions{
		FontPathRegular: "Montserrat-Medium.ttf",
		FontPathBold: "Montserrat-Bold.ttf",
		FontSize: 70,
		FontColor: color.RGBA{ R: 8, G: 205, B: 237, A: 255 },
		FontSelectedColor: color.RGBA{ R: 5, G: 253, B: 249, A: 255 },
		StrokeColor: color.RGBA{ R: 255, G: 0, B: 255, A: 255 },
		StrokeWidth: 15,
		HighlightColor: color.RGBA{ R: 12, G: 119, B: 135, A: 255 },
		HighlightBorderRadius: 15,
		HighlightPadding: 10,
		Padding: 50,
		LineWidth: 5,
		NSpaces: 2,
		LineHeight: 1.3,
		TextOffsetX: 0,
		TextOffsetY: 0,
		HighlightScale: 1,
	}

	var wg sync.WaitGroup

	updater := styles.ScrollingBox{}

	durationS := 0.2
	duration := int(durationS * float64(fps))

	count := 0
	for i:=1; i<len(words); i++ {
		current := words[i]
		prev := words[i-1]
		frames :=  current.Frames - prev.Frames
		for j := 0; j < frames; j++ {
			count++;
			nframes := min(frames, duration)
			perc := float64(j) / float64(nframes)
			perc = min(perc, 1.0)

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

	cmd := exec.Command("ffmpeg",
		"-y",
		"-framerate", fmt.Sprint(fps),
		"-i", "./images/img_%05d.png",
		"-c:v", "libx264",
		"-pix_fmt", "yuva420p",
	    "output.mp4",
	)

    err = cmd.Run()

    if err != nil {
        fmt.Println("Error:", err)
        return
    }

}


func main(){
	Render()
}


