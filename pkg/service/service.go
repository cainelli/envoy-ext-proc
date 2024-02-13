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

func (svc *ExtProcessor) Process(procsrv extproc.ExternalProcessor_ProcessServer) error {
	ctx := procsrv.Context()
	req := &processor.Request{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			procreq, err := procsrv.Recv()
			if errors.Is(err, io.EOF) || errors.Is(err, status.Error(codes.Canceled, context.Canceled.Error())) {
				return nil
			} else if err != nil {
				slog.Error("an error occured while processing the requets", "error", err)
				return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
			}
			req.Process(procreq.Request)

			switch msg := procreq.Request.(type) {
			case *extproc.ProcessingRequest_RequestHeaders:
				if err := svc.requestHeadersMessage(ctx, req, procsrv); err != nil {
					return err
				}
			case *extproc.ProcessingRequest_ResponseHeaders:
				if err := svc.responseHeadersMessage(ctx, req, procsrv); err != nil {
					return err
				}
			default:
				slog.Warn("unhandled message type", "type", fmt.Sprintf("%T", msg))
			}
		}
	}
}

func (svc *ExtProcessor) requestHeadersMessage(ctx context.Context, req *processor.Request, procsrv extproc.ExternalProcessor_ProcessServer) error {
	cr := &extproc.CommonResponse{}
	for _, p := range svc.Processors {
		ir, err := p.ImmediateResponse(ctx, req)
		if err != nil {
			return fmt.Errorf("ImmediateResponse: failed running processor %T: %w", p, err)
		}
		if ir != nil {
			return procsrv.Send(&extproc.ProcessingResponse{
				Response: ir,
			})
		}
		if err := p.RequestHeaders(ctx, cr, req); err != nil {
			return fmt.Errorf("RequestHeaders: failed running processor %T: %w", p, err)
		}
		if err := cr.Validate(); err != nil {
			return fmt.Errorf("RequestHeaders: failed validating response in processor %T: %w", p, err)
		}

	}
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_RequestHeaders{
			RequestHeaders: &extproc.HeadersResponse{
				Response: cr,
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

func (svc *ExtProcessor) responseHeadersMessage(ctx context.Context, req *processor.Request, procsrv extproc.ExternalProcessor_ProcessServer) error {
	cr := &extproc.CommonResponse{}
	for _, p := range svc.Processors {
		ir, err := p.ImmediateResponse(ctx, req)
		if err != nil {
			return fmt.Errorf("ImmediateResponse: failed running processor %T: %w", p, err)
		}
		if ir != nil {
			return procsrv.Send(&extproc.ProcessingResponse{
				Response: ir,
			})
		}
		if err := p.ResponseHeaders(ctx, cr, req); err != nil {
			return fmt.Errorf("ResponseHeaders: failed running processor %T: %w", p, err)
		}
		if err := cr.Validate(); err != nil {
			return fmt.Errorf("ResponseHeaders: failed validating response in processor %T: %w", p, err)
		}
	}
	r := &extproc.ProcessingResponse{
		Response: &extproc.ProcessingResponse_ResponseHeaders{
			ResponseHeaders: &extproc.HeadersResponse{
				Response: cr,
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
