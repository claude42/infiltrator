package ui

import (
	"github.com/claude42/infiltrator/model"
	"github.com/gdamore/tcell/v2"
)

type FilterInput struct {
	InputImpl

	filter model.Filter
}

func NewFilterInput() *FilterInput {
	fi := &FilterInput{}
	fi.InputImpl.inputCorrect = true
	fi.InputImpl.updateWatchers = fi.updateWatchers

	return fi
}

func (fi *FilterInput) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEnter:
			// posting this globally makes things easier but not sure if it's
			// the right thing to do
			screen.PostEvent(NewEventPressedEnterInInputField())
			model.GetFilterManager().FindMatch(1)
			return true
		}
	}

	return fi.InputImpl.HandleEvent(ev)
}

func (fi *FilterInput) SetFilter(filter model.Filter) {
	fi.filter = filter
}

func (fi *FilterInput) updateWatchers() {
	model.GetFilterManager().UpdateFilterKey(fi.filter, string(fi.InputImpl.content))
	fi.InputImpl.defaultUpdateWatchers()
}
