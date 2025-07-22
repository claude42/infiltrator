package ui

import (
	//"fmt"
	//"log"

	//"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Panel interface {
	GetHeight() int
	ResizableRenderer
	tcell.EventHandler

	Component
}
