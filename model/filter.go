package model

import (
	//"fmt"
	//"log"
	"time"

	"github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Filter interface {
	SetKey(key string) error
	GetLine(line int) (Line, error)
	Source() (Filter, error)
	SetSource(source Filter)
	Size() (int, int, error)
	SetColorIndex(colorIndex uint8)
	SetMode(mode int)
	SetCaseSensitive(on bool) error

	tcell.EventHandler
	util.Observable
}

type EventFilterOutput struct {
	time time.Time
}

func NewEventFilterOutput() *EventFilterOutput {
	e := &EventFilterOutput{}
	e.time = time.Now()

	return e
}

func (e *EventFilterOutput) When() time.Time {
	return e.time
}
