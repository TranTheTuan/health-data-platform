package tcp

import (
	"context"
	"fmt"
	"net"

	tcp_handler "github.com/TranTheTuan/health-data-platform/internal/handler/tcp"
)

type Server struct {
	addr    string
	handler *tcp_handler.TCPConnectHandler
}

func NewServer(addr string, handler *tcp_handler.TCPConnectHandler) *Server {
	return &Server{addr: addr, handler: handler}
}

func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp: listen %s: %w", s.addr, err)
	}

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("tcp: accept: %w", err)
			}
		}

		go s.handler.HandleConnection(conn)
	}
}
