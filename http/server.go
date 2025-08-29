package http

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
)

const (
	reqTmpBufInKB = 1024 * 4 // 4KB buffer
)

type Server struct {
	running  bool
	logger   *slog.Logger
	listener net.Listener

	Address string
	Handler *Mux
	Created chan bool
}

func NewServer(address string, handler *Mux, logger *slog.Logger) (*Server, error) {
	return &Server{
		logger: logger,

		Address: address,
		Handler: handler,
		Created: make(chan bool, 1),
	}, nil
}

func (s *Server) Start() error {
	var listenErr error
	s.listener, listenErr = net.Listen("tcp", s.Address)
	if listenErr != nil {
		return fmt.Errorf("cannot start tcp server on %v: %w", s.Address, listenErr)
	}

	s.Created <- true
	s.running = true

	defer func() {
		err := s.listener.Close()
		if err != nil {
			s.logger.Error("error closing listener", "error", err)
		}

		s.logger.Info("server is stopped")
	}()

	for {
		if !s.running {
			s.logger.Info("server is stopping...")
			return nil
		}

		conn, acceptErr := s.listener.Accept()
		if acceptErr != nil {
			return fmt.Errorf("error accepting connection: %w", acceptErr)
		}

		// TODO: limit the number of goroutines
		go s.handleConnection(conn)
	}
}

func (s *Server) Stop() error {
	err := s.listener.Close()
	if err != nil {
		return fmt.Errorf("error closing listener: %w", err)
	}

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer s.closeConnection(conn)

	s.logger.Info("new connection", "remote_addr", conn.RemoteAddr().String())

	// Connection is only closed when one of the following cases happens:
	// - server is told to stop
	// - client sends EOF
	// - conn.Read returns an error
	// - cannot parse request
	// - "Connection: close" header is present in the request
	// Otherwise connection is re-used.
	for {
		// Stop handling requests when server is told to stop
		if !s.running {
			return
		}

		var requestData []byte

		for {
			var reqBuf = make([]byte, reqTmpBufInKB)
			n, readErr := conn.Read(reqBuf)
			if readErr != nil {
				if errors.Is(readErr, io.EOF) {
					s.logger.Info("connection closed by client")
				} else {
					s.logger.Error("error reading request", "error", readErr)
				}

				return
			}

			requestData = append(requestData, reqBuf[:n]...)
			if n < len(reqBuf) {
				break
			}
		}

		req, parseReqErr := ParseRequest(requestData)
		if parseReqErr != nil {
			s.logger.Error("error parsing request", "error", parseReqErr)
			return
		}

		// Handle request and write response
		resp := NewResponse()
		s.Handler.HandleRequest(req, resp)
		conn.Write(resp.Bytes())

		// Don't close TCP connection; waiting for new requests from the same connection.
		if resp.Headers["Connection"] != "close" {
			continue
		}

		break
	}
}

func (s *Server) closeConnection(conn net.Conn) {
	if conn == nil {
		return
	}

	closeErr := conn.Close()
	if closeErr != nil {
		s.logger.Error("error closing connection", "error", closeErr)
	}
}
