package main

import (
	"os"
	"os/signal"

	apps "github.com/MamangRust/monolith-payment-gateway-apigateway/internal/app"
	"go.uber.org/zap"
)

func main() {
	client, shutdown, err := apps.RunClient()

	if err != nil {
		client.Logger.Fatal("failed to run client: %v", zap.Error(err))
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	client.Logger.Info("Gracefully shutting down...")
	shutdown()
}
