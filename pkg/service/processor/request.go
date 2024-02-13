package processor

import (
	"cmp"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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
	requestHeaders  http.Header
	responseHeaders http.Header
}

// RequestHeaders returns the HTTP request headers. All header keys will be lower-cased.
// When using this method to acces the headers, make sure you use the correct HeaderValue property, GetValue() or GetRawValue().
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-headervalue
func (r *Request) RequestHeaders() map[string][]string {
	return r.requestHeaders
}

// GetRequestHeader returns the value of the header with the given key. If the header is not found, it returns an empty string.
// If the header has multiple values, it returns the first one. If you need all the values, use RequestHeaders()
func (r *Request) GetRequestHeader(key string) string {
	return r.requestHeaders.Get(key)
}

// RequestHeaderValues returns all the values of the header with the given key. If the header is not found, it returns an empty slice.
func (r *Request) RequestHeaderValues(key string) []string {
	return r.requestHeaders.Values(key)
}

// ResponseHeaders returns the HTTP response headers. All header keys will be lower-cased.
// When using this method to acces the headers, make sure you use the correct HeaderValue property, GetValue() or GetRawValue().
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/core/v3/base.proto#envoy-v3-api-msg-config-core-v3-headervalue
func (r *Request) ResponseHeaders() map[string][]string {
	return r.responseHeaders
}

// GetResponseHeader returns the value of the header with the given key. If the header is not found, it returns an empty string.
// If the header has multiple values, it returns the first one. If you need all the values, use ResponseHeaders()
func (r *Request) GetResponseHeader(key string) string {
	return r.responseHeaders.Get(key)
}

// RequestHeaderValues returns all the values of the header with the given key. If the header is not found, it returns an empty slice.
func (r *Request) ResponseHeaderValues(key string) []string {
	return r.responseHeaders.Values(key)
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
	if r.requestHeaders == nil {
		r.requestHeaders = make(http.Header)
	}
	if r.responseHeaders == nil {
		r.responseHeaders = make(http.Header)
	}

	switch msg := any(message).(type) {
	case *extproc.ProcessingRequest_RequestHeaders:
		for _, header := range msg.RequestHeaders.GetHeaders().GetHeaders() {
			headerValue := cmp.Or(string(header.GetRawValue()), header.GetValue())
			r.requestHeaders.Add(header.Key, headerValue)
		}
	case *extproc.ProcessingRequest_ResponseHeaders:
		for _, header := range msg.ResponseHeaders.GetHeaders().GetHeaders() {
			headerValue := cmp.Or(string(header.GetRawValue()), header.GetValue())
			r.responseHeaders.Add(header.Key, headerValue)
		}
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
