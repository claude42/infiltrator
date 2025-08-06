package model

import (
	"fmt"
	"regexp"
	// "github.com/claude42/infiltrator/util"
)

func RegexFilterFuncFactory(key string, caseSensitive bool) (func(input string) (string, [][]int, bool), error) {
	// don't use regex if we don't have to
	var specialRegexChars = regexp.MustCompile(`[.^$|?*+(){}\[\]\\]`)
	if !specialRegexChars.MatchString(key) {
		return DefaultStringFilterFuncFactory(key, caseSensitive)
	}

	if !caseSensitive {
		key = fmt.Sprintf("(?i)%s", key)
	}

	re, err := regexp.Compile(key)
	if err != nil {
		return nil, err
	}

	return func(input string) (string, [][]int, bool) {
		indeces := re.FindAllStringIndex(input, -1)
		if indeces == nil {
			return "", indeces, false
		}

		return input, indeces, true
	}, nil
}
