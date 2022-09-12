package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var Logger *zerolog.Logger

func NewLogger() {
	consoleWriter := zerolog.ConsoleWriter{
		Out: os.Stderr,
		PartsExclude: []string{
			zerolog.TimestampFieldName,
		},
	}

	l := zerolog.New(&consoleWriter)

	Logger = &l
}
