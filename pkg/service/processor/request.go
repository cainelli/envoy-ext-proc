package processor

import (
	"cmp"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

// RequestContext stores the context between the different gRPC messages received from Envoy and it is used to store the request headers, response headers, and other information about the request.
// The Process method should be called on every message received from Envoy in order to update the request object.
// Note that the request object is not thread-safe and should not be shared between goroutines.
type RequestContext struct {
	scheme          string
	authority       string
	method          string
	url             *url.URL
	requestID       string
	status          int
	requestHeaders  http.Header
	responseHeaders http.Header
	cookies         []http.Cookie
	setCookies      []http.Cookie
	metadata        map[string]any
}

// RequestHeaders returns the key-value pairs in an HTTP header.
// The keys should be in canonical form, as returned by http.CanonicalHeaderKey.
func (r *RequestContext) RequestHeaders() map[string][]string {
	return r.requestHeaders
}

// GetRequestHeader gets the first value associated with the given key.
// If there are no values associated with the key, GetRequestHeader returns "". It is case insensitive;
// [textproto.CanonicalMIMEHeaderKey] is used to canonicalize the provided key. GetRequestHeader assumes that all keys are stored in canonical form.
// To use non-canonical keys, access the map directly
func (r *RequestContext) GetRequestHeader(key string) string {
	return r.requestHeaders.Get(key)
}

// RequestHeaderValues returns all values associated with the given key.
// It is case insensitive; [textproto.CanonicalMIMEHeaderKey] is used to canonicalize the provided key.
// To use non-canonical keys, access the map directly. The returned slice is not a copy.
func (r *RequestContext) RequestHeaderValues(key string) []string {
	return r.requestHeaders.Values(key)
}

// ResponseHeaders returns the key-value pairs in an HTTP header.
// The keys should be in canonical form, as returned by http.CanonicalHeaderKey.

func (r *RequestContext) ResponseHeaders() map[string][]string {
	return r.responseHeaders
}

// GetResponseHeader gets the first value associated with the given key.
// If there are no values associated with the key, GetResponseHeader returns "". It is case insensitive;
// [textproto.CanonicalMIMEHeaderKey] is used to canonicalize the provided key. GetResponseHeader assumes that all keys are stored in canonical form.
// To use non-canonical keys, access the map directly
func (r *RequestContext) GetResponseHeader(key string) string {
	return r.responseHeaders.Get(key)
}

// ResponseHeaderValues returns all values associated with the given key.
// It is case insensitive; [textproto.CanonicalMIMEHeaderKey] is used to canonicalize the provided key.
// To use non-canonical keys, access the map directly. The returned slice is not a copy.
func (r *RequestContext) ResponseHeaderValues(key string) []string {
	return r.responseHeaders.Values(key)
}

// Scheme returns the scheme of the request (http or https)
func (r *RequestContext) Scheme() string {
	return r.scheme
}

// Authority returns the authority of the request
func (r *RequestContext) Authority() string {
	return r.authority
}

// Method returns the method of the request (GET, POST, PUT, etc)
func (r *RequestContext) Method() string {
	return r.method
}

// URL returns the URL of the request
func (r *RequestContext) URL() *url.URL {
	return r.url
}

// RequestID returns the request ID of the request
func (r *RequestContext) RequestID() string {
	return r.requestID
}

// Status returns the status of the response
func (r *RequestContext) Status() int {
	return r.status
}

// Cookies returns the cookies of the request
func (r *RequestContext) Cookies() []http.Cookie {
	return r.cookies
}

// SetCookies returns the set cookies of the response
func (r *RequestContext) SetCookies() []http.Cookie {
	return r.setCookies
}

// Metadata returns the metadata of the request, it can be used to excange information between the different processors
func (r *RequestContext) Metadata() map[string]any {
	return r.metadata
}

// Process processes the given message and updates the request object accordingly
// It should be called on every message received from Envoy
func (r *RequestContext) Process(message any) {
	if r.requestHeaders == nil {
		r.requestHeaders = make(http.Header)
	}
	if r.responseHeaders == nil {
		r.responseHeaders = make(http.Header)
	}
	if r.metadata == nil {
		r.metadata = make(map[string]any)
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
	if r.requestID == "" {
		r.requestID = r.GetRequestHeader("x-request-id")
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
	if r.cookies == nil && r.requestHeaders.Get("cookie") != "" {
		httpreq := http.Request{Header: r.requestHeaders}
		for _, cookie := range httpreq.Cookies() {
			r.cookies = append(r.cookies, *cookie)
		}
	}
	if r.setCookies == nil && r.responseHeaders.Get("set-cookie") != "" {
		httpresp := http.Response{Header: r.responseHeaders}
		for _, cookie := range httpresp.Cookies() {
			r.setCookies = append(r.setCookies, *cookie)
		}
	}
}
