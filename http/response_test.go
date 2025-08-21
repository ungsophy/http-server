package http_test

import (
	"bytes"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

func TestBytes(t *testing.T) {
	var testCases = []struct {
		description   string
		response      *http.Response
		expectedBytes []byte
	}{
		{
			description: "response with 201 status code with no body",
			response: &http.Response{
				StatusCode: 201,
			},
			expectedBytes: []byte("HTTP/1.1 201 Created\r\nContent-Length: 0\r\n\r\n"),
		},
		{
			description: "response with 204 status code with body",
			response: &http.Response{
				StatusCode: 204,
			},
			expectedBytes: []byte("HTTP/1.1 204\r\nContent-Length: 0\r\n\r\n"),
		},
		{
			description: "response with 200 status code with body",
			response: &http.Response{
				StatusCode: 200,
				Body:       []byte("Hello, World!"),
			},
			expectedBytes: []byte("HTTP/1.1 200 OK\r\nContent-Length: 13\r\n\r\nHello, World!"),
		},
		{
			description: "response with 202 status code with JSON body",
			response: &http.Response{
				StatusCode: 202,
				Headers:    map[string]string{"Content-Type": "application/json"},
				Body:       []byte(`{"message":"accepted"}`),
			},
			expectedBytes: []byte("HTTP/1.1 202\r\nContent-Type: application/json\r\nContent-Length: 22\r\n\r\n{\"message\":\"accepted\"}"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			b := tc.response.Bytes()
			if !bytes.Equal(b, tc.expectedBytes) {
				t.Errorf("expected %q, got %q", tc.expectedBytes, b)
			}
		})
	}
}
