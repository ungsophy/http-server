package main

import (
	"testing"
)

func TestParseRequest(t *testing.T) {
	data := []byte("GET / HTTP/1.1\r\nHost: example.com\r\n\r\n")
	req, err := ParseRequest(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if req.Method != "GET" {
		t.Errorf("expected method 'GET', got '%s'", req.Method)
	}
	if req.Path != "/" {
		t.Errorf("expected path '/', got '%s'", req.Path)
	}
	if len(req.Headers) != 1 {
		t.Error("expected 1 header, got", len(req.Headers))
	}
	if req.Headers["Host"] != "example.com" {
		t.Errorf("expected header 'Host' to be 'example.com', got '%s'", req.Headers["Host"])
	}
}
