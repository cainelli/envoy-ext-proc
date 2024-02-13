package processor

import (
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type CommonResponseWriter struct {
	commonResponse *extproc.CommonResponse
}

func NewCommonResponseWriter() *CommonResponseWriter {
	crw := &CommonResponseWriter{
		commonResponse: &extproc.CommonResponse{
			HeaderMutation: &extproc.HeaderMutation{},
			Trailers:       &corev3.HeaderMap{},
			BodyMutation:   &extproc.BodyMutation{},
		},
	}
	return crw
}

func (crw *CommonResponseWriter) HeaderAction(key string, value string, appendAction corev3.HeaderValueOption_HeaderAppendAction) *CommonResponseWriter {
	// FIXME: This is not the documented behavior but it seems to be the only way to append a header.
	var append *wrappers.BoolValue
	switch appendAction {
	case corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD:
		append = &wrappers.BoolValue{Value: true}
	}
	return crw.setHeaders(&corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{
			Key: key,
			// FIXME: This should be configurable.
			// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto#envoy-v3-api-msg-service-ext-proc-v3-httpheaders
			// The headers encoding is based on the runtime guard envoy_reloadable_features_send_header_raw_value setting.
			// When it is true, the header value is encoded in the raw_value field. When it is false, the header value is encoded in the value field.
			RawValue: []byte(value), // FIXME: This depends on Envoy
		},
		AppendAction: appendAction,
		Append:       append,
	})
}

func (crw *CommonResponseWriter) HeaderSet(key string, value string) *CommonResponseWriter {
	return crw.HeaderAction(key, value, corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD)
}

func (crw *CommonResponseWriter) HeaderAppend(key string, value string) *CommonResponseWriter {
	return crw.HeaderAction(key, value, corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD)
}

func (crw *CommonResponseWriter) RemoveHeaders(headers ...string) *CommonResponseWriter {
	for _, h := range headers {
		if slices.Contains(crw.commonResponse.HeaderMutation.RemoveHeaders, h) {
			continue
		}
		crw.commonResponse.HeaderMutation.RemoveHeaders = append(crw.commonResponse.HeaderMutation.RemoveHeaders, h)
	}
	return crw
}

func (crw *CommonResponseWriter) SetStatus(status extproc.CommonResponse_ResponseStatus) *CommonResponseWriter {
	crw.commonResponse.Status = status
	return crw
}

func (crw *CommonResponseWriter) ClearRouteCache(clear bool) *CommonResponseWriter {
	crw.commonResponse.ClearRouteCache = clear
	return crw
}

func (crw *CommonResponseWriter) BodyMutation(m *extproc.BodyMutation) *CommonResponseWriter {
	crw.commonResponse.BodyMutation = m
	return crw
}

func (crw *CommonResponseWriter) CommonResponse() *extproc.CommonResponse {
	return crw.commonResponse
}

func (crw *CommonResponseWriter) setHeaders(headers ...*corev3.HeaderValueOption) *CommonResponseWriter {
	crw.commonResponse.HeaderMutation.SetHeaders = append(crw.commonResponse.HeaderMutation.SetHeaders, headers...)
	return crw
}
