package types

type Updater interface {
	Update(opts *SubtitlesOptions, perc float64)
}

/*
csdfjsdlf jsfkjsdhf kljsdfh
*/
type SubtitlesOptions struct {
	FontFamily            string
	FontSize              float64
	FontColor             string
	FontSelectedColor     string
	StrokeColor           string
	StrokeWidth           float64
	HighlightColor        string
	HighlightBorderRadius int
	HighlightPadding      float64
	HighlightScale        float64
	Padding               int
	LineWidth             int
	WordSpacing           int
	LineSpacing           float64
	TextOffsetX           float64
	TextOffsetY           float64
	TextOpacity           float64
	RTL                   bool
	MaxLines              int
	CurrentLine           int
	FPS                   int
	Width                 int
	Height                int
}

type Word struct {
	Time        float64 `json:"time"`
	Duration    float64 `json:"duration"`
	Value       string  `json:"word"`
	Frames      int64   `json:"frames"`
	StartFrames int64   `json:"startFrames"`
}

type Interpolator func(float64) float64

type SpringOptions struct {
	Stiffness float64 // Spring stiffness
	Damping   float64 // Damping coefficient
	Mass      float64 // Mass of the object
}

type Style struct {
	options *SubtitlesOptions
	Update  func(opts *SubtitlesOptions, perc float64)
	Check   func(words []Word, index int, i int) bool
}

func NewSubtitlesStyle(
	opts *SubtitlesOptions,
	Update func(opts *SubtitlesOptions, perc float64),
	Check func(words []Word, index int, i int) bool,
) *Style {

	return &Style{
		options: &SubtitlesOptions{},
		Update:  Update,
		Check:   Check,
	}

}
