package ui

import (
	//"fmt"
	//"log"

	"github.com/claude42/infiltrator/model"
	//"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Panel interface {
	Height() int
	Position() (int, int)
	SetColorIndex(colorIndex uint8)
	SetFilter(filter model.Filter)
	Filter() model.Filter
	Component
	tcell.EventHandler

	Component
}
