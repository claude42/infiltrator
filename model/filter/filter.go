package filter

import "github.com/claude42/infiltrator/model/reader"

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
