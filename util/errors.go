package util

import (
	"errors"
)

var (
	ErrOutOfBounds     = errors.New("out of bounds")
	ErrLineDidNotMatch = errors.New("line did not match")
	ErrNotInBetween    = errors.New("number not in between the two values")
	ErrNotFound        = errors.New("no (further) match found")
)
