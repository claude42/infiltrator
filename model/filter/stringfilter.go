package filter

import (
	"fmt"
	"sync"

	"log"
	"strings"

	"github.com/claude42/infiltrator/model/lines"
)

type StringFilter struct {
	FilterImpl
	sync.Mutex
	filterFunc        func(input string) (string, [][]int, bool)
	filterFuncFactory StringFilterFuncFactory
	colorIndex        uint8
	mode              FilterMode
	key               string
	caseSensitive     bool
}

type StringFilterFuncFactory func(key string, caseSensitive bool) (func(input string) (string, [][]int, bool), error)

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

func NewStringFilter(fn StringFilterFuncFactory, mode FilterMode) *StringFilter {
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

	return nil
}

func (s *StringFilter) SetKey(name string, key string) error {
	// don't care about the name
	s.Lock()
	s.key = key
	s.Unlock()
	return s.updateFilterFunc(s.key, s.caseSensitive)
}

func (s *StringFilter) SetCaseSensitive(on bool) error {
	s.Lock()
	s.caseSensitive = on
	s.Unlock()
	return s.updateFilterFunc(s.key, s.caseSensitive)
}

func (s *StringFilter) SetMode(mode FilterMode) {
	s.Lock()
	s.mode = mode
	s.Unlock()
}

// ErrLineDidNotMatch errors are handled within GetLine() and will not
// buble up.
func (s *StringFilter) GetLine(line int) (*lines.Line, error) {
	sourceLine, err := s.source.GetLine(line)
	if err != nil {
		return sourceLine, err
	}

	s.Lock()
	defer s.Unlock()

	if sourceLine.Status == lines.LineHidden {
		return sourceLine, nil
	}

	if s.filterFunc == nil || s.key == "" {
		return sourceLine, nil
	}

	_, indeces, matched := s.filterFunc(sourceLine.Str)

	s.updateStatusAndMatched(matched, indeces, sourceLine)

	if !matched {
		// no further coloring necessary, bail out here
		return sourceLine, nil
	}

	if (s.mode == FilterMatch || s.mode == FilterFocus) &&
		sourceLine.Status != lines.LineHidden {
		s.colorizeLine(sourceLine, indeces)
	}
	return sourceLine, nil
}

func (s *StringFilter) updateStatusAndMatched(matched bool, indeces [][]int, sourceLine *lines.Line) {
	newStatus := sourceLine.Status
	newMatched := sourceLine.Matched
	switch s.mode {
	case FilterMatch:
		// Status
		if sourceLine.Status == lines.LineWithoutStatus && matched {
			newStatus = lines.LineMatched
		} else if !matched {
			newStatus = lines.LineHidden
		}

		// Matched
		if !sourceLine.Matched && matched &&
			(sourceLine.Status == lines.LineWithoutStatus || sourceLine.Status == lines.LineDimmed) {

			newMatched = true
		}
	case FilterFocus:
		// Status
		switch sourceLine.Status {
		case lines.LineWithoutStatus:
			if matched {
				newStatus = lines.LineMatched
			} else {
				newStatus = lines.LineDimmed
			}
		case lines.LineMatched:
			if !matched {
				newStatus = lines.LineDimmed
			}
		}

		// Matched
		if !sourceLine.Matched && matched &&
			(sourceLine.Status == lines.LineWithoutStatus || sourceLine.Status == lines.LineDimmed) {

			newMatched = true
		}
	case FilterHide:
		// Status
		if matched && indeces[0][1] != 0 {
			newStatus = lines.LineHidden
		}

		// Matched
		if sourceLine.Matched && matched &&
			(sourceLine.Status == lines.LineMatched || sourceLine.Status == lines.LineDimmed) {
			newMatched = false
		}
	default:
		log.Panicf("Unkwon filter mdoe %d", s.mode)
	}

	sourceLine.Status = newStatus
	sourceLine.Matched = newMatched
}

func (s *StringFilter) colorizeLine(line *lines.Line, indeces [][]int) {
	for _, index := range indeces {
		for i := index[0]; i < index[1]; i++ {
			line.ColorIndex[i] = s.colorIndex
		}
	}
}

func (s *StringFilter) SetColorIndex(colorIndex uint8) {
	s.colorIndex = colorIndex
}
