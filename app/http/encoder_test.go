package http_test

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

func TestEncode(t *testing.T) {
	var expectedData = []byte("foo bar baz")
	gzipEncoder := &http.GZipEncoder{}
	compressedBytes, err := gzipEncoder.Encode(expectedData)
	if err != nil {
		t.Errorf("expect no error but got %v", err)
	}

	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedBytes))
	if err != nil {
		t.Errorf("expect no error but got %v", err)
	}

	var data = make([]byte, len(expectedData))
	n, err := gzipReader.Read(data)
	if err != nil {
		t.Errorf("expect no error but got %v", err)
	}
	if len(expectedData) != n {
		t.Errorf("expect %v but got %v", len(expectedData), n)
	}
	if string(expectedData) != string(data) {
		t.Errorf("expect %v but got %v", string(expectedData), string(data))
	}
}
