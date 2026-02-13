package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hytonhan/certwatch/internal/app"
	"github.com/hytonhan/certwatch/internal/audit"
	"github.com/hytonhan/certwatch/internal/config"
)

func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	logger := audit.NewLogger()
	conf := config.New()

	app, aerr := app.New(conf)
	if aerr != nil {
		logger.Warn("error occured")
		log.Fatal(aerr)
	}
	if err := app.Run(ctx); err != nil {
		logger.Warn("error occured")
		log.Fatal(err)
	}
}
