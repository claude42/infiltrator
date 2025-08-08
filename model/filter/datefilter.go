package filter

import (
	"github.com/claude42/infiltrator/model/reader"
)

type DateFilter struct {
	FilterImpl
}

func (d *DateFilter) GetLine(lineNo int) (*reader.Line, error) {
	sourceLine, err := d.source.GetLine(lineNo)
	if err != nil {
		return sourceLine, err
	}

	return sourceLine, nil
}
