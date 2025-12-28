package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/devathh/coderun/xcutr-service/internal/app"
)

func main() {
	app, err := app.New()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	go func() {
		if err := app.Start(); err != nil {
			slog.Error(err.Error())
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT)
	<-stop

	app.Shutdown()
}
