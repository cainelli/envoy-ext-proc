package server

import (
	"log"
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

func (extProcSrv *ExtProcServer) Run(grpcAddr string) {
	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}

	extProcSrv.grpcServer = grpc.NewServer()
	extproc.RegisterExternalProcessorServer(extProcSrv.grpcServer, extProcSrv.extProc)

	log.Printf("starting gRPC server at %s", listener.Addr())
	if err := extProcSrv.grpcServer.Serve(listener); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
		return
	}
}

func (extProcSrv *ExtProcServer) Stop() {
	extProcSrv.grpcServer.GracefulStop()
}
