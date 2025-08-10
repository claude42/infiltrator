package fail

import (
	"log"
)

func OnError(err error, message string) {
	IfNotNil(err, message)
}

func If(condition bool, message string, v ...any) {
	if condition {
		log.Panicf(message, v...)
	}
}

func Unless(condition bool, message string, v ...any) {
	If(!condition, message, v...)
}

func IfNil(x any, message string) {
	if x == nil {
		log.Panicf("%s: %+v", message, x)
	}
}

func IfNotNil(x any, message string) {
	if x != nil {
		log.Panicf("%s: %+v", message, x)
	}
}

// TODO: conditional compile

func Assert(condition bool, message string, v ...any) {
	if !condition {
		log.Panicf(message, v...)
	}
}

func Must0(err error) {
	if err != nil {
		log.Panicf("%+v", err)
	}
}

func Must1[T any](x T, err error) T {
	if err != nil {
		log.Panicf("%+v", err)
	}
	return x
}
