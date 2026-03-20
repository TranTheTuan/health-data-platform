package main

import (
	"context"
	"log/slog"
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

	http_delivery "github.com/TranTheTuan/health-data-platform/internal/delivery/http"
	tcp_delivery "github.com/TranTheTuan/health-data-platform/internal/delivery/tcp"
	http_handler "github.com/TranTheTuan/health-data-platform/internal/handler/http"
	tcp_handler "github.com/TranTheTuan/health-data-platform/internal/handler/tcp"
	"github.com/TranTheTuan/health-data-platform/internal/repository"
	"github.com/TranTheTuan/health-data-platform/internal/service"
)

func initLogger(cfg *configs.Config) {
	level := slog.LevelDebug
	if cfg.Environment == "PRODUCT" {
		level = slog.LevelInfo
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})
	slog.SetDefault(slog.New(handler))
}

func main() {
	cfg := configs.LoadConfig()
	initLogger(cfg)
	auth.InitGoogleOAuth(cfg)

	// ── Database pool ────────────────────────────────────────────────────────
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// ── Dependency Injection Wiring ──────────────────────────────────────────
	// 1. Repositories
	devRepo := repository.NewDeviceRepository(pool)

	// 2. Services
	devSvc := service.NewDeviceService(devRepo)
	authSvc := service.NewAuthService()

	// 3. Handlers
	devHttpHandler := http_handler.NewDeviceHandler(devSvc)
	authHttpHandler := http_handler.NewAuthHandler(cfg, authSvc)

	devTcpHandler := tcp_handler.NewTCPConnectHandler(devSvc)

	// ── Echo HTTP server (Delivery) ──────────────────────────────────────────
	e := echo.New()

	renderer, err := api.NewTemplateRenderer("web/templates")
	if err != nil {
		slog.Error("Failed to parse templates", "error", err)
		os.Exit(1)
	}
	e.Renderer = renderer

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	http_delivery.RegisterRoutes(e, authHttpHandler, devHttpHandler)

	// ── TCP server (Delivery) ────────────────────────────────────────────────
	tcpSrv := tcp_delivery.NewServer(cfg.TCPAddr, devTcpHandler)

	// ── Graceful shutdown ────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		slog.Info("HTTP server starting", "addr", ":8080")
		return e.Start(":8080")
	})

	g.Go(func() error {
		slog.Info("TCP server starting", "addr", cfg.TCPAddr)
		return tcpSrv.Start(gCtx)
	})

	g.Go(func() error {
		<-gCtx.Done()
		return e.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		slog.Info("Server stopped", "error", err)
	}
}
