package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/sync/errgroup"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/api"
	"github.com/TranTheTuan/health-data-platform/internal/auth"
	"github.com/TranTheTuan/health-data-platform/internal/db"
	"github.com/TranTheTuan/health-data-platform/internal/tcp"
)

func main() {
	cfg := configs.LoadConfig()
	auth.InitGoogleOAuth(cfg)

	// ── Database pool ────────────────────────────────────────────────────────
	// Shared by both the HTTP server (device API) and the TCP server (packet storage).
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// ── Echo HTTP server ─────────────────────────────────────────────────────
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	api.RegisterRoutes(e, cfg, pool)

	// ── TCP server ───────────────────────────────────────────────────────────
	tcpSrv := tcp.NewServer(cfg.TCPAddr, pool)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	// SIGINT (Ctrl+C) or SIGTERM cancels the root context, stopping both servers.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		log.Println("HTTP server starting on :8080")
		return e.Start(":8080")
	})

	g.Go(func() error {
		log.Printf("TCP server starting on %s", cfg.TCPAddr)
		return tcpSrv.Start(gCtx)
	})

	// Shut down Echo when errgroup context is cancelled
	g.Go(func() error {
		<-gCtx.Done()
		return e.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Println("Server stopped:", err)
	}
}

// compile-time assertion that pool satisfies the interface expected by device handler
var _ *sql.DB = (*sql.DB)(nil)
