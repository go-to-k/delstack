package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/go-to-k/delstack/app"
	"github.com/go-to-k/delstack/logger"
	"github.com/go-to-k/delstack/option"
)

func main() {
	logger.NewLogger()
	ctx := context.TODO()
	app := app.NewApp(getVersion())

	if err := app.Run(ctx); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
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
