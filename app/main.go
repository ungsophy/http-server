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
	server, listenErr := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", PORT))
	if listenErr != nil {
		fmt.Printf("failed to bind to port %s\n", PORT)
		os.Exit(1)
	}

	conn, acceptErr := server.Accept()
	if acceptErr != nil {
		fmt.Println("error accepting connection: ", acceptErr.Error())
		os.Exit(1)
	}
	defer conn.Close()

	conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
}
