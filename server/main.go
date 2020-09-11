package server

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

// Main ...
func Main() {

	logger, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(logger)

	listen := os.Getenv("LISTEN")

	if listen == "" {
		listen = ":8080"
		zap.L().Debug("Listening on :8080 (default). Set LISTEN to override. For example LISTEN=:8081")
	} else {
		zap.L().Debug(fmt.Sprintf("Listening on %s", listen))
	}

	zap.L().Debug("Starting")
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	searchServer := NewServer(listen)

	<-sig

	searchServer.Shutdown()

	zap.L().Debug("Shutting Down")

}
