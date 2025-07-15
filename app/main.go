package main

import (
	"fmt"
	"net"
	"os"
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

			var statusCode int
			if req.Path == "/" {
				statusCode = 200
			} else {
				statusCode = 404
			}

			resp := &Response{
				StatusCode: statusCode,
			}
			c.Write(resp.Bytes())
		}(conn)
	}
}
