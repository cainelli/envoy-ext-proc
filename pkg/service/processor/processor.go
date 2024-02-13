package processor

import (
	"context"

	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

type Processor interface {
	RequestHeaders(ctx context.Context, crw *CommonResponseWriter, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error)
	ResponseHeaders(ctx context.Context, crw *CommonResponseWriter, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error)
}

type NoOpProcessor struct{}

var _ Processor = &NoOpProcessor{}

func (*NoOpProcessor) RequestHeaders(ctx context.Context, crw *CommonResponseWriter, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error) {
	return nil, nil
}

func (*NoOpProcessor) ResponseHeaders(ctx context.Context, crw *CommonResponseWriter, req *Request) (*extproc.ProcessingResponse_ImmediateResponse, error) {
	return nil, nil
}
