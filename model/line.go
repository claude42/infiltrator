package model

const (
	LineWithoutStatus int = iota
	LineMatched
	LineDimmed
	LineHidden
	LineDoesNotExist = -1
)

type Line struct {
	No      int
	Status  int
	Matched bool
	Str     string
	// each byte in ColorIndex is a color index for each byte in Str
	ColorIndex []uint8
}

func (l *Line) CleanUp() {
	l.Status = LineWithoutStatus
	l.Matched = false
	for i := range l.ColorIndex {
		l.ColorIndex[i] = 0
	}
}
