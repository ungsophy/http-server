package main

import (
	"bytes"
	"compress/gzip"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/app/http"
)

const (
	PORT = "4221"
)

var (
	directory = flag.String("directory", "", "--directory /tmp")
)

func main() {
	flag.Parse()

	if *directory != "" {
		_, err := os.Stat(*directory)
		if errors.Is(err, os.ErrNotExist) {
			fmt.Printf("directory %s does not exist\n", *directory)
			os.Exit(1)
		}
	}

	mux := http.NewMux()
	mux.HandleFunc("GET /", homeHandler)
	mux.HandleFunc("GET /user-agent", getUserAgentHandler)
	mux.HandleFunc("GET /echo/{str}", echoHandler)
	mux.HandleFunc("GET /files/{filename}", readFileHandler)
	mux.HandleFunc("POST /files/{filename}", createFileHandler)

	server, err := http.NewServer(fmt.Sprintf(":%s", PORT), mux)
	if err != nil {
		fmt.Printf("cannot create HTTP server: %v", err.Error())
		os.Exit(1)
	}

	startErr := server.Start()
	if startErr != nil {
		fmt.Printf("cannot start HTTP server: %v", startErr.Error())
		os.Exit(1)
	}

	fmt.Println("closing HTTP server...")
}

func homeHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
}

func getUserAgentHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "text/plain"
	resp.Body = []byte(req.Headers["User-Agent"])
}

func echoHandler(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "text/plain"
	resp.Body = []byte(req.Params["str"])

	if strings.Contains(req.Headers["Accept-Encoding"], "gzip") {
		body, err := gzipCompress(resp.Body)
		if err != nil {
			fmt.Printf("cannot gzip %v: %v\n", req.Params["str"], err.Error())
			resp.StatusCode = 500
			resp.Body = []byte("cannot gzip")
			return
		}

		resp.Headers["Content-Encoding"] = "gzip"
		resp.Body = body
	}
}

func readFileHandler(req *http.Request, resp *http.Response) {
	filepath := filepath.Join(*directory, req.Params["filename"])
	file, openErr := os.Open(filepath)
	if openErr != nil {
		fmt.Println("error opening file: ", openErr.Error())
		resp.StatusCode = 404
		return
	}
	defer file.Close()

	body, readErr := io.ReadAll(file)
	if readErr != nil {
		fmt.Printf("error reading file from %v: %v\n", filepath, readErr.Error())
		resp.StatusCode = 500
		resp.Body = []byte("cannot read from file")
		return
	}

	resp.StatusCode = 200
	resp.Headers["Content-Type"] = "application/octet-stream"
	resp.Body = body
}

func createFileHandler(req *http.Request, resp *http.Response) {
	filepath := filepath.Join(*directory, req.Params["filename"])
	file, createErr := os.Create(filepath)
	if createErr != nil {
		fmt.Println("error creating file: ", createErr.Error())
		resp.StatusCode = 500
		resp.Body = []byte("cannot create file")
		return
	}
	defer file.Close()

	_, writeErr := file.Write(req.Body)
	if writeErr != nil {
		fmt.Println("error writing file: ", writeErr.Error())
		resp.StatusCode = 500
		resp.Body = []byte("cannot write to file")
		return
	}

	resp.StatusCode = 201
}

func gzipCompress(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	zw := gzip.NewWriter(buf)

	_, writeErr := zw.Write(data)
	if writeErr != nil {
		return nil, writeErr
	}
	zw.Flush()
	// Close gzip writer before reading from buffer
	// to make sure that gzip footer is written to the buffer
	zw.Close()

	return buf.Bytes(), nil
}
