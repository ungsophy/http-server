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
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%s", PORT))
	if err != nil {
		fmt.Printf("failed to bind to port %s\n", PORT)
		os.Exit(1)
	}

	_, err = l.Accept()
	if err != nil {
		fmt.Println("error accepting connection: ", err.Error())
		os.Exit(1)
	}
}
