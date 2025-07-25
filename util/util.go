package util

import (
	"errors"
	"fmt"

	// "log"
	"math"
)

var (
	ErrOutOfBounds     = errors.New("out of bounds")
	ErrLineDidNotMatch = errors.New("line did not match")
)

func IntMax(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func IntMin(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}

func InBetween(i, min, max int) (int, error) {
	if i < min {
		return min, &NotInBetweenError{}
	} else if i > max {
		return max, &NotInBetweenError{}
	} else {
		return i, nil
	}
}

type NotInBetweenError struct {
}

func (e *NotInBetweenError) Error() string {
	return fmt.Sprintf("Value not in between error")
}

func CountDigits(i int) int {
	if i == 0 {
		return 1
	}

	return int(math.Floor(math.Log10(float64(i)))) + 1
}

func InsertRune(runes []rune, r rune, index int) ([]rune, error) {
	if index < 0 || index > len(runes) {
		return nil, ErrOutOfBounds
	}

	result := make([]rune, len(runes)+1)

	copy(result[:index], runes[:index])
	result[index] = r
	copy(result[index+1:], runes[index:])

	return result, nil
}
