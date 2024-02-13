package setcookie

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/cainelli/ext-proc/pkg/service/builder"
	"github.com/cainelli/ext-proc/pkg/service/processor"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

type SetCookieProcessor struct {
	processor.NoOpProcessor
}

var _ processor.Processor = &SetCookieProcessor{}

func (*SetCookieProcessor) ResponseHeaders(ctx context.Context, cr *extproc.CommonResponse, req *processor.Request) error {
	rb := builder.NewFromCommonResponse(cr)
	setCookies := parseSetCookies(req)

	action := corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
	for i, cookie := range setCookies {
		if i > 0 {
			action = corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		}
		cookie.SameSite = http.SameSiteStrictMode
		rb.Header("set-cookie", cookie.String(), action) // works
		slog.Info("processing", "processor", "SetCookie", "method", "ResponseHeaders", "set-cookie", cookie.String())
	}

	return nil
}

func parseSetCookies(req *processor.Request) []*http.Cookie {
	header := http.Header{}
	for _, p := range req.ResponseHeaders() {
		if p.GetKey() == "set-cookie" {
			header.Add("set-cookie", string(p.GetRawValue()))
		}
	}

	r := http.Response{Header: header}
	return r.Cookies()
}
