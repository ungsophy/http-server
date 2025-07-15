package main

import (
	"bytes"
	"fmt"
)

var (
	statusMap = map[int]string{
		200: "OK",
		404: "Not Found",
	}
)

type Response struct {
	StatusCode int
	Protocol   string
	Headers    map[string]string
	Body       []byte
}

func (r *Response) Bytes() []byte {
	var b bytes.Buffer

	// Write the status line
	b.WriteString(fmt.Sprintf("%v %v\r\n", r.HTTPProtocol(), r.Status()))

	// Write headers
	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	b.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(r.Body)))

	b.Write(r.Body)

	return b.Bytes()
}

func (r *Response) HTTPProtocol() string {
	if r.Protocol == "" {
		return "HTTP/1.1"
	}

	return r.Protocol
}

func (r *Response) Status() string {
	strStatus, exists := statusMap[r.StatusCode]
	if exists {
		return fmt.Sprintf("%d %s", r.StatusCode, strStatus)
	}

	return fmt.Sprintf("%d", r.StatusCode)
}
