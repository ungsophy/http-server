package main

import (
	"bytes"
	"compress/gzip"
)

type Encoder interface {
	Encode(data []byte) ([]byte, error)
}

type GZipEncoder struct{}

func (g *GZipEncoder) Encode(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)

	_, writeErr := zw.Write(data)
	if writeErr != nil {
		return nil, writeErr
	}
	zw.Flush()
	zw.Close()

	return buf.Bytes(), nil
}
