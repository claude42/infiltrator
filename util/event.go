package util

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

type Event interface {
	tcell.Event
	When() time.Time
}

type EventImpl struct {
	t time.Time
}

func (ev *EventImpl) SetWhen() {
	ev.t = time.Now()
}

// When returns the time when the Event was created.
func (ev *EventImpl) When() time.Time {
	return ev.t
}

// EventHandler

type EventHandler interface {
	HandleEvent(tcell.Event) bool
}

type EventHandlerIgnoreImpl struct {
}

func (eh *EventHandlerIgnoreImpl) HandleEvent(tcell.Event) bool {
	return false
}

type EventHandlerPanicImpl struct {
}

func (eh *EventHandlerPanicImpl) HandleEvent(tcell.Event) bool {
	panic("HandlEvent() implementation missing!")
}
