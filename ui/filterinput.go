package ui

import (
	"time"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model"
	"github.com/claude42/infiltrator/model/filter"
	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

type FilterInput struct {
	*ColoredInput

	filter filter.Filter
	name   string

	history             *[]string
	currentHistoryIndex int
	saveHistoryDelay    *util.Delay
}

func NewFilterInput(name string) *FilterInput {
	fi := &FilterInput{}
	fi.name = name
	fi.currentHistoryIndex = -1
	fi.ColoredInput = NewColoredInput()
	fi.ColoredInput.SetUpdateWatchersFunc(fi.updateWatchers)

	fi.saveHistoryDelay = util.NewCustomDelay(fi.storeInHistory, 2*time.Second)

	return fi
}

func (fi *FilterInput) storeInHistory() {
	// Certainly don't store empty lines in history
	if fi.ColoredInput.Content() == "" {
		return
	}

	// Also don't store anything when we're just browsing the history
	// without making any modifications
	if fi.currentHistoryIndex != -1 {
		currentHistoryEntry, err := config.GetConfiguration().FromHistory(fi.name,
			fi.currentHistoryIndex)
		if err != nil {
			// TODO: error handling
			return
		}
		if fi.ColoredInput.Content() == currentHistoryEntry {
			return
		}
	}

	config.GetConfiguration().AddToHistory(fi.name, fi.ColoredInput.Content())
	fi.currentHistoryIndex = 0
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
		case tcell.KeyUp:
			if ev.Modifiers() == 0 {
				content, err := config.GetConfiguration().FromHistory(fi.name,
					fi.currentHistoryIndex+1)
				if err != nil {
					// TODO error handling / beep
					return true
				}
				fi.currentHistoryIndex++
				fi.SetContent(content)
				return true
			}
		case tcell.KeyDown:
			if ev.Modifiers() == 0 {
				if fi.currentHistoryIndex <= -1 {
					// TODO error handling / beep
					return true
				}
				fi.currentHistoryIndex--
				if fi.currentHistoryIndex == -1 {
					fi.SetContent("")
				} else {
					content, err := config.GetConfiguration().FromHistory(fi.name,
						fi.currentHistoryIndex)
					if err != nil {
						// TODO error handling / beep
						return true
					}
					fi.SetContent(content)
				}
				return true
			}
		}
	}

	return fi.InputImpl.HandleEvent(ev)
}

func (fi *FilterInput) SetFilter(filter filter.Filter) {
	fi.filter = filter
}

// TODO TODO TODO
func (fi *FilterInput) updateWatchers() {
	if fi.filter == nil {
		return
	}
	model.GetFilterManager().UpdateFilterKey(fi.filter, fi.name, string(fi.ColoredInput.Content()))
	fi.saveHistoryDelay.Now()
	fi.ColoredInput.OldUpdateWatchersFunc()
}

func (fi *FilterInput) SetName(name string) {
	fi.name = name
}

func (fi *FilterInput) SetHistory(history *[]string) {
	fi.history = history
}
