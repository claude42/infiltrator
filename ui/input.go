package ui

import (
	//"log"

	//"github.com/claude42/infiltrator/util"
	// "github.com/claude42/infiltrator/model"

	"github.com/gdamore/tcell/v2"
)

type Input interface {
	SetContent(content string)
	SetEventHandler(eh tcell.EventHandler)
	SetColorIndex(colorIndex uint8)
	Component
	tcell.EventHandler

	Component
}
