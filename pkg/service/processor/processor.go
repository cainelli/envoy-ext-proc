package processor

import (
	"context"

	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

type Processor interface {
	RequestHeaders(ctx context.Context, cr *extproc.CommonResponse, req *Request) error
	ResponseHeaders(ctx context.Context, cr *extproc.CommonResponse, req *Request) error
	ImmediateResponse(ctx context.Context, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error)
}

type NoOpProcessor struct{}

var _ Processor = &NoOpProcessor{}

func (*NoOpProcessor) RequestHeaders(ctx context.Context, cr *extproc.CommonResponse, req *Request) error {
	return nil
}

func (*NoOpProcessor) ResponseHeaders(ctx context.Context, cr *extproc.CommonResponse, req *Request) error {
	return nil
}

func (*NoOpProcessor) ImmediateResponse(ctx context.Context, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error) {
	return nil, nil
}
