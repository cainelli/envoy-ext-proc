package echo

import (
	"encoding/json"
	"net/http"
)

func RequestHeadersHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("content-type", "application/json")
	resp := struct {
		Headers map[string]string `json:"headers"`
	}{
		Headers: make(map[string]string),
	}
	for headerName := range request.Header {
		resp.Headers[headerName] = request.Header.Get(headerName)
	}

	resp.Headers["Host"] = request.Host
	resp.Headers["Method"] = request.Method

	raw, err := json.Marshal(resp)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = writer.Write(raw)
}

func ResponseHeadersHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("content-type", "application/json")

	resp := make(map[string][]string)
	for k, v := range request.URL.Query() {
		if len(v) <= 0 {
			continue
		}
		writer.Header()[k] = v
		resp[k] = v
	}
	raw, err := json.Marshal(resp)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		return
	}

	_, _ = writer.Write(raw)
}
