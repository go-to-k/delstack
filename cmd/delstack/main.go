package main

import (
	"context"
	"os"

	"github.com/go-to-k/delstack/internal/app"
	"github.com/go-to-k/delstack/internal/version"
	"github.com/go-to-k/delstack/pkg/logger"
)

func main() {
	logger.NewLogger(version.IsDebug())
	ctx := context.TODO()
	app := app.NewApp(version.GetVersion())

	if err := app.Run(ctx); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
}
