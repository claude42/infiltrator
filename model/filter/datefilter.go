package filter

import (
	"log"
	"math"
	"sort"
	"time"

	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/lines"

	dateparser "github.com/markusmobius/go-dateparser"
)

const (
	DateFilterFrom = "From"
	DateFilterTo   = "To"
)

type DateFilter struct {
	FilterImpl
	fromLineNo int
	toLineNo   int
}

func NewDateFilter() *DateFilter {
	d := &DateFilter{
		toLineNo: math.MaxInt,
	}

	return d
}

func (d *DateFilter) SetKey(name string, key string) error {
	keyTime, err := dateparser.Parse(nil, key)
	if err != nil {
		// TODO: error handling
		if name == DateFilterFrom {
			d.fromLineNo = 0
		} else if name == DateFilterTo {
			d.toLineNo = math.MaxInt
		}
		return err
	}

	if name == DateFilterFrom {
		d.fromLineNo = d.findFirstAfter(keyTime.Time)
	} else if name == DateFilterTo {
		d.toLineNo = d.findLastBefore(keyTime.Time)
	} else {
		log.Panicf("Neither from nor to but '%s'", name)
	}

	return nil
}

func (d *DateFilter) GetLine(lineNo int) (*lines.Line, error) {
	sourceLine, err := d.source.GetLine(lineNo)
	if err != nil {
		return sourceLine, err
	}

	if sourceLine.Status == lines.LineHidden {
		return sourceLine, nil
	}

	if d.fromLineNo == d.toLineNo {
		return sourceLine, nil
	}

	if sourceLine.No < d.fromLineNo || sourceLine.No > d.toLineNo {
		sourceLine.Status = lines.LineHidden
	}

	return sourceLine, nil
}

func (d *DateFilter) findFirstAfter(fromTime time.Time) int {
	lineNo := sort.Search(d.Length(), func(lineNo int) bool {
		lineTime, err := d.getDateForLineNo(lineNo)
		// TODO error handling
		fail.OnError(err, "Paaanik")
		return fromTime.Before(lineTime)
	})
	// TODO error handling
	return lineNo
}

func (d *DateFilter) findLastBefore(toTime time.Time) int {
	lineNo := sort.Search(d.source.Length(), func(lineNo int) bool {
		lineTime, err := d.getDateForLineNo(lineNo)
		// TODO error handling
		fail.OnError(err, "Paaanik")
		return lineTime.After(toTime)
	})
	return lineNo - 1
}

func (d *DateFilter) calculateLineDate(line *lines.Line) (time.Time, error) {
	if !line.When.IsZero() {
		return line.When, nil
	}

	// x, results, err := dateparser.Search(nil, line.Str)
	lineTime, err := dateparser.Parse(nil, line.Str[:15], "Jan 02 03:04:05")
	if err != nil {
		// TODO: error handling
		return time.Time{}, err
	}

	// line.When = results[0].Date.Time
	line.When = lineTime.Time
	return lineTime.Time, nil
}

func (d *DateFilter) getDateForLineNo(lineNo int) (time.Time, error) {
	line, err := d.GetLine(lineNo)
	if err != nil {
		// TODO: error handling
		return time.Time{}, err
	}

	return d.calculateLineDate(line)
}

func (s *DateFilter) SetColorIndex(colorIndex uint8) {
	// don't care, filter won't highlight anything
}
