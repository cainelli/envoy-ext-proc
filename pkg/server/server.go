package server

import (
	"fmt"
	"log/slog"
	"net"

	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	"google.golang.org/grpc"
)

type ExtProcServer struct {
	grpcServer *grpc.Server
	extProc    extproc.ExternalProcessorServer
}

func NewExtProcServer(extProcServerv3 extproc.ExternalProcessorServer) *ExtProcServer {
	return &ExtProcServer{
		extProc: extProcServerv3,
	}
}

func (extProcSrv *ExtProcServer) Run(grpcAddr string) error {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	extProcSrv.grpcServer = grpc.NewServer()
	extproc.RegisterExternalProcessorServer(extProcSrv.grpcServer, extProcSrv.extProc)

	slog.Info("starting gRPC server", "port", grpcAddr)
	if err := extProcSrv.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

func (extProcSrv *ExtProcServer) Stop() {
	extProcSrv.grpcServer.GracefulStop()
}
