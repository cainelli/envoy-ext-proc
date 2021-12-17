package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cainelli/ext-proc/pkg/extproc"
	"github.com/cainelli/ext-proc/pkg/server"
)

func main() {
	extProc := &extproc.ExtProcessor{}
	srv := server.NewExtProcServer(extProc)

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		log.Print("shutting down...")
		srv.Stop()
	}()

	srv.Run(":9000")
}
