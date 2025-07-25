package util

import (
	"time"
	// "github.com/gdamore/tcell/v2"
)

type EventText struct {
	t    time.Time
	text string
}

// NewEventResize creates an EventResize with the new updated window size,
// which is given in character cells.
func NewEventText(text string) *EventText {
	return &EventText{t: time.Now(), text: text}
}

// When returns the time when the Event was created.
func (ev *EventText) When() time.Time {
	return ev.t
}

func (ev *EventText) Text() string {
	return ev.text
}

func (ev *EventText) SetText(text string) {
	ev.text = text
}
