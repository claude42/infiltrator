package reader

type LineStatus int

const (
	LineWithoutStatus LineStatus = iota
	LineMatched
	LineDimmed
	LineHidden
	LineDoesNotExist = -1
)

type Line struct {
	No      int
	Status  LineStatus
	Matched bool
	Str     string
	// each byte in ColorIndex is a color index for each byte in Str
	ColorIndex []uint8
}

func NewLine(lineNo int, text string) *Line {
	return &Line{lineNo, LineWithoutStatus, false, text, make([]uint8, len(text))}
}

func (l *Line) CleanUp() {
	l.Status = LineWithoutStatus
	l.Matched = false
	for i := range l.ColorIndex {
		l.ColorIndex[i] = 0
	}
}
