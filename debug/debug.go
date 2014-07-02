package debug

import (
	"log"
	"os"
)

var Enable = false

var logger = log.New(os.Stderr, "", log.LstdFlags)

func Log(message string) {
	if Enable {
		logger.Printf(message)
	}
}

func Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v)
}
