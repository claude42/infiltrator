package lines

import "time"

type LineStatus int

const (
	LineWithoutStatus LineStatus = iota
	LineMatched
	LineDimmed
	LineHidden
	LineDoesNotExist = -1
)

var NonExistingLine = &Line{
	No:      -1,
	Status:  LineDoesNotExist,
	Matched: false,
	Str:     "",
}

type Line struct {
	No      int
	Status  LineStatus
	Matched bool
	Str     string
	When    time.Time
	// each byte in ColorIndex is a color index for each byte in Str
	ColorIndex []uint8
}

func NewLine(lineNo int, text string) *Line {
	return &Line{
		No:         lineNo,
		Status:     LineWithoutStatus,
		Matched:    false,
		Str:        text,
		ColorIndex: make([]uint8, len(text)),
	}
}

func (l *Line) CleanUp() {
	l.Status = LineWithoutStatus
	l.Matched = false
	for i := range l.ColorIndex {
		l.ColorIndex[i] = 0
	}
}
