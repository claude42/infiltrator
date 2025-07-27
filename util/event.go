package util

import (
	"time"

	"github.com/gdamore/tcell/v2"
)

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

type EventText struct {
	EventImpl

	text string
}

// NewEventResize creates an EventResize with the new updated window size,
// which is given in character cells.
func NewEventText(text string) *EventText {
	ev := &EventText{text: text}
	ev.EventImpl.SetWhen()
	return ev
}

func (ev *EventText) Text() string {
	return ev.text
}

func (ev *EventText) SetText(text string) {
	ev.text = text
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
