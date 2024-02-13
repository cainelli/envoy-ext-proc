package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/cainelli/ext-proc/pkg/echo"
	setcookie "github.com/cainelli/ext-proc/pkg/processors/set-cookie"
	"github.com/cainelli/ext-proc/pkg/server"
	"github.com/cainelli/ext-proc/pkg/service"
	"github.com/cainelli/ext-proc/pkg/service/processor"
)

func main() {
	extProc := &service.ExtProcessor{
		Processors: []processor.Processor{
			&setcookie.SetCookieProcessor{},
		},
	}
	grpcSrv := server.NewExtProcServer(extProc)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	go func() {
		<-ctx.Done()
		log.Print("shutting down...")
		grpcSrv.Stop()
	}()

	http.HandleFunc("/headers", echo.RequestHeadersHandler)
	http.HandleFunc("/response-headers", echo.ResponseHeadersHandler)
	go func() {
		slog.Info("starting HTTP server", "port", ":8000")
		if err := http.ListenAndServe(":8000", nil); err != nil {
			slog.Error("could not listen http", "error", err)
			cancel()
		}
	}()

	if err := grpcSrv.Run(":9000"); err != nil {
		slog.Error("could not listen grpc", "error", err)
		cancel()
	}
}
