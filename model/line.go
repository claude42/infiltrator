package model

//"errors"
//"fmt"
//"log"
//"time"

// "github.com/claude42/infiltrator/util"

//"github.com/gdamore/tcell/v2"

const (
	LineWithoutStatus int = iota
	LineMatched
	LineDimmed
	LineHidden
)

type Line struct {
	No     int
	Status int
	Str    string
	// each byte in ColorIndex is a color index for each byte in Str
	ColorIndex []uint8
}
