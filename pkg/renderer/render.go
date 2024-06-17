package renderer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/elweday/go-subtitles/pkg/styles"
	"github.com/elweday/go-subtitles/pkg/types"
	"github.com/elweday/go-subtitles/pkg/utils"

	"github.com/fogleman/gg"
	"golang.org/x/image/font"
)

func SplitIntoLines(words []types.Word, bFont []byte, opts types.SubtitlesOptions) ([][]types.Word, map[int]int, map[int]float64) {

	indexLineMap := map[int]int{}
	lineWidthMap := map[int]float64{}

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
			lineWidthMap[lineIndex] = currWidth
			lineIndex += 1
			currWidth = float64(opts.Padding)
		}

		current = append(current, word)

		currWidth += wordWidth + spaceWidth

	}
	result = append(result, current)

	return result, indexLineMap, lineWidthMap
}

func DrawFrame2(lines [][]types.Word, widths []float64, idx int, perc float64, opts types.SubtitlesOptions, u types.Updater, regFont, boldFont font.Face) []byte {
	u.Update(&opts, perc)

	height := float64(opts.FontSize)*float64(opts.MaxLines)*opts.LineSpacing + 2*float64(opts.Padding)
	dc := gg.NewContext(opts.Width, int(height))

	dc.Clear()

	// dir := Iff(opts.RTL, 1, -1)

	dc.SetFontFace(regFont)

	currWidth := utils.Iff(opts.Center, (float64(opts.Width)-widths[0])/2, 0)
	currHeight := float64(opts.Padding)
	sep := strings.Repeat(" ", opts.WordSpacing)
	spaceWidth, _ := dc.MeasureString(sep)
	startX := utils.Iff(opts.RTL, opts.Width-opts.Padding, opts.Padding)
	startY := opts.Padding
	dir := utils.Iff(opts.RTL, -1.0, 1.0)
	lineHeight := opts.FontSize

	dc.SetHexColor(opts.FontColor)
	c := 0
	for i, line := range lines {
		lineStart := 0.0
		lineWidth := widths[i]
		if opts.Center {
			lineStart = (float64(opts.Width) - lineWidth) / 2
		}
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
				dc.SetFontFace(boldFont)
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
				dc.SetFontFace(regFont)
				dc.SetHexColor(opts.FontColor)
				dc.DrawString(word.Value, wordX, wordY)
			}

			currWidth += wordWidth + spaceWidth
		}
		currWidth = lineStart
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

func getLineWidths(m map[int]float64, start int, end int) []float64 {
	widths := []float64{}
	for i := start; i < end; i++ {
		widths = append(widths, m[i])
	}
	return widths
}

func (vid *VidoePayload) RenderWithSubtitles() error {

	// fontMap, err := GetFontWeightMapFromGoogle(opts.FontFamily, "arabic")

	/* if err != nil {
		fmt.Printf(err.Error(), "font not found")
		return
	}
	*/
	PREFIX := "serverless_function_source_code"
	if os.Getenv("GO_ENVIRONMENT") == "DEV" {
		PREFIX = ""
	}
	regFont, err1 := os.ReadFile(filepath.Join(PREFIX, "assets", "fonts", "Montserrat-Medium.ttf"))
	boldFont, err2 := os.ReadFile(filepath.Join(PREFIX, "assets", "fonts", "Montserrat-Bold.ttf"))
	if err1 != nil || err2 != nil {
		return fmt.Errorf("failed to read fonts: %v, %v", err1, err2)
	}

	lines, lineIndexMap, lineWidthMap := SplitIntoLines(vid.Words, regFont, vid.Opts)
	// fmt.Println(lines)

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
		for j := int64(0); j < frames; j++ {

			nframes := min(frames, duration)
			perc := float64(j) / float64(nframes)
			perc = min(perc, 1.0)

			wg.Add(1)

			go func(c int, idx int) {
				defer wg.Done()
				startLine := lineIndexMap[idx] - (lineIndexMap[idx] % vid.Opts.MaxLines)
				endLine := startLine + vid.Opts.MaxLines
				if endLine > len(lines) {
					endLine = len(lines)
				}
				widths := getLineWidths(lineWidthMap, startLine, endLine)
				selectedlines := lines[startLine:endLine]
				relativeIndex := calcRelativeIndex(lines, startLine, idx)
				reg := utils.ReadFont(regFont, vid.Opts.FontSize)
				bold := utils.ReadFont(boldFont, vid.Opts.FontSize)

				b := DrawFrame2(selectedlines, widths, relativeIndex, perc, vid.Opts, updater, reg, bold)
				mu.Lock()
				m[c] = b
				mu.Unlock()

			}(frameCount, iWord)
			frameCount++
		}
	}

	wg.Wait()

	arr := [][]byte{}

	for i := 0; i < frameCount; i++ {
		arr = append(arr, m[i])
	}

	aspectRatio := fmt.Sprintf("%dx%d", vid.Opts.Width, vid.Opts.Height)
	offset := 0.0
	videoHeight := float64(vid.Opts.FontSize)*float64(vid.Opts.MaxLines)*vid.Opts.LineSpacing + 2*float64(vid.Opts.Padding)

	switch vid.Opts.Alignment {
	case "top":
		offset = 0
	case "bottom":
		offset = float64(vid.Opts.Height) - videoHeight
	case "center":
		offset = float64(vid.Opts.Height)/2 - videoHeight/2
	}

	video, err := FFmpegCombineImagesToVideo(arr, vid.InputVideo, aspectRatio, vid.Opts.FPS, offset)

	vid.OutputVideo = video
	fmt.Println("video rendered")
	return err

}
