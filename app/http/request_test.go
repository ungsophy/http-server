package http_test

import (
	"fmt"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func TestParseRequest(t *testing.T) {
	var testCases = []struct {
		name            string
		rawRequest      []byte
		expectedRequest *http.Request
		expectedErr     error
	}{
		{
			name:        "invalid http request",
			rawRequest:  []byte("foo bar"),
			expectedErr: fmt.Errorf("invalid request line format"),
		},
		{
			name:       "request with only one header and has no body",
			rawRequest: []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
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
			name:       "request with two headers and has no body",
			rawRequest: []byte("GET /foo HTTP/1.1\r\nHost: example.com\r\nUser-Agent: mango/pear-raspberry\r\n\r\n"),
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
			name:       "request with two headers and has body",
			rawRequest: []byte("POST /foo HTTP/1.1\r\nHost: example.com\r\nContent-Length: 6\r\n\r\nfoobar"),
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
		t.Run(tc.name, func(tt *testing.T) {
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

func TestEncodings(t *testing.T) {
	var testCases = []struct {
		name       string
		rawRequest []byte
		encodings  []string
	}{
		{
			name:       "request has no Accept-Encoding header",
			rawRequest: []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n"),
			encodings:  []string{},
		},
		{
			name:       "request has Accept-Encoding header",
			rawRequest: []byte("GET /foo HTTP/1.1\r\nHost: example.com\r\nAccept-Encoding: foo, gzip , bar\r\n\r\n"),
			encodings:  []string{"foo", "gzip", "bar"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			req, err := http.ParseRequest(tc.rawRequest)
			if err != nil {
				tt.Errorf("do not expect error but got %v", err)
			}

			if len(tc.encodings) != len(req.Encodings()) {
				tt.Errorf("expect to have %v encodings but got %v", len(tc.encodings), len(req.Encodings()))
			}

			for i, encoding := range req.Encodings() {
				if tc.encodings[i] != encoding {
					tt.Errorf("expect %v but got %v", tc.encodings[i], encoding)
				}
			}
		})
	}
}
