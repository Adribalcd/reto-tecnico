package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"reto-tecnico/go-api/internal/auth"
	"reto-tecnico/go-api/internal/config"
	"reto-tecnico/go-api/internal/httpapi"
	"reto-tecnico/go-api/internal/statistics"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error de configuración: %v", err)
	}

	app := fiber.New(fiber.Config{
		AppName:               "go-api",
		DisableStartupMessage: true,
		ReadTimeout:           cfg.RequestTimeout,
		WriteTimeout:          cfg.RequestTimeout,
		ErrorHandler:          httpapi.ErrorHandler,
	})
	app.Use(recover.New())
	app.Use(logger.New())

	(&httpapi.Server{
		Tokens:       auth.NewVerifier(cfg.JWTSecret, cfg.JWTIssuer),
		Statistics:   statistics.NewClient(cfg.NodeAPIURL, cfg.RequestTimeout),
		MaxMatrixDim: cfg.MaxMatrixDim,
	}).Register(app)

	go func() {
		log.Printf("go-api escuchando en %s", cfg.Address)
		if err := app.Listen(cfg.Address); err != nil {
			log.Fatalf("el servidor se detuvo: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("apagando go-api")
	if err := app.ShutdownWithTimeout(10 * time.Second); err != nil {
		log.Printf("el apagado terminó con error: %v", err)
	}
}
