package filter

import (
	"log"
	"sort"
	"time"

	"github.com/claude42/infiltrator/model/reader"

	dateparser "github.com/markusmobius/go-dateparser"
)

type DateFilter struct {
	FilterImpl
	startLineNo int
	endLineNo   int
}

func NewDateFilter() *DateFilter {
	d := &DateFilter{}

	return d
}

func (d *DateFilter) GetLine(lineNo int) (*reader.Line, error) {
	sourceLine, err := d.source.GetLine(lineNo)
	if err != nil {
		return sourceLine, err
	}

	if sourceLine.Status == reader.LineHidden {
		return sourceLine, nil
	}

	if d.startLineNo == d.endLineNo {
		return sourceLine, nil
	}

	if sourceLine.No < d.startLineNo || sourceLine.No > d.endLineNo {
		sourceLine.Status = reader.LineHidden
	}

	return sourceLine, nil
}

// func (d *DateFilter) GetLine(lineNo int) (*reader.Line, error) {
// 	sourceLine, err := d.source.GetLine(lineNo)
// 	if err != nil {
// 		return sourceLine, err
// 	}

// 	if sourceLine.Status == reader.LineHidden {
// 		return sourceLine, nil
// 	}

// 	if !sourceLine.When.IsZero() {
// 		return sourceLine, nil
// 	}

// 	_, err = d.getDate(sourceLine)
// 	if err != nil {
// 		// log.Printf("Date parsing error %T, %s", err, x)
// 		log.Printf("Date parsing error %T", err)
// 		return sourceLine, nil
// 	}
// 	log.Printf("Parsed line %d", sourceLine.No)

// 	return sourceLine, nil
// }

func (d *DateFilter) SetStart(startTime time.Time) {
	d.startLineNo = d.findFirstAfter(startTime)
}

func (d *DateFilter) SetEnd(endTime time.Time) {
	d.endLineNo = d.findLastBefore(endTime)
}

func (d *DateFilter) findFirstAfter(startTime time.Time) int {
	lineNo := sort.Search(d.Length(), func(lineNo int) bool {
		lineTime, err := d.getDateForLineNo(lineNo)
		if err != nil {
			// TODO error handling
			log.Panicf("Paaanik %T", err)
		}
		return startTime.Before(lineTime)
	})
	// TODO error handling
	return lineNo
}

func (d *DateFilter) findLastBefore(endTime time.Time) int {
	lineNo := sort.Search(d.source.Length(), func(lineNo int) bool {
		lineTime, err := d.getDateForLineNo(lineNo)
		if err != nil {
			// TODO error handling
			log.Panicf("Paaanik %T", err)
		}
		return endTime.After(lineTime)
	})
	return lineNo
}

func (d *DateFilter) getDate(line *reader.Line) (time.Time, error) {
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

	return d.getDate(line)
}
