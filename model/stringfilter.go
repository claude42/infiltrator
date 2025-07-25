package model

import (
	"errors"
	"fmt"

	"log"
	"strings"

	"github.com/claude42/infiltrator/util"
	"github.com/gdamore/tcell/v2"
)

const (
	FilterMatch int = iota
	FilterFocus
	FilterHide
)

type StringFilter struct {
	source            Filter
	filterFunc        func(input string) (string, [][]int, error)
	filterFuncFactory StringFilterFuncFactory
	eventHandler      tcell.EventHandler
	colorIndex        uint8
	mode              int
}

type StringFilterFuncFactory func(key string) (func(input string) (string, [][]int, error), error)

// TODO: include error handling?

func DefaultStringFilterFuncFactory(key string) (func(input string) (string, [][]int, error), error) {
	return func(input string) (string, [][]int, error) {
		var indeces [][]int
		if len(key) == 0 {
			// Handle empty substring: returns all positions as zero-width matches.
			// This behavior mirrors regexp.FindAllStringIndex for empty patterns.
			for i := range input {
				indeces = append(indeces, []int{i, i})
			}
			return input, indeces, nil
		}

		offset := 0

		for {
			index := strings.Index(input[offset:], key)
			if index == -1 {
				break
			}

			start := offset + index
			end := start + len(key)
			indeces = append(indeces, []int{start, end})

			offset = end
		}

		if len(indeces) == 0 {
			return "", nil, ErrLineDidNotMatch
		}

		return input, indeces, nil
	}, nil
}

func NewStringFilter(fn StringFilterFuncFactory, mode int) *StringFilter {
	k := &StringFilter{}

	if fn != nil {
		k.filterFuncFactory = fn
	} else {
		k.filterFuncFactory = DefaultStringFilterFuncFactory
	}

	k.mode = mode
	return k
}

func (k *StringFilter) SetKey(key string) error {
	var err error
	if k.filterFuncFactory != nil {
		k.filterFunc, err = k.filterFuncFactory(key)
		if err != nil {
			return fmt.Errorf("error creating filter function: %w", err)
		}
	}
	if k.eventHandler != nil {
		k.eventHandler.HandleEvent(NewEventFilterOutput())
	}
	return nil
}

func (k *StringFilter) SetMode(mode int) {
	k.mode = mode

	if k.eventHandler != nil {
		k.eventHandler.HandleEvent(NewEventFilterOutput())
	}
}

// ErrLineDidNotMatch errors are handled within GetLine() and will not
// buble up.
func (k *StringFilter) GetLine(line int) (Line, error) {
	sourceLine, err := k.source.GetLine(line)
	if err != nil {
		return sourceLine, err
	}

	if k.filterFunc == nil {
		// For now just return the sourceLine, don't touch its status
		// We'll determine later if this is the right thing to do
		return sourceLine, nil
	}

	_, indeces, err := k.filterFunc(sourceLine.Str)

	if err != nil && !errors.Is(err, ErrLineDidNotMatch) {
		log.Panicf("Unknown filter mode %d", k.mode)
		return sourceLine, err
	}

	matched := !errors.Is(err, ErrLineDidNotMatch)
	k.updateStatus(matched, indeces, &sourceLine)

	if !matched {
		return sourceLine, nil
	}

	if (k.mode == FilterMatch || k.mode == FilterFocus) &&
		sourceLine.Status != LineHidden {
		k.colorizeLine(sourceLine, indeces)
	}
	return sourceLine, nil
}

func (k *StringFilter) updateStatus(matched bool, indeces [][]int, sourceLine *Line) {
	switch k.mode {
	case FilterMatch:
		if sourceLine.Status == LineWithoutStatus && matched {
			sourceLine.Status = LineMatched
		} else if !matched {
			sourceLine.Status = LineHidden
		}
	case FilterFocus:
		switch sourceLine.Status {
		case LineWithoutStatus:
			if matched {
				sourceLine.Status = LineMatched
			} else {
				sourceLine.Status = LineDimmed
			}
		case LineMatched:
			if !matched {
				sourceLine.Status = LineDimmed
			}
		}
	case FilterHide:
		if matched && indeces[0][1] != 0 {
			sourceLine.Status = LineHidden
		}
	default:
		log.Panicf("Unkwon filter mdoe %d", k.mode)
	}
}

func (k *StringFilter) colorizeLine(line Line, indeces [][]int) {
	for _, index := range indeces {
		for i := index[0]; i < index[1]; i++ {
			line.ColorIndex[i] = k.colorIndex
		}
	}
}

func (k *StringFilter) Source() (Filter, error) {
	if k.source == nil {
		return nil, fmt.Errorf("no source defined")
	}

	return k.source, nil
}

func (k *StringFilter) SetSource(source Filter) {
	k.source = source
}

func (k *StringFilter) Size() (int, int, error) {
	//return 80, 0, nil // FIXME
	return k.source.Size()
}

func (k *StringFilter) Watch(eventHandler tcell.EventHandler) {
	k.eventHandler = eventHandler
}

func (k *StringFilter) Unwatch(eventHandler tcell.EventHandler) {
	// TODO: really, fix this!
	k.eventHandler = nil
}

func (k *StringFilter) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *util.EventText:
		err := k.SetKey(ev.Text())
		return err == nil
	default:
		return false
	}
}

func (k *StringFilter) SetColorIndex(colorIndex uint8) {
	k.colorIndex = colorIndex
}
