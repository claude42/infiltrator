package filter

import (
	"log"

	"github.com/claude42/infiltrator/model/reader"
	"github.com/claude42/infiltrator/util"
)

type Source struct {
	FilterImpl
	lines []*reader.Line
	width int
}

func NewSource() *Source {
	s := &Source{}

	return s
}

func (s *Source) calculateNewWidthFrom(start int) {
	for _, line := range s.lines[start:] {
		s.width = max(s.width, len(line.Str))
	}
}

func (s *Source) StoreNewLines(newLines []*reader.Line) int {
	start := len(s.lines)
	s.lines = append(s.lines, newLines...)
	s.calculateNewWidthFrom(start)
	return len(s.lines)
}

func (s *Source) Size() (int, int) {
	return s.width, len(s.lines)
}

func (s *Source) Length() int {
	return len(s.lines)
}

func (s *Source) GetLine(line int) (*reader.Line, error) {
	length := len(s.lines)

	if line < 0 || line >= length {
		return reader.NonExistingLine, util.ErrOutOfBounds
	}

	s.lines[line].CleanUp()

	return s.lines[line], nil
}

func (s *Source) SetSource(source Filter) {
	log.Panicln("SetSource() should never be called on a source!")
}

func (s *Source) SetColorIndex(colorIndex uint8) {
	// Buffers don't have a color
}

func (s *Source) IsEmpty() bool {
	return len(s.lines) == 0
}

func (s *Source) LastLine() *reader.Line {
	return s.lines[len(s.lines)-1]
}
