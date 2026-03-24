package main

import (
	"context"
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
	http_delivery "github.com/TranTheTuan/health-data-platform/internal/delivery/http"
	tcp_delivery "github.com/TranTheTuan/health-data-platform/internal/delivery/tcp"
	http_handler "github.com/TranTheTuan/health-data-platform/internal/handler/http"
	tcp_handler "github.com/TranTheTuan/health-data-platform/internal/handler/tcp"
	"github.com/TranTheTuan/health-data-platform/internal/repository"
	"github.com/TranTheTuan/health-data-platform/internal/service"
)

func main() {
	cfg := configs.LoadConfig()
	auth.InitGoogleOAuth(cfg)

	// ── Database pool ────────────────────────────────────────────────────────
	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	demoSvc := service.NewDemoService(cfg.TCPAddr)
	demoHandler := http_handler.NewDemoHandler(demoSvc, devSvc)

	devTcpHandler := tcp_handler.NewTCPConnectHandler(devSvc)

	// ── Echo HTTP server (Delivery) ──────────────────────────────────────────
	e := echo.New()

	renderer, err := api.NewTemplateRenderer("web/templates")
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	e.Renderer = renderer

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	http_delivery.RegisterRoutes(e, authHttpHandler, devHttpHandler, demoHandler)

	// ── TCP server (Delivery) ────────────────────────────────────────────────
	tcpSrv := tcp_delivery.NewServer(cfg.TCPAddr, devTcpHandler)

	// ── Graceful shutdown ────────────────────────────────────────────────────
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

	g.Go(func() error {
		<-gCtx.Done()
		return e.Shutdown(context.Background())
	})

	if err := g.Wait(); err != nil {
		log.Println("Server stopped:", err)
	}
}
