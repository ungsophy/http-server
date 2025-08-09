package http

import (
	"errors"
	"fmt"
	"io"
	"net"
)

type Server struct {
	Address string
	Handler *Mux
}

func NewServer(address string, handler *Mux) (*Server, error) {
	return &Server{
		Address: address,
		Handler: handler,
	}, nil
}

func (s *Server) Start() error {
	listener, listenErr := net.Listen("tcp", s.Address)
	if listenErr != nil {
		return fmt.Errorf("cannot start tcp server on %v: %w", s.Address, listenErr)
	}
	defer listener.Close()

	for {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return fmt.Errorf("error accepting connection: %w", acceptErr)
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	fmt.Println("new connection from", conn.RemoteAddr().String())

	// Connection is only closed when "Connection: close" header is present in the request.
	// Otherwise connection is re-used.
	for {
		var reqBuf = make([]byte, 1024)
		_, readErr := conn.Read(reqBuf)
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) {
				fmt.Println("error reading request: ", readErr.Error())
			}
			return
		}

		req, parseReqErr := ParseRequest(reqBuf)
		if parseReqErr != nil {
			fmt.Println("error parsing request: ", parseReqErr.Error())
			return
		}

		// Handle request and write response
		resp := NewResponse()
		s.Handler.HandleReqeust(req, resp)
		conn.Write(resp.Bytes())

		// Don't close TCP connection; waiting for new requests from the same connection.
		if resp.Headers["Connection"] != "close" {
			continue
		}

		// Close TCP connection
		closeErr := conn.Close()
		if closeErr != nil {
			fmt.Println("error closing connection: ", closeErr.Error())
		}
		break
	}
}
