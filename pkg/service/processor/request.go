package processor

import (
	"net/url"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// Request stores the context between the different gRPC messages received from Envoy and it is used to store the request headers, response headers, and other information about the request.
// The Process method should be called on every message received from Envoy in order to update the request object.
// Note that the request object is not thread-safe and should not be shared between goroutines.
type Request struct {
	scheme          string
	authority       string
	method          string
	url             *url.URL
	requestId       string
	status          int
	requestHeaders  *extproc.ProcessingRequest_RequestHeaders
	responseHeaders *extproc.ProcessingRequest_ResponseHeaders
}

// RequestHeaders returns the HTTP request headers. All header keys will be lower-cased.
// When using this method to acces the headers, make sure you use the correct HeaderValue property, GetValue() or GetRawValue().
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-headervalue
func (r *Request) RequestHeaders() []*corev3.HeaderValue {
	return r.requestHeaders.RequestHeaders.GetHeaders().GetHeaders()
}

// GetRequestHeader returns the value of the header with the given key. If the header is not found, it returns an empty string.
// If the header has multiple values, it returns the first one. If you need all the values, use RequestHeaders()
func (r *Request) GetRequestHeader(key string) string {
	for _, header := range r.RequestHeaders() {
		if header.Key == key {
			if header.GetValue() != "" {
				return header.GetValue()
			}
			return string(header.GetRawValue())
		}
	}
	return ""
}

// ResponseHeaders returns the HTTP response headers. All header keys will be lower-cased.
// When using this method to acces the headers, make sure you use the correct HeaderValue property, GetValue() or GetRawValue().
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-headervalue
func (r *Request) ResponseHeaders() []*corev3.HeaderValue {
	return r.responseHeaders.ResponseHeaders.GetHeaders().GetHeaders()
}

// GetResponseHeader returns the value of the header with the given key. If the header is not found, it returns an empty string.
// If the header has multiple values, it returns the first one. If you need all the values, use ResponseHeaders()
func (r *Request) GetResponseHeader(key string) string {
	for _, header := range r.ResponseHeaders() {
		if header.Key == key {
			if header.GetValue() != "" {
				return header.GetValue()
			}
			return string(header.GetRawValue())
		}
	}
	return ""
}

// Scheme returns the scheme of the request (http or https)
func (r *Request) Scheme() string {
	return r.scheme
}

// Authority returns the authority of the request
func (r *Request) Authority() string {
	return r.authority
}

// Method returns the method of the request (GET, POST, PUT, etc)
func (r *Request) Method() string {
	return r.method
}

// URL returns the URL of the request
func (r *Request) URL() *url.URL {
	return r.url
}

// RequestID returns the request ID of the request
func (r *Request) RequestID() string {
	return r.requestId
}

// Status returns the status of the response
func (r *Request) Status() int {
	return r.status
}

// Process processes the given message and updates the request object accordingly
// It should be called on every message received from Envoy
func (r *Request) Process(message any) {
	switch msg := any(message).(type) {
	case *extproc.ProcessingRequest_RequestHeaders:
		r.requestHeaders = msg
	case *extproc.ProcessingRequest_ResponseHeaders:
		r.responseHeaders = msg
	}

	switch {
	case r.requestHeaders == nil:
		r.requestHeaders = &extproc.ProcessingRequest_RequestHeaders{}
	case r.responseHeaders == nil:
		r.responseHeaders = &extproc.ProcessingRequest_ResponseHeaders{}
	}

	var err error
	if r.scheme == "" {
		r.scheme = r.GetRequestHeader(":scheme")
	}
	if r.authority == "" {
		r.authority = r.GetRequestHeader(":authority")
	}
	if r.method == "" {
		r.method = r.GetRequestHeader(":method")
	}
	if r.requestId == "" {
		r.requestId = r.GetRequestHeader("x-request-id")
	}
	if r.url == nil {
		r.url, err = url.Parse(r.GetRequestHeader(":path"))
		if err != nil {
			r.url = &url.URL{
				Path:     strings.Split(r.GetRequestHeader(":path"), "?")[0],
				RawPath:  r.GetRequestHeader(":path"),
				RawQuery: strings.Split(r.GetRequestHeader(":path"), "?")[1],
			}
		}
	}
	if r.status == 0 {
		status, _ := strconv.Atoi(r.GetResponseHeader(":status"))
		r.status = status
	}
}
