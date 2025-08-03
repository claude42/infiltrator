package model

type Filter interface {
	getLine(line int) (Line, error)
	setSource(source Filter)
	size() (int, int, error)

	setKey(key string) error
	setColorIndex(colorIndex uint8)
	setMode(mode int)
	setCaseSensitive(on bool) error
}
