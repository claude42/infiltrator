package model

//"errors"
//"fmt"
//"log"
//"time"

// "github.com/claude42/infiltrator/util"

//"github.com/gdamore/tcell/v2"

type Line struct {
	No  int
	Str string
	// each byt in ColorIndex is a color index for each byte in Str
	ColorIndex []uint8
}
