package model

// "sync"

type Display struct {
	// this is the buffer that should be rendered on screen. It can potentially
	// happen that its size is out of sync with the actual screen size then
	// its data should be ignored until there's an updated version with the
	// correct dimensions
	Buffer []Line

	// at what percentage of the whole buffer are we currently
	// TODO: decide: display percentag in relation to whole file or to the
	// filters' output. Latter option obviously will be very expensive.
	Percentage int

	// indicates whether an error has occured that should be indicated with a
	// beep
	Beep bool

	ErrorMessage string

	// the following parameters reference the source (or lines in the source) -
	// NOT the screen buffer
	TotalLength  int
	CurrentMatch int
}

func (d *Display) Height() int {
	height := len(d.Buffer)
	return height
}

func (d *Display) SetHeight(height int) {
	currentHeight := len(d.Buffer)
	if height < currentHeight {
		d.Buffer = d.Buffer[:height]
	} else if height > currentHeight {
		d.Buffer = append(d.Buffer, make([]Line, height-currentHeight)...)
	}
}

func (d *Display) SetTotalLength(length int) {
	d.TotalLength = length
}
