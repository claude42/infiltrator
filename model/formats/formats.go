package formats

import (
	"log"
	"regexp"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/model/reader"
)

const testCases = 100

func Identify(lines []*reader.Line) {
	formats := config.GetConfiguration().Formats
	regexs := make(map[string]*regexp.Regexp)
	results := make(map[string]int, len(formats))

	for key := range formats {
		regexs[key] = regexp.MustCompile(formats[key])
		results[key] = 0
	}

	n := min(testCases, len(lines))

nextLine:
	for i := range n {
		for format, regex := range regexs {
			if regex.MatchString(lines[i].Str) {
				results[format]++
				continue nextLine
			} else {
				log.Printf("Raw: %q", lines[i].Str)
			}

		}
	}

	var maxResult int
	var fileFormat string
	for format, result := range results {
		if result > maxResult {
			maxResult = result
			fileFormat = format
		}
	}

	if float64(maxResult)/float64(n) > 0.9 {
		config.GetConfiguration().FileFormat = fileFormat
		config.GetConfiguration().FileFormatRegex = regexs[fileFormat]
	}
}
