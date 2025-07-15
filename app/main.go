package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
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

	listener, listenErr := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", PORT))
	if listenErr != nil {
		fmt.Printf("failed to bind to port %s\n", PORT)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			fmt.Println("error accepting connection: ", acceptErr.Error())
			os.Exit(1)
		}

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	fmt.Println("new connection from", conn.RemoteAddr().String())

	reqBuf := make([]byte, 1024)
	_, readErr := conn.Read(reqBuf)
	if readErr != nil {
		fmt.Println("error reading request: ", readErr.Error())
		return
	}
	req, err := ParseRequest(reqBuf)
	if err != nil {
		fmt.Println("error parsing request: ", err.Error())
		return
	}

	var statusCode int = 404
	var headers = make(map[string]string)
	var body []byte

	if req.Path == "/" {
		statusCode = 200
	} else if req.Path == "/user-agent" {
		statusCode = 200
		headers["Content-Type"] = "text/plain"
		body = []byte(req.Headers["User-Agent"])
	} else if strings.Index(req.Path, "/echo/") == 0 {
		statusCode = 200
		headers["Content-Type"] = "text/plain"
		body = []byte(strings.Replace(req.Path, "/echo/", "", 1))
	} else if strings.Index(req.Path, "/files/") == 0 {
		// Ensure the directory is set
		_, filename := path.Split(req.Path)
		filepath := filepath.Join(*directory, filename)

		file, openErr := os.Open(filepath)
		if openErr == nil {
			var readErr error
			body, readErr = io.ReadAll(file)
			if readErr != nil {
				fmt.Println("error reading file: ", readErr.Error())
				return
			}

			statusCode = 200
			headers["Content-Type"] = "application/octet-stream"
		} else {
			fmt.Println("error opening file: ", openErr.Error())
			statusCode = 404
		}
		defer file.Close()
	}

	resp := &Response{
		StatusCode: statusCode,
		Headers:    headers,
		Body:       body,
	}
	conn.Write(resp.Bytes())
}
