package render

type MarkerType int

const (
	MarkerNormal MarkerType = iota
	MarkerStar
	MarkerDot
	MarkerWarn
)

type Run struct {
	Text    string
	Bold    bool
	Italic  bool
	Formula bool
}

type Block struct {
	Marker MarkerType
	Runs   []Run
}

type ParseOptions struct {
	FormulaDelimiter string
	StarPrefix       string
	DotPrefix        string
	WarnPrefix       string
}

type Warning struct {
	Code    string
	Message string
}

type Column struct {
	Lang   string
	Blocks []Block
}

type Slide struct {
	FontSize           int
	Columns            []Column
	ENNumCol           int
	CNNumCol           int
	HasTruncationBadge bool
}
