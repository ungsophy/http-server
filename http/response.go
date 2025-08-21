package http

import (
	"bytes"
	"fmt"
)

const (
	protocolVersion1_1 = "HTTP/1.1"
)

var (
	statusMap = map[int]string{
		200: "OK",
		201: "Created",
		404: "Not Found",
	}
)

type Response struct {
	protocol string

	StatusCode int
	Headers    map[string]string
	Body       []byte
}

func NewResponse() *Response {
	return &Response{
		protocol: protocolVersion1_1,

		Headers: make(map[string]string),
		Body:    make([]byte, 0),
	}
}

func (r *Response) Bytes() []byte {
	var b bytes.Buffer

	// Write the status line
	b.WriteString(fmt.Sprintf("%v %v\r\n", r.strProtocol(), r.strStatus()))

	// Write headers
	for k, v := range r.Headers {
		b.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	_, exists := r.Headers["Content-Length"]
	if exists {
		b.WriteString("\r\n")
	} else {
		b.WriteString(fmt.Sprintf("Content-Length: %d\r\n\r\n", len(r.Body)))
	}

	b.WriteString(string(r.Body))

	return b.Bytes()
}

func (r *Response) strProtocol() string {
	if r.protocol == "" {
		return protocolVersion1_1
	}

	return r.protocol
}

func (r *Response) strStatus() string {
	strStatus, exists := statusMap[r.StatusCode]
	if exists {
		return fmt.Sprintf("%d %s", r.StatusCode, strStatus)
	}

	return fmt.Sprintf("%d", r.StatusCode)
}
