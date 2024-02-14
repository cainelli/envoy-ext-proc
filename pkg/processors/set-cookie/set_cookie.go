package setcookie

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cainelli/ext-proc/pkg/service/processor"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

type SetCookieProcessor struct {
	processor.NoOpProcessor
}

var _ processor.Processor = &SetCookieProcessor{}

func (*SetCookieProcessor) ResponseHeaders(ctx context.Context, crw *processor.CommonResponseWriter, req *processor.Request) (*extproc.ProcessingResponse_ImmediateResponse, error) {
	for i, cookie := range req.SetCookies() {
		cookie.SameSite = http.SameSiteLaxMode
		cookie.HttpOnly = true
		// The first set-cookie we overwrite the header, the others we append
		if i == 0 {
			crw.HeaderSet("set-cookie", cookie.String())
			continue
		}
		crw.HeaderAppend("set-cookie", cookie.String())
		slog.Info("processing",
			"processor", "SetCookie",
			"phase", "ResponseHeaders",
			"scheme", req.Scheme(),
			"authority", req.Authority(),
			"method", req.Method(),
			"url", req.URL().Path,
			"query", req.URL().RawQuery,
			"request-id", req.RequestID(),
			"status", req.Status(),
			"set-cookie", cookie.String(),
		)
	}

	return nil, nil
}
