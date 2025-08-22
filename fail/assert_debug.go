//go:build debug

package fail

import "log"

func Assert(condition bool, message string, v ...any) {
	if !condition {
		log.Panicf(message, v...)
	}
}
