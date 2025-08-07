package model

type Filter interface {
	getLine(line int) (*Line, error)
	setSource(source Filter)
	size() (int, int)
	length() int

	setKey(key string) error
	setColorIndex(colorIndex uint8)
	setMode(mode FilterMode)
	setCaseSensitive(on bool) error
}
