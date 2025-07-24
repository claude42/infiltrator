package model

import (
	"regexp"
)

func RegexFilterFuncFactory(key string) (func(input string) (string, [][]int, error), error) {
	re, err := regexp.Compile(key)
	if err != nil {
		return nil, err
	}

	return func(input string) (string, [][]int, error) {
		if key == "" {
			return input, [][]int{}, nil
		}

		indeces := re.FindAllStringIndex(input, -1)
		if indeces == nil {
			return "", indeces, ErrLineDidNotMatch
		}

		return input, indeces, nil
	}, nil
}
