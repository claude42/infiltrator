package ui

import (
	//"log"

	//"github.com/claude42/infiltrator/util"
	"github.com/claude42/infiltrator/model"

	"github.com/gdamore/tcell/v2"
)

type Input interface {
	SetContent(content string)
	SetReceiver(receiver model.UpdatedTextReceiver)
	ResizableRenderer
	tcell.EventHandler

	Component
}
