package formats

import (
	"bytes"
	"io"
	"os"
	"regexp"

	"github.com/claude42/infiltrator/config"
	"github.com/claude42/infiltrator/fail"
	"github.com/claude42/infiltrator/model/lines"
)

const testCases = 100

func Identify(lines []*lines.Line) {
	formats := config.Formats()
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
		config.User().FileFormat = fileFormat
		config.User().FileFormatRegex = regexs[fileFormat]
	}
}

func IsGzip(file *os.File) (bool, error) {
	var gzipMagicNumber = []byte{0x1f, 0x8b}

	// Read the first two bytes of the file.
	magicBytes := make([]byte, 2)
	n, err := file.Read(magicBytes)
	if err != nil {
		return false, err
	}

	_, err = file.Seek(0, io.SeekStart)
	fail.OnError(err, "Seek() failed")

	// Check if the two bytes match the gzip magic number.
	return n == 2 && bytes.Equal(magicBytes, gzipMagicNumber), nil
}
