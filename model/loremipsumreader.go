package model

import (
	"bufio"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/claude42/infiltrator/model/reader"
)

var (
	loremIpsumInstance *LoremIpsumReader
	loremIpsumOnce     sync.Once
)

type LoremIpsumReader struct {
	sync.Mutex

	newLines      []reader.Line
	lineNo        int
	contentUpdate chan<- []reader.Line
}

func GetLoremIpsumReader() *LoremIpsumReader {
	loremIpsumOnce.Do(func() {
		loremIpsumInstance = &LoremIpsumReader{}
	})
	return loremIpsumInstance
}

func (c *LoremIpsumReader) Read(ch chan<- []reader.Line) {
	text := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed sed facilisis
ligula, non finibus erat. Sed viverra leo elit, quis posuere magna sagittis
id. Morbi eget dolor sem. Nulla consequat dignissim velit vitae
consectetur. Vestibulum sit amet nulla vitae mi finibus pharetra. Etiam
egestas magna at bibendum scelerisque. Proin ornare nisi non pretium
porttitor. Quisque hendrerit faucibus sapien sit amet ullamcorper. Vivamus
pretium elit sed ipsum sodales, at vehicula tortor euismod. Pellentesque
consectetur enim at dolor consectetur sollicitudin.

Nulla in quam at quam sodales viverra. Nullam aliquet in lacus sed
condimentum. Suspendisse tristique nec dui sed porta. Fusce imperdiet
maximus euismod. Nam sed magna mauris. Proin placerat metus sapien, vel
sodales eros ultricies sit amet. Cras nec arcu neque. Ut dictum, felis nec
convallis porta, risus tellus accumsan orci, in lacinia est nisl nec neque.
Sed consectetur luctus dolor, sit amet finibus sapien luctus nec. Donec
tellus tortor, placerat ac lacus at, dictum fermentum leo. Mauris pretium
lacus ut imperdiet tristique.

Nunc fringilla luctus ornare. Integer suscipit vel ligula non eleifend.
Phasellus efficitur ipsum sed ipsum aliquam, bibendum blandit velit
porttitor. Maecenas ac neque nunc. Nullam malesuada augue a nibh pulvinar
faucibus. Etiam ac diam pharetra, accumsan massa ac, vestibulum nulla. In
hac habitasse platea dictumst. Nullam vestibulum imperdiet lacus, ut ornare
enim hendrerit fringilla. Nulla a dolor mi.

Sed euismod turpis id purus pharetra, quis sodales metus vehicula. Donec
feugiat turpis vel massa mattis euismod. Aenean tempus nec dui a aliquet.
Integer ullamcorper ante et justo cursus eleifend. Mauris id justo sed odio
pellentesque volutpat ut sit amet tellus. Integer eget malesuada turpis,
eget iaculis neque. Sed pulvinar, nisl eu semper vehicula, libero nisl
sagittis felis, a pulvinar tortor eros vitae lectus. Nullam iaculis pretium
risus, eleifend pulvinar lacus tristique placerat. Nam sit amet tellus et
sapien laoreet euismod. Cras vitae finibus felis, sit amet vehicula lacus.
Pellentesque tincidunt tellus sem, a placerat odio pulvinar quis. Aliquam
erat volutpat. Nunc iaculis sollicitudin maximus. Aliquam vitae metus
rhoncus velit egestas blandit.

Fusce pulvinar lorem nunc, a volutpat eros lobortis sed. Nunc dapibus felis
quis feugiat mollis. In efficitur dictum suscipit. Duis a suscipit lorem.
Morbi suscipit viverra tortor, ut consectetur erat elementum non. Cras
pulvinar purus at justo consectetur, nec aliquet sapien hendrerit.
Pellentesque habitant morbi tristique senectus et netus et malesuada fames
ac turpis egestas.`

	c.contentUpdate = ch

	reader := strings.NewReader((text))

	err := c.readNewLines(reader)
	if err != nil {
		log.Println(err)
	}
}

func (c *LoremIpsumReader) readNewLines(rd *strings.Reader) error {
	c.newLines = c.newLines[:0]
	scanner := bufio.NewScanner(rd)
	for scanner.Scan() {
		text := scanner.Text()
		c.newLines = append(c.newLines, *reader.NewLine(c.lineNo, text))
		c.lineNo++
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	c.contentUpdate <- c.newLines

	return nil
}
