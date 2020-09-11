package main

import (
	"kbridge/internal/controller"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {

	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	zap.L().Debug("Starting")
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	c := &controller.Config{
		HTTPListen: ":8080",
		NonceLife:  120,
	}

	controller := controller.NewController(c)

	controller.GetKeytabStore().AddPrincipal("superman@EXAMPLE.COM")
	controller.GetKeytabStore().UpdateNow()

	<-sig

	zap.L().Debug("Shutting Down")

	controller.Shutdown()

}
