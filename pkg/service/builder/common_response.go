package builder

import (
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
)

type CommonResponseBuilder struct {
	commonResponse *extproc.CommonResponse
}

func NewFromCommonResponse(cr *extproc.CommonResponse) *CommonResponseBuilder {
	return &CommonResponseBuilder{
		commonResponse: cr,
	}
}

func (b *CommonResponseBuilder) Header(key string, value string, action corev3.HeaderValueOption_HeaderAppendAction) *CommonResponseBuilder {
	// FIXME: This is not the documented behavior but it seems to be the only way to append a header.
	var append *wrappers.BoolValue
	if action == corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD {
		append = &wrappers.BoolValue{Value: true}
	}
	return b.SetHeaders(&corev3.HeaderValueOption{
		Header: &corev3.HeaderValue{
			Key: key,
			// FIXME: This should be configurable.
			// https://www.envoyproxy.io/docs/envoy/latest/api-v3/service/ext_proc/v3/external_processor.proto#envoy-v3-api-msg-service-ext-proc-v3-httpheaders
			// The headers encoding is based on the runtime guard envoy_reloadable_features_send_header_raw_value setting.
			// When it is true, the header value is encoded in the raw_value field. When it is false, the header value is encoded in the value field.
			RawValue: []byte(value), // FIXME: This depends on Envoy
		},
		AppendAction: action,
		Append:       append,
	})
}

func (b *CommonResponseBuilder) SetHeaders(headers ...*corev3.HeaderValueOption) *CommonResponseBuilder {
	b.init()
	b.commonResponse.HeaderMutation.SetHeaders = append(b.commonResponse.HeaderMutation.SetHeaders, headers...)
	return b
}

func (b *CommonResponseBuilder) RemoveHeaders(headers ...string) *CommonResponseBuilder {
	b.init()
	for _, h := range headers {
		if slices.Contains(b.commonResponse.HeaderMutation.RemoveHeaders, h) {
			continue
		}
		b.commonResponse.HeaderMutation.RemoveHeaders = append(b.commonResponse.HeaderMutation.RemoveHeaders, h)
	}
	return b
}

func (b *CommonResponseBuilder) SetStatus(status extproc.CommonResponse_ResponseStatus) *CommonResponseBuilder {
	b.commonResponse.Status = status
	return b
}

func (b *CommonResponseBuilder) ClearRouteCache(clear bool) *CommonResponseBuilder {
	b.commonResponse.ClearRouteCache = clear
	return b
}

func (b *CommonResponseBuilder) BodyMutation(m *extproc.BodyMutation) *CommonResponseBuilder {
	b.commonResponse.BodyMutation = m
	return b
}

func (b *CommonResponseBuilder) init() {
	if b.commonResponse.HeaderMutation == nil {
		b.commonResponse.HeaderMutation = &extproc.HeaderMutation{}
	}
}
