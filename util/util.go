package util

import (
	"math"
)

func InBetween(i, min, max int) (int, error) {
	if i < min {
		return min, ErrNotInBetween
	} else if i > max {
		return max, ErrNotInBetween
	} else {
		return i, nil
	}
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

func Contains[T comparable](array []T, value T) bool {
	for _, elem := range array {
		if elem == value {
			return true
		}
	}
	return false
}

// returns true if element was present before
func Remove[T comparable](array []T, value T) bool {
	for i, elem := range array {
		if elem == value {
			array = append(array[:i], array[i+1:]...)
			return true
		}
	}
	return false
}
