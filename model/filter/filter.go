package filter

import (
	"log"

	"github.com/claude42/infiltrator/model/reader"
)

type Filter interface {
	GetLine(line int) (*reader.Line, error)
	SetSource(source Filter)
	Size() (int, int)
	Length() int

	SetKey(key string) error
	SetColorIndex(colorIndex uint8)
	SetMode(mode FilterMode)
	SetCaseSensitive(on bool) error
}

type FilterImpl struct {
	source Filter
}

func (f *FilterImpl) GetLine(lineNo int) (*reader.Line, error) {
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

func (f *FilterImpl) SetKey(key string) error {
	log.Panicln("SetKey() not implemented!")
	return nil
}

func (f *FilterImpl) SetMode(mode FilterMode) {
	log.Panicln("SetMode() not implemented!")
}

func (f *FilterImpl) SetCaseSensitive(caseSensitive bool) error {
	log.Panicln("SetCaseSensitive() not implemented!")
	return nil
}

func (f *FilterImpl) SetColorIndex(colorIndex uint8) {
	log.Panicln("setColorIndex() not implemented!")
}
