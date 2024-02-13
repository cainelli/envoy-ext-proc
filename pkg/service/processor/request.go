package processor

import (
	"net/url"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	extproc "github.com/envoyproxy/go-control-plane/envoy/service/ext_proc/v3"
)

type Request struct {
	scheme          string
	authority       string
	method          string
	url             *url.URL
	requestId       string
	requestHeaders  *extproc.ProcessingRequest_RequestHeaders
	responseHeaders *extproc.ProcessingRequest_ResponseHeaders
	commonResponse  *extproc.CommonResponse
}

func (r *Request) RequestHeaders() []*corev3.HeaderValue {
	return r.requestHeaders.RequestHeaders.GetHeaders().GetHeaders()
}

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

func (r *Request) ResponseHeaders() []*corev3.HeaderValue {
	return r.responseHeaders.ResponseHeaders.GetHeaders().GetHeaders()
}

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

func (r *Request) CommonResponse() *extproc.CommonResponse {
	return r.commonResponse
}

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
}

func (r *Request) init() {
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
}
