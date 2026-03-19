package main

import (
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/TranTheTuan/health-data-platform/configs"
	"github.com/TranTheTuan/health-data-platform/internal/api"
	"github.com/TranTheTuan/health-data-platform/internal/auth"
)

func main() {
	cfg := configs.LoadConfig()
	auth.InitGoogleOAuth(cfg)

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	api.RegisterRoutes(e, cfg)

	log.Println("Starting server on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
