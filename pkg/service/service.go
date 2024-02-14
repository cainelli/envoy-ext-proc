package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/cainelli/ext-proc/pkg/service/processor"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ExtProcessor struct {
	Processors []processor.Processor
}

var _ extproc.ExternalProcessorServer = &ExtProcessor{}

// Process is the main entry point for the ExternalProcessor service.
// The protocol itself is based on a bidirectional gRPC stream. Envoy will send the server ProcessingRequest messages, and the server must reply with ProcessingResponse.
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ext_proc/v3/ext_proc.proto#envoy-v3-api-msg-extensions-filters-http-ext-proc-v3-externalprocessor
func (svc *ExtProcessor) Process(procsrv extproc.ExternalProcessor_ProcessServer) error {
	ctx := procsrv.Context()
	req := &processor.RequestContext{}
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		procreq, err := procsrv.Recv()
		switch {
		case errors.Is(err, io.EOF), errors.Is(err, status.Error(codes.Canceled, context.Canceled.Error())):
			return nil
		case err != nil:
			slog.Error("an error occured while processing the requets", "error", err)
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}
		req.Process(procreq.Request)

		switch msg := procreq.Request.(type) {
		case *extproc.ProcessingRequest_RequestHeaders:
			if err := svc.requestHeadersMessage(ctx, req, procsrv); err != nil {
				return err
			}
		case *extproc.ProcessingRequest_RequestBody:
			if err := svc.requestBodyMessage(ctx, req, procsrv); err != nil {
				return err
			}
		case *extproc.ProcessingRequest_RequestTrailers:
			if err := svc.requestTrailersMessage(ctx, req, procsrv); err != nil {
				return err
			}
		case *extproc.ProcessingRequest_ResponseHeaders:
			if err := svc.responseHeadersMessage(ctx, req, procsrv); err != nil {
				return err
			}
		case *extproc.ProcessingRequest_ResponseBody:
			if err := svc.responseBodyMessage(ctx, req, procsrv); err != nil {
				return err
			}
		case *extproc.ProcessingRequest_ResponseTrailers:
			if err := svc.responseTrailersMessage(ctx, req, procsrv); err != nil {
				return err
			}
		default:
			slog.Warn("unhandled message type", "type", fmt.Sprintf("%T", msg))
			return nil
		}
	}
}

// Step 1. Request headers: Contains the headers from the original HTTP request.
func (svc *ExtProcessor) requestHeadersMessage(ctx context.Context, req *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	crw := processor.NewCommonResponseWriter()
	for _, p := range svc.Processors {
		immediateResponse, err := p.RequestHeaders(ctx, crw, req)
		if err != nil {
			return fmt.Errorf("RequestHeaders: failed running processor %T: %w", p, err)
		}
		if immediateResponse != nil {
			return procsrv.Send(&extproc.ProcessingResponse{
				Response: immediateResponse,
			})
		}
		if err := crw.CommonResponse().Validate(); err != nil {
			return fmt.Errorf("RequestHeaders: failed validating response in processor %T: %w", p, err)
		}

	}
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extproc.HeadersResponse{
				Response: crw.CommonResponse(),
			},
		},
	}
	if err := r.ValidateAll(); err != nil {
		return fmt.Errorf("RequestHeaders: failed validating response in processor: %w", err)
	}

	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("RequestHeaders: failed sending response: %w", err)
	}
	return nil
}

// Step 2. (Not implemented) Request body: Delivered if they are present and sent in a single message if the BUFFERED or BUFFERED_PARTIAL mode is chosen, in multiple messages if the STREAMED mode is chosen, and not at all otherwise.
func (svc *ExtProcessor) requestBodyMessage(_ context.Context, _ *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestBody{},
	}
	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("RequestBody: failed sending response: %w", err)
	}
	return nil
}

// Step 3. (Not implemented) Request trailers: Delivered if they are present and if the trailer mode is set to SEND.
func (svc *ExtProcessor) requestTrailersMessage(_ context.Context, _ *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestTrailers{},
	}
	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("RequestTrailers: failed sending response: %w", err)
	}
	return nil
}

// Step 4. Response headers: Contains the headers from the HTTP response. Keep in mind that if the upstream system sends them before processing the request body that this message may arrive before the complete body.
func (svc *ExtProcessor) responseHeadersMessage(ctx context.Context, req *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	crw := processor.NewCommonResponseWriter()
	for _, p := range svc.Processors {
		immediateResponse, err := p.ResponseHeaders(ctx, crw, req)
		if err != nil {
			return fmt.Errorf("ResponseHeaders: failed running processor %T: %w", p, err)
		}
		if immediateResponse != nil {
			return procsrv.Send(&extproc.ProcessingResponse{
				Response: immediateResponse,
			})
		}
		if err := crw.CommonResponse().Validate(); err != nil {
			return fmt.Errorf("ResponseHeaders: failed validating response in processor %T: %w", p, err)
		}
	}
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extproc.HeadersResponse{
				Response: crw.CommonResponse(),
			},
		},
	}
	if err := r.ValidateAll(); err != nil {
		return fmt.Errorf("ResponseHeaders: failed validating response: %w", err)
	}
	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("ResponseHeaders: failed sending response: %w", err)
	}
	return nil
}

// Step 5. (Not implemented) Response body: Sent according to the processing mode like the request body.
func (svc *ExtProcessor) responseBodyMessage(_ context.Context, _ *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseBody{},
	}
	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("ResponseBody: failed sending response: %w", err)
	}
	return nil
}

// Step 6. (Not implemented) Response trailers: Delivered according to the processing mode like the request trailers.
func (svc *ExtProcessor) responseTrailersMessage(_ context.Context, _ *processor.RequestContext, procsrv extproc.ExternalProcessor_ProcessServer) error {
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseTrailers{},
	}
	if err := procsrv.Send(r); err != nil {
		return fmt.Errorf("ResponseTrailers: failed sending response: %w", err)
	}
	return nil
}
