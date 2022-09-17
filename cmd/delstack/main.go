package main

import (
	"context"
	"os"

	"github.com/go-to-k/delstack/app"
	"github.com/go-to-k/delstack/logger"
)

var version = "version"

func main() {
	logger.NewLogger()
	ctx := context.TODO()
	app := app.NewApp(version)

	if err := app.Run(ctx); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
}
