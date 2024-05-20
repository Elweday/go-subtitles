package renderer

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/elweday/go-subtitles/pkg/styles"
	"github.com/elweday/go-subtitles/pkg/types"
	"github.com/elweday/go-subtitles/pkg/utils"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func SplitIntoLines(words []types.Word, bFont []byte, opts types.SubtitlesOptions) ([][]types.Word, map[int]int) {

	indexLineMap := map[int]int{}

	fontFace := utils.ReadFont(bFont, opts.FontSize)
	maxWidth := float64(opts.Width)

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

func DrawFrame2(lines [][]types.Word, idx int, perc float64, opts types.SubtitlesOptions, u types.Updater, regFont, boldFont []byte) []byte {
	u.Update(&opts, perc)

	dc := gg.NewContext(opts.Width, opts.Height)

	dc.SetRGBA(0, 0, 0, 255)
	dc.Clear()

	// dir := Iff(opts.RTL, 1, -1)
	reg := utils.ReadFont(regFont, opts.FontSize)
	bold := utils.ReadFont(regFont, opts.FontSize)
	defer reg.Close()
	defer bold.Close()

	dc.SetFontFace(reg)

	currWidth := 0.0
	currHeight := float64(opts.Padding)
	sep := strings.Repeat(" ", opts.WordSpacing)
	spaceWidth, _ := dc.MeasureString(sep)
	startX := utils.Iff(opts.RTL, opts.Width-opts.Padding, opts.Padding)
	startY := opts.Padding
	dir := utils.Iff(opts.RTL, -1.0, 1.0)
	lineHeight := opts.FontSize

	dc.SetHexColor(opts.FontColor)
	c := 0
	for _, line := range lines {
		for _, word := range line {
			c++
			wordWidth, _ := dc.MeasureString(word.Value)
			wordX := float64(startX) + float64(currWidth)*dir + utils.Iff(opts.RTL, -wordWidth, 0)
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
				dc.SetHexColor(opts.HighlightColor)
				dc.ScaleAbout(opts.HighlightScale, opts.HighlightScale, cx, cy)
				dc.DrawRoundedRectangle(x, y, w, h, float64(opts.HighlightBorderRadius))
				dc.Fill()
				dc.SetHexColor(opts.FontSelectedColor)
				dc.Stroke()
				dc.DrawString(word.Value, wordX+opts.TextOffsetX, wordY+opts.TextOffsetY)
				dc.Pop()
			} else {
				dc.SetFontFace(reg)
				dc.SetHexColor(opts.FontColor)
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

func calcRelativeIndex(lines [][]types.Word, lineIndex, index int) int {
	sum := 0
	for _, line := range lines[:lineIndex] {
		sum += len(line)
	}
	return index - sum
}

type VidoePayload struct {
	InputVideoObj  string                 `firestore:"inputVideo"`
	OutputVideoObj string                 `firestore:"outputVideo"`
	Words          []types.Word           `firestore:"words"`
	Opts           types.SubtitlesOptions `firestore:"opts"`
	InputVideo     []byte
	OutputVideo    []byte
}

func (vid *VidoePayload) RenderWithSubtitles() error {

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

	aspectRatio := fmt.Sprintf("%dx%d", vid.Opts.Width, vid.Opts.Height)

	video, err := FFmpegCombineImagesToVideo(arr, vid.InputVideo, aspectRatio, vid.Opts.FPS)

	vid.OutputVideo = video

	return err

}
