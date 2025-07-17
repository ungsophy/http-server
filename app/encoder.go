package main

import (
	"bytes"
	"compress/gzip"
)

type Encoder interface {
	Encode(data []byte) ([]byte, error)
	Name() string
}

type GZipEncoder struct{}

func (g *GZipEncoder) Name() string {
	return ENCODER_GZIP
}

func (g *GZipEncoder) Encode(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	defer zw.Close()

	_, writeErr := zw.Write(data)
	if writeErr != nil {
		return nil, writeErr
	}
	zw.Flush()

	return buf.Bytes(), nil
}
