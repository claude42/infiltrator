package model

import (
	"fmt"

	// "log"
	"strings"

	"github.com/gdamore/tcell/v2"
)

type StringFilter struct {
	source            Filter
	filterFunc        func(input string) (string, [][]int, error)
	filterFuncFactory StringFilterFuncFactory
	eventHandler      tcell.EventHandler
	colorIndex        uint8
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

func NewStringFilter(fn StringFilterFuncFactory) *StringFilter {
	k := &StringFilter{}
	if fn != nil {
		k.filterFuncFactory = fn
	} else {
		k.filterFuncFactory = DefaultStringFilterFuncFactory
	}
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
	k.HandleEvent(NewEventFilterOutput())
	return nil
}

func (k *StringFilter) UpdateText(text string) error {
	return k.SetKey(text)
}

func (k *StringFilter) GetLine(line int) (Line, error) {
	sourceLine, err := k.source.GetLine(line)
	if err != nil {
		return Line{}, err
	}
	if k.filterFunc == nil {
		// For now just return the sourceLine
		// We'll determine later if this is the right thing to do
		return sourceLine, nil
	}
	_, indeces, err := k.filterFunc(sourceLine.Str)
	if err != nil {
		return sourceLine, err
	}

	k.colorizeLine(sourceLine, indeces)

	return sourceLine, nil
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

func (k *StringFilter) SetEventHandler(eventHandler tcell.EventHandler) {
	k.eventHandler = eventHandler
}

func (k *StringFilter) HandleEvent(ev tcell.Event) bool {
	if k.eventHandler == nil {
		return false
	}
	return k.eventHandler.HandleEvent(ev)
}

func (k *StringFilter) SetColorIndex(colorIndex uint8) {
	k.colorIndex = colorIndex
}
