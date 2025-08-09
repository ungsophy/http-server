package http

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

var (
	delimiter = []byte("\r\n")
)

type Request struct {
	Method   string
	Path     string
	Protocol string
	Headers  map[string]string
	Body     []byte
	Params   map[string]string
}

func ParseRequest(data []byte) (*Request, error) {
	// http request format:
	//
	// GET /index HTTP/1.1\r\n
	// Host: example.com\r\n
	// \r\n
	// body

	lines := bytes.Split(data, delimiter)
	if len(lines) < 1 {
		return nil, fmt.Errorf("invalid request format")
	}

	// first line is the request line
	// e.g., "GET /index HTTP/1.1"
	requestLineParts := bytes.SplitN(lines[0], []byte(" "), 3)
	if len(requestLineParts) < 3 {
		return nil, fmt.Errorf("invalid request line format")
	}

	// headers
	headers := make(map[string]string)
	for _, line := range lines[1:] {
		if len(line) == 0 {
			break // End of headers
		}

		headerParts := bytes.SplitN(line, []byte(":"), 2)
		if len(headerParts) != 2 {
			return nil, fmt.Errorf("invalid header format: %s", line)
		}
		key := string(bytes.TrimSpace(headerParts[0]))
		value := string(bytes.TrimSpace(headerParts[1]))
		headers[key] = value
	}

	// body
	bodyStartIndex := bytes.Index(data, []byte("\r\n\r\n"))
	if bodyStartIndex == -1 {
		return nil, fmt.Errorf("invalid end of headers")
	}
	bodyStartIndex = bodyStartIndex + 4 // +4 to skip the \r\n\r\n

	var body []byte
	contentLength, exists := headers["Content-Length"]
	if exists {
		contentLengthInt, err := strconv.Atoi(contentLength)
		if err != nil {
			return nil, fmt.Errorf("invalid Content-Length header: %s", contentLength)
		}

		body = data[bodyStartIndex : bodyStartIndex+contentLengthInt]
	}

	return &Request{
		Method:   string(requestLineParts[0]),
		Path:     string(requestLineParts[1]),
		Protocol: string(requestLineParts[2]),
		Headers:  headers,
		Body:     body,
	}, nil
}

func (r *Request) Encodings() []string {
	var encodings []string

	acceptEncoding, exists := r.Headers["Accept-Encoding"]
	if exists {
		tmpEncodings := strings.Split(acceptEncoding, ",")
		for _, encoding := range tmpEncodings {
			encodings = append(encodings, strings.TrimSpace(encoding))
		}
	}

	return encodings
}
