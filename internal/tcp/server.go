// Package tcp implements the TCP listener for smartwatch IW protocol ingestion.
// The server accepts persistent connections, one goroutine per connection.
package tcp

import (
	"context"
	"database/sql"
	"fmt"
	"net"
)

// Server listens for incoming TCP connections from smartwatch devices.
type Server struct {
	addr string
	db   *sql.DB
}

// NewServer creates a new TCP Server bound to addr (e.g. ":9090").
func NewServer(addr string, db *sql.DB) *Server {
	return &Server{addr: addr, db: db}
}

// Start begins listening for connections. It blocks until ctx is cancelled,
// at which point the listener is closed and Start returns.
//
// Designed to run under golang.org/x/sync/errgroup alongside the HTTP server.
func (s *Server) Start(ctx context.Context) error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("tcp: listen %s: %w", s.addr, err)
	}

	// Close listener when context is cancelled — this unblocks ln.Accept()
	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			// Context cancellation closes ln, causing Accept to return with an error.
			// Treat any error after ctx cancellation as a clean shutdown.
			select {
			case <-ctx.Done():
				return nil
			default:
				return fmt.Errorf("tcp: accept: %w", err)
			}
		}

		// Each device connection runs in its own goroutine.
		// At wearable scale (hundreds concurrent) this is safe; see phase-03 Key Insights.
		go HandleConnection(conn, s.db)
	}
}
