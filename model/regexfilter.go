package model

import (
	"regexp"

	"github.com/claude42/infiltrator/util"
)

func RegexFilterFuncFactory(key string) (func(input string) (string, [][]int, error), error) {
	re, err := regexp.Compile(key)
	if err != nil {
		return nil, err
	}

	return func(input string) (string, [][]int, error) {
		indeces := re.FindAllStringIndex(input, -1)
		if indeces == nil {
			return "", indeces, util.ErrLineDidNotMatch
		}

		return input, indeces, nil
	}, nil
}
