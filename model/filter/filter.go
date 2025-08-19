package filter

import (
	"log"

	"github.com/claude42/infiltrator/model/lines"
)

type Filter interface {
	GetLine(line int) (*lines.Line, error)
	SetSource(source Filter)
	Size() (int, int)
	Length() int

	SetKey(name string, key string) error
	SetColorIndex(colorIndex uint8)
}

type FilterImpl struct {
	source     Filter
	colorIndex uint8
}

func (f *FilterImpl) GetLine(lineNo int) (*lines.Line, error) {
	sourceLine, err := f.source.GetLine(lineNo)
	if err != nil {
		return sourceLine, err
	}

	return sourceLine, nil
}

func (f *FilterImpl) SetSource(source Filter) {
	f.source = source
}

func (f *FilterImpl) Size() (int, int) {
	return f.source.Size()
}

func (f *FilterImpl) Length() int {
	return f.source.Length()
}

func (f *FilterImpl) SetColorIndex(colorIndex uint8) {
	f.colorIndex = colorIndex
}

func (f *FilterImpl) SetKey(name string, key string) error {
	log.Panicln("SetKey() not implemented!")
	return nil
}
