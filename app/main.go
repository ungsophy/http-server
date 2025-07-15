package main

import (
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	PORT = "4221"
)

func main() {
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

		go func(c net.Conn) {
			defer c.Close()

			reqBuf := make([]byte, 1024)
			_, readErr := c.Read(reqBuf)
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
			} else if strings.Index(req.Path, "/echo") == 0 {
				statusCode = 200
				headers["Content-Type"] = "text/plain"
				body = []byte(strings.Replace(req.Path, "/echo/", "", 1))
			}

			resp := &Response{
				StatusCode: statusCode,
				Headers:    headers,
				Body:       body,
			}
			c.Write(resp.Bytes())
		}(conn)
	}
}
