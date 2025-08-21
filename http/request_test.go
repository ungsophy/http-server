package http_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

func TestParseRequest(t *testing.T) {
	var testCases = []struct {
		description     string
		rawRequest      []byte
		expectedRequest *http.Request
		expectedErr     error
	}{
		{
			description: "invalid http request",
			rawRequest:  []byte("foo bar"),
			expectedErr: fmt.Errorf("invalid request line format"),
		},
		{
			description: "request with only one header and has no body",
			rawRequest:  []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			expectedRequest: &http.Request{
				Method:   "GET",
				Path:     "/",
				Protocol: "HTTP/1.1",
				Headers: map[string]string{
					"Host": "example.com",
				},
			},
		},
		{
			description: "request with two headers and has no body",
			rawRequest:  []byte("GET /foo HTTP/1.1\r\nHost: example.com\r\nUser-Agent: mango/pear-raspberry\r\n\r\n"),
			expectedRequest: &http.Request{
				Method:   "GET",
				Path:     "/foo",
				Protocol: "HTTP/1.1",
				Headers: map[string]string{
					"Host":       "example.com",
					"User-Agent": "mango/pear-raspberry",
				},
			},
		},
		{
			description: "request with two headers and has body",
			rawRequest:  []byte("POST /foo HTTP/1.1\r\nHost: example.com\r\nContent-Length: 6\r\n\r\nfoobar"),
			expectedRequest: &http.Request{
				Method:   "POST",
				Path:     "/foo",
				Protocol: "HTTP/1.1",
				Headers: map[string]string{
					"Host":           "example.com",
					"Content-Length": "6",
				},
				Body: []byte("foobar"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(tt *testing.T) {
			req, err := http.ParseRequest(tc.rawRequest)
			if tc.expectedErr != nil {
				if tc.expectedErr.Error() != err.Error() {
					tt.Errorf("expected %v but got %v", tc.expectedErr, err)
				}
				return
			}

			if err != nil {
				tt.Errorf("expected no error but got %v", err)
			}

			if tc.expectedRequest.Method != req.Method {
				tt.Errorf("expected %v method but got %v", tc.expectedRequest.Method, req.Method)
			}

			if tc.expectedRequest.Path != req.Path {
				tt.Errorf("expected %v path but got %v", tc.expectedRequest.Path, req.Path)
			}

			if tc.expectedRequest.Protocol != req.Protocol {
				tt.Errorf("expected %v protocol but got %v", tc.expectedRequest.Protocol, req.Protocol)
			}

			if len(tc.expectedRequest.Headers) != len(req.Headers) {
				tt.Errorf("expected %v headers but got %v", len(tc.expectedRequest.Headers), len(req.Headers))
			}

			for name, value := range req.Headers {
				if tc.expectedRequest.Headers[name] != value {
					tt.Errorf(
						"expected %v header to have value %v but got %v",
						name,
						tc.expectedRequest.Headers[name],
						value,
					)
				}
			}
		})
	}
}
