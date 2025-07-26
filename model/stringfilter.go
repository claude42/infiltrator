package model

import (
	// "errors"
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
	filterFunc        func(input string) (string, [][]int, bool)
	filterFuncFactory StringFilterFuncFactory
	eventHandler      tcell.EventHandler
	colorIndex        uint8
	mode              int
	key               string
	caseSensitive     bool
}

type StringFilterFuncFactory func(key string, caseSensitive bool) (func(input string) (string, [][]int, bool), error)

// TODO: include error handling?

func DefaultStringFilterFuncFactory(key string, caseSensitive bool) (func(input string) (string, [][]int, bool), error) {
	if !caseSensitive {
		key = strings.ToLower(key)
	}

	// Will return the matched string, an array of start/end pairs of matches
	// in the line and bool that's true if there was at least one match
	return func(input string) (string, [][]int, bool) {
		var indeces [][]int
		if len(key) == 0 {
			// Handle empty substring: returns all positions as zero-width matches.
			// This behavior mirrors regexp.FindAllStringIndex for empty patterns.
			for i := range input {
				indeces = append(indeces, []int{i, i})
			}
			return input, indeces, true
		}

		if !caseSensitive {
			input = strings.ToLower(input)
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
			return "", nil, false
		}

		return input, indeces, true
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

func (s *StringFilter) updateFilterFunc(key string, caseSensitive bool) error {
	var err error
	if s.filterFuncFactory != nil {
		s.filterFunc, err = s.filterFuncFactory(key, caseSensitive)
		if err != nil {
			return fmt.Errorf("error creating filter function: %w", err)
		}
	}
	if s.eventHandler != nil {
		s.eventHandler.HandleEvent(NewEventFilterOutput())
	}
	return nil
}

func (s *StringFilter) SetKey(key string) error {
	s.key = key
	return s.updateFilterFunc(s.key, s.caseSensitive)
}

func (s *StringFilter) SetCaseSensitive(on bool) error {
	s.caseSensitive = on
	return s.updateFilterFunc(s.key, s.caseSensitive)
}

func (s *StringFilter) SetMode(mode int) {
	s.mode = mode

	if s.eventHandler != nil {
		s.eventHandler.HandleEvent(NewEventFilterOutput())
	}
}

// ErrLineDidNotMatch errors are handled within GetLine() and will not
// buble up.
func (s *StringFilter) GetLine(line int) (Line, error) {
	sourceLine, err := s.source.GetLine(line)
	if err != nil {
		return sourceLine, err
	}

	if s.filterFunc == nil {
		// For now just return the sourceLine, don't touch its status
		// We'll determine later if this is the right thing to do
		return sourceLine, nil
	}

	_, indeces, matched := s.filterFunc(sourceLine.Str)

	if err != nil {
		log.Panicf("Unknown error from filter function: %w", err)
		return sourceLine, err
	}

	s.updateStatus(matched, indeces, &sourceLine)

	if !matched {
		return sourceLine, nil
	}

	if (s.mode == FilterMatch || s.mode == FilterFocus) &&
		sourceLine.Status != LineHidden {
		s.colorizeLine(sourceLine, indeces)
	}
	return sourceLine, nil
}

func (s *StringFilter) updateStatus(matched bool, indeces [][]int, sourceLine *Line) {
	switch s.mode {
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
		log.Panicf("Unkwon filter mdoe %d", s.mode)
	}
}

func (s *StringFilter) colorizeLine(line Line, indeces [][]int) {
	for _, index := range indeces {
		for i := index[0]; i < index[1]; i++ {
			line.ColorIndex[i] = s.colorIndex
		}
	}
}

func (s *StringFilter) Source() (Filter, error) {
	if s.source == nil {
		return nil, fmt.Errorf("no source defined")
	}

	return s.source, nil
}

func (s *StringFilter) SetSource(source Filter) {
	s.source = source
}

func (s *StringFilter) Size() (int, int, error) {
	//return 80, 0, nil // FIXME
	return s.source.Size()
}

func (s *StringFilter) Watch(eventHandler tcell.EventHandler) {
	s.eventHandler = eventHandler
}

func (s *StringFilter) Unwatch(eventHandler tcell.EventHandler) {
	// TODO: really, fix this!
	s.eventHandler = nil
}

func (s *StringFilter) HandleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *util.EventText:
		err := s.SetKey(ev.Text())
		return err == nil
	default:
		return false
	}
}

func (s *StringFilter) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}
