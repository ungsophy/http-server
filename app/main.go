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
			defer conn.Close()
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		}(conn)
	}
}
