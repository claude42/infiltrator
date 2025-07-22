package model

import (
	//"bufio"
	//"fmt"
	//"log"
	//"os"
	//"time"

	// "github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type EventHandler interface {
	HandleEvent(ev tcell.Event) bool
}
