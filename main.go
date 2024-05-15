package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	_ "net/http/pprof"

	"math"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/goki/freetype"
	"github.com/goki/freetype/truetype"

	"golang.org/x/exp/constraints"
	"golang.org/x/image/font"

	"github.com/abdullahdiaa/garabic"

	types "github.com/elweday/go-subtitles/utils"
	styles "github.com/elweday/go-subtitles/utils/styles"
	"github.com/fogleman/gg"
)

type Number interface {
	constraints.Integer | constraints.Float
}

type FontPool struct {
	sync.Pool
	fontSize float64
}

func (p *FontPool) Get(key string) *font.Face {

	fonts := p.Pool.Get().(map[string]font.Face)
	defer p.Pool.Put(fonts)

	font, ok := fonts[key]
	if ok {
		return &font
	}

	return nil
}

func NewPool(m map[string][]byte, fontSize float64) *FontPool {
	return &FontPool{
		fontSize: fontSize,
		Pool: sync.Pool{
			New: func() any {
				items := map[string]font.Face{}
				for k, v := range m {
					face := ReadFont(v, fontSize)
					items[k] = face
				}
				return items
			},
		},
	}
}

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

const (
	Width  = 720
	Height = 1280
)

func Iff[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func ReadFont(fontBytes []byte, size float64) font.Face {
	f, _ := freetype.ParseFont(fontBytes)

	face := truetype.NewFace(f, &truetype.Options{
		Size: size,
	})
	return face
}

func SplitIntoLines(words []types.Word, bFont []byte, opts types.SubtitlesOptions) ([][]types.Word, map[int]int) {

	indexLineMap := map[int]int{}

	fontFace := ReadFont(bFont, opts.FontSize)
	maxWidth := float64(Width)

	currWidth := float64(opts.Padding)
	current := []types.Word{}
	result := [][]types.Word{}
	drawer := &font.Drawer{
		Face: fontFace,
	}
	sep := strings.Repeat(" ", opts.WordSpacing)
	spaceWidth := float64(drawer.MeasureString(sep) >> 6)
	lineIndex := 0

	for i, word := range words {
		indexLineMap[i] = lineIndex
		wordWidth := float64(drawer.MeasureString(word.Value) >> 6)

		if currWidth+wordWidth+spaceWidth+float64(opts.Padding) > maxWidth-float64(opts.Padding) {
			result = append(result, current)
			current = []types.Word{}
			lineIndex += 1
			currWidth = float64(opts.Padding)
		}

		current = append(current, word)

		currWidth += wordWidth + spaceWidth

	}
	result = append(result, current)

	return result, indexLineMap
}

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

	/* 	arr := []string{}
	   	for _, word := range words {
	   		arr = append(arr, word.Value)
	   	}


	   	fmt.Println(len(dc.WordWrap(strings.Join(arr, sep), maxWidth)))
	*/

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

func DrawFrame2(lines [][]types.Word, idx int, perc float64, opts types.SubtitlesOptions, u types.Updater, regFont, boldFont []byte) []byte {
	u.Update(&opts, perc)

	dc := gg.NewContext(Width, Height)

	dc.SetRGBA(0, 0, 0, 255)
	dc.Clear()

	// dir := Iff(opts.RTL, 1, -1)
	reg := ReadFont(regFont, opts.FontSize)
	bold := ReadFont(regFont, opts.FontSize)
	defer reg.Close()
	defer bold.Close()

	dc.SetFontFace(reg)

	currWidth := 0.0
	currHeight := float64(opts.Padding)
	sep := strings.Repeat(" ", opts.WordSpacing)
	spaceWidth, _ := dc.MeasureString(sep)
	startX := Iff(opts.RTL, Width-opts.Padding, opts.Padding)
	startY := opts.Padding
	dir := Iff(opts.RTL, -1.0, 1.0)
	lineHeight := opts.FontSize

	dc.SetColor(opts.FontColor)
	c := 0
	for _, line := range lines {
		for _, word := range line {
			c++
			wordWidth, _ := dc.MeasureString(word.Value)
			wordX := float64(startX) + float64(currWidth)*dir + Iff(opts.RTL, -wordWidth, 0)
			wordY := float64(startY) + float64(currHeight)
			x := wordX - opts.HighlightPadding + opts.TextOffsetX
			y := wordY - lineHeight - opts.HighlightPadding + opts.TextOffsetY + (opts.FontSize * 0.23)
			w := wordWidth + 2*opts.HighlightPadding
			h := lineHeight + 2*opts.HighlightPadding
			cx := x + w/2
			cy := y + h/2

			if c == idx {
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

			currWidth += wordWidth + spaceWidth
		}
		currWidth = 0
		currHeight += lineHeight * opts.LineSpacing

	}

	var buf bytes.Buffer
	if err := dc.EncodePNG(&buf); err != nil {
		panic(err)
	}

	return buf.Bytes()
}

func min[T Number](a, b T) T {
	if a < b {
		return a
	}
	return b
}

func calcRelativeIndex(lines [][]types.Word, lineIndex, index int) int {
	sum := 0
	for _, line := range lines[:lineIndex] {
		sum += len(line)
	}
	return index - sum
}

func Render(vid *types.VidoePayload) error {

	// fontMap, err := GetFontWeightMapFromGoogle(opts.FontFamily, "arabic")

	/* if err != nil {
		fmt.Printf(err.Error(), "font not found")
		return
	}
	*/

	regFont, _ := os.ReadFile("./assets/fonts/Montserrat-Medium.ttf")
	boldFont, _ := os.ReadFile("./assets/fonts/Montserrat-Bold.ttf")

	lines, lineIndexMap := SplitIntoLines(vid.Words, regFont, vid.Opts)
	fmt.Println(lineIndexMap)

	// fmt.Println(lines)
	os.Stdout = nil

	var wg sync.WaitGroup

	updater := styles.ScrollingBox{}

	durationS := 0.2
	duration := int64(durationS * float64(vid.Opts.FPS))

	m := map[int][]byte{}
	mu := sync.Mutex{}

	frameCount := 0
	for iWord := 1; iWord < len(vid.Words); iWord++ {
		current := vid.Words[iWord]
		prev := vid.Words[iWord-1]
		frames := current.Frames - prev.Frames
		if current.Value == "" {
			continue
		}
		for j := int64(0); j < frames; j++ {

			nframes := min(frames, duration)
			perc := float64(j) / float64(nframes)
			perc = min(perc, 1.0)

			wg.Add(1)
			go func(c int) {
				defer wg.Done()

				startLine := lineIndexMap[iWord] - (lineIndexMap[iWord] % vid.Opts.MaxLines)
				endLine := startLine + vid.Opts.MaxLines
				if endLine > len(lines) {
					endLine = len(lines)
				}
				selectedlines := lines[startLine:endLine]
				relativeIndex := calcRelativeIndex(lines, startLine, iWord)
				b := DrawFrame2(selectedlines, relativeIndex, perc, vid.Opts, updater, regFont, boldFont)
				mu.Lock()
				m[c] = b
				mu.Unlock()
			}(frameCount)
			frameCount++
		}
	}

	wg.Wait()

	arr := [][]byte{}

	for i := range frameCount {
		arr = append(arr, m[i])
	}

	fmt.Println("images created")

	aspectRatio := fmt.Sprintf("%dx%d", Width, Height)

	video, err := CombineImagesToVideo(arr, vid.InputVideo, aspectRatio, vid.Opts.FPS)

	vid.SubtitledVideo = video

	return err

}

func main() {
	start := time.Now()
	defer since(start)

	backgroundVideo, err := os.ReadFile("inputVideo.mp4")
	if err != nil {
		fmt.Println(err)
		return
	}

	w, h, err := GetVideoDimensions(backgroundVideo)

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

	fileData, _ := os.ReadFile("time_stamps_en.json")
	words, err := ReadAndConvertToFrames(fileData, opts.FPS)

	if err != nil {
		panic(err)
	}

	vid := &types.VidoePayload{
		InputVideo: backgroundVideo,
		Words:      words,
		Opts:       opts,
	}
	err = Render(vid)
	if err != nil {
		fmt.Println(err)
		return
	}
	err = os.WriteFile("output.mkv", vid.SubtitledVideo, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}

}

func since(start time.Time) {
	fmt.Printf("\nElapsed time: %s\n", time.Since(start))
}
