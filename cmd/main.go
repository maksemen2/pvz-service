package main

import (
	"context"
	"github.com/maksemen2/pvz-service/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/maksemen2/pvz-service/internal/app"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()

	if err != nil {
		panic("configuration load failed: " + err.Error())
	}

	application, err := app.Initialize(cfg)
	if err != nil {
		panic("failed to initialize application" + err.Error())
	}

	servers, err := application.StartServers()
	if err != nil {
		application.Logger.Fatal("Failed to start servers", zap.Error(err))
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	cancelTimeout := 5 * time.Second

	ctx, cancel := context.WithTimeout(context.Background(), cancelTimeout)
	defer cancel()

	servers.Stop(ctx)

	application.Logger.Info("Application stopped")
}
