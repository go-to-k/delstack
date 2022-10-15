package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/go-to-k/delstack/internal/app"
	"github.com/go-to-k/delstack/internal/option"
	"github.com/go-to-k/delstack/pkg/logger"
)

func main() {
	logger.NewLogger(isDebug())
	ctx := context.TODO()
	app := app.NewApp(getVersion())

	if err := app.Run(ctx); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
}

func isDebug() bool {
	if option.Version == "" || option.Revision != "" {
		return true
	}
	return false
}

func getVersion() string {
	if option.Version != "" && option.Revision != "" {
		return option.Version + "-" + option.Revision
	}
	if option.Version != "" {
		return option.Version
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return i.Main.Version
}
