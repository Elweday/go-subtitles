package types

type Updater interface {
	Update(opts *SubtitlesOptions, perc float64)
}

/*
csdfjsdlf jsfkjsdhf kljsdfh
*/
type SubtitlesOptions struct {
	FontFamily            string  `firestore:"fontFamily"`
	FontSize              float64 `firestore:"fontSize"`
	FontColor             string  `firestore:"fontColor"`
	FontSelectedColor     string  `firestore:"fontSelectedColor"`
	StrokeColor           string  `firestore:"strokeColor"`
	StrokeWidth           float64 `firestore:"strokeWidth"`
	HighlightColor        string  `firestore:"highlightColor"`
	HighlightBorderRadius int     `firestore:"highlightBorderRadius"`
	HighlightPadding      float64 `firestore:"highlightPadding"`
	Padding               int     `firestore:"padding"`
	WordSpacing           int     `firestore:"wordSpacing"`
	LineSpacing           float64 `firestore:"lineSpacing"`
	RTL                   bool    `firestore:"rtl"`
	MaxLines              int     `firestore:"maxLines"`
	Center                bool    `firestore:"center"`
	Alignment             string  `firestore:"alignment"`
	HighlightScale        float64
	TextOffsetX           float64
	TextOffsetY           float64
	TextOpacity           float64
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
