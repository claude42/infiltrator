package model

import (
	"log"

	"github.com/claude42/infiltrator/util"
)

type Source struct {
	lines []Line
	width int
}

func (s *Source) calculateNewWidthFrom(start int) {
	for _, line := range s.lines[start:] {
		s.width = util.IntMax(s.width, len(line.Str))
	}
}

func (s *Source) storeNewLines(newLines *[]Line) int {
	start := len(s.lines)
	s.lines = append(s.lines, *newLines...)
	s.calculateNewWidthFrom(start)
	return len(s.lines)
}

func (s *Source) size() (int, int) {
	return s.width, len(s.lines)
}

func (s *Source) length() int {
	return len(s.lines)
}

func (s *Source) getLine(line int) (Line, error) {
	length := len(s.lines)

	if line < 0 || line >= length {
		return Line{Str: "ErrOutOfBounds"}, util.ErrOutOfBounds
	}

	s.lines[line].CleanUp()

	return s.lines[line], nil
}

func (s *Source) setSource(source Filter) {
	log.Panicln("SetSource() should never be called on a source!")
}

func (s *Source) setKey(key string) error {
	log.Panicln("SetKey() should never be called on a source!")
	return nil
}

func (s *Source) setMode(mode FilterMode) {
	log.Panicln("SetMode() should never be called on a source!")
}

func (s *Source) setCaseSensitive(caseSensitive bool) error {
	log.Panicln("SetCaseSensitive() should never be called on a source!")
	return nil
}

func (s *Source) setColorIndex(colorIndex uint8) {
	// Buffers don't have a color
}

func (s *Source) isEmpty() bool {
	return len(s.lines) == 0
}

func (s *Source) LastLine() Line {
	return s.lines[len(s.lines)-1]
}
