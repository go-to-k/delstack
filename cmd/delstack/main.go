package main

import (
	"context"
	"os"

	"github.com/go-to-k/delstack/internal/app"
	"github.com/go-to-k/delstack/internal/option"
	"github.com/go-to-k/delstack/pkg/logger"
)

func main() {
	logger.NewLogger(option.IsDebug())
	ctx := context.TODO()
	app := app.NewApp(option.GetVersion())

	if err := app.Run(ctx); err != nil {
		logger.Logger.Error().Msg(err.Error())
		os.Exit(1)
	}
}
