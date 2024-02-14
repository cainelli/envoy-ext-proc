package processor

import (
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
	"github.com/golang/protobuf/ptypes/wrappers"
)

// CommonResponseWriter is a wraper on top of extproc.CommonResponse
// It provides a fluent API to mutate the request and response headers and body
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

// HeaderAction sets a header with the given key and value and the given append action
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

// HeaderSet sets a header with the given key and value using the OVERWRITE_IF_EXISTS_OR_ADD action
// This action will overwrite the specified value by discarding any existing values if the header already exists. If the header doesn't exist then this will add the header with specified key and value.
func (crw *CommonResponseWriter) HeaderSet(key string, value string) *CommonResponseWriter {
	return crw.HeaderAction(key, value, corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD)
}

// HeaderAppend appends a header with the given key and value using the APPEND_IF_EXISTS_OR_ADD action
// This action will append the specified value to the existing values if the header already exists. If the header doesn't exist then this will add the header with specified key and value.
func (crw *CommonResponseWriter) HeaderAppend(key string, value string) *CommonResponseWriter {
	return crw.HeaderAction(key, value, corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD)
}

// RemoveHeaders removes these HTTP headers. Attempts to remove system headers -- any header starting with “:“, plus “host“ -- will be ignored.
func (crw *CommonResponseWriter) RemoveHeaders(headers ...string) *CommonResponseWriter {
	for _, h := range headers {
		if slices.Contains(crw.commonResponse.HeaderMutation.RemoveHeaders, h) {
			continue
		}
		crw.commonResponse.HeaderMutation.RemoveHeaders = append(crw.commonResponse.HeaderMutation.RemoveHeaders, h)
	}
	return crw
}

// SetStatus sets the status of the GRPC response.
// If set, provide additional direction on how the Envoy proxy should handle the rest of the HTTP filter chain.
func (crw *CommonResponseWriter) SetStatus(status extproc.CommonResponse_ResponseStatus) *CommonResponseWriter {
	crw.commonResponse.Status = status
	return crw
}

// ClearRouteCache clears the route cache for the current client request. This is necessary if the remote server modified headers that are used to calculate the route. This field is ignored in the response direction.
func (crw *CommonResponseWriter) ClearRouteCache(clear bool) *CommonResponseWriter {
	crw.commonResponse.ClearRouteCache = clear
	return crw
}

// Replace the body of the last message sent to the remote server on this stream.
// If responding to an HttpBody request, simply replace or clear the body chunk that was sent with that request.
// Body mutations may take effect in response either to “header“ or “body“ messages. When it is in response to “header“ messages, it only take effect if the :ref:`status <envoy_v3_api_field_service.ext_proc.v3.CommonResponse.status>` is set to CONTINUE_AND_REPLACE.
func (crw *CommonResponseWriter) BodyMutation(m *extproc.BodyMutation) *CommonResponseWriter {
	crw.commonResponse.BodyMutation = m
	return crw
}

// CommonResponse returns the underlying extproc.CommonResponse
func (crw *CommonResponseWriter) CommonResponse() *extproc.CommonResponse {
	return crw.commonResponse
}

func (crw *CommonResponseWriter) setHeaders(headers ...*corev3.HeaderValueOption) *CommonResponseWriter {
	crw.commonResponse.HeaderMutation.SetHeaders = append(crw.commonResponse.HeaderMutation.SetHeaders, headers...)
	return crw
}
