package extproc

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	cfgcorev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExtProcessor struct{}

var _ extproc.ExternalProcessorServer = &ExtProcessor{}

func (x *ExtProcessor) Process(procSrv extproc.ExternalProcessor_ProcessServer) error {
	ctx := procSrv.Context()

	reqHeaders := http.Header{}
	resHeaders := http.Header{}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			req, err := procSrv.Recv()
			if errors.Is(err, io.EOF) || errors.Is(err, status.Error(codes.Canceled, context.Canceled.Error())) {
				return nil
			} else if err != nil {
				log.Println("an error occured while processing the requets:", err)
				return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
			}
			switch req.Request.(type) {
			// Process Request Headers
			case *extproc.ProcessingRequest_RequestHeaders:
				if err := x.processingRequestHeaders(procSrv, req, reqHeaders); err != nil {
					return status.Errorf(codes.Unknown, "cannot process request headers: %+v", err)
				}
			// Process Requesponse Headers
			case *extproc.ProcessingRequest_ResponseHeaders:
				if err := x.processingResponseHeaders(procSrv, req, reqHeaders, resHeaders); err != nil {
					return status.Errorf(codes.Unknown, "cannot process response headers: %+v", err)
				}
			// Process Request Body
			case *extproc.ProcessingRequest_RequestBody:
				if err := x.processingRequestBody(procSrv); err != nil {
					return status.Errorf(codes.Unknown, "cannot process request body: %+v", err)
				}
			// Process Response Body
			case *extproc.ProcessingRequest_ResponseBody:
				if err := x.processingResponseBody(procSrv); err != nil {
					return status.Errorf(codes.Unknown, "cannot process response body: %+v", err)
				}
			// Process Request Trailers
			case *extproc.ProcessingRequest_RequestTrailers:
				if err := x.processingRequestTrailers(procSrv); err != nil {
					return status.Errorf(codes.Unknown, "cannot process request trailers: %+v", err)
				}
			// Process Response Trailers
			case *extproc.ProcessingRequest_ResponseTrailers:
				if err := x.processingResponseTrailers(procSrv); err != nil {
					return status.Errorf(codes.Unknown, "cannot process response trailers: %+v", err)
				}
			default:
				log.Println("WARNING: Unknown request type", req)
				return status.Errorf(codes.Unknown, "unknown request type")
			}
		}
	}

}

func (x *ExtProcessor) processingRequestHeaders(procSrv extproc.ExternalProcessor_ProcessServer, req *extproc.ProcessingRequest, reqHeaders http.Header) error {
	for _, h := range req.GetRequestHeaders().GetHeaders().GetHeaders() {
		reqHeaders.Set(h.Key, h.Value)
	}
	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extproc.HeadersResponse{
				Response: &extproc.CommonResponse{
					HeaderMutation: &extproc.HeaderMutation{
						SetHeaders: []*cfgcorev3.HeaderValueOption{
							{
								Header: &cfgcorev3.HeaderValue{
									Key:   "x-custom-header",
									Value: "ok",
								},
							},
						},
					},
				},
			},
		},
	})

	return nil
}

func (x *ExtProcessor) processingResponseHeaders(procSrv extproc.ExternalProcessor_ProcessServer, req *extproc.ProcessingRequest, reqHeaders, resHeaders http.Header) error {
	for _, h := range req.GetResponseHeaders().GetHeaders().GetHeaders() {
		resHeaders.Set(h.Key, h.Value)
	}

	log.Printf("reqHeaders:%+v\n", reqHeaders)
	log.Printf("resHeaders:%+v\n", resHeaders)

	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extproc.HeadersResponse{
				Response: &extproc.CommonResponse{
					HeaderMutation: &extproc.HeaderMutation{
						SetHeaders: []*cfgcorev3.HeaderValueOption{
							{
								Header: &cfgcorev3.HeaderValue{
									Key:   "x-request-id",
									Value: reqHeaders.Get("x-request-id"),
								},
							},
						},
					},
				},
			},
		},
	})

	return nil
}

func (x *ExtProcessor) processingRequestBody(procSrv extproc.ExternalProcessor_ProcessServer) error {
	log.Println("not implemented: request body")

	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestBody{},
	})

	return nil
}

func (x *ExtProcessor) processingResponseBody(procSrv extproc.ExternalProcessor_ProcessServer) error {
	log.Println("not implemented: response body")

	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseBody{},
	})

	return nil
}

func (x *ExtProcessor) processingRequestTrailers(procSrv extproc.ExternalProcessor_ProcessServer) error {
	log.Println("not implemented: trailers request")

	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestTrailers{},
	})
	return nil
}

func (x *ExtProcessor) processingResponseTrailers(procSrv extproc.ExternalProcessor_ProcessServer) error {
	log.Println("not implemented: trailers response")

	procSrv.Send(&extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseTrailers{},
	})
	return nil
}
