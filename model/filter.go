package model

import (
	"errors"
	//"fmt"
	//"log"
	"time"

	// "github.com/claude42/infiltrator/util"

	"github.com/gdamore/tcell/v2"
)

type Filter interface {
	SetKey(key string) error
	GetLine(line int) (Line, error)
	Source() (Filter, error)
	SetSource(source Filter)
	Size() (int, int, error)
	Watch(eventHandler tcell.EventHandler)
	SetColorIndex(colorIndex uint8)

	tcell.EventHandler
}

type EventFilterOutput struct {
	time time.Time
}

var ErrOutOfBounds = errors.New("out of bounds")
var ErrLineDidNotMatch = errors.New("line did not match")

func NewEventFilterOutput() *EventFilterOutput {
	e := &EventFilterOutput{}
	e.time = time.Now()

	return e
}

func (e *EventFilterOutput) When() time.Time {
	return e.time
}
