package main

import (
	"context"
	"os"
	"runtime/debug"

	"github.com/go-to-k/delstack/app"
	"github.com/go-to-k/delstack/logger"
)

var version = ""

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
	if version != "" {
		return version
	}
	i, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	return i.Main.Version
}
