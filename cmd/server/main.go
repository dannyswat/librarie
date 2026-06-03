package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"librarie/internal/auth"
	"librarie/internal/config"
	"librarie/internal/db"
	"librarie/internal/user"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		slog.Error("configuration error", "error", err)
		os.Exit(1)
	}

	// Database pool
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to create DB pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("failed to connect to DB", "error", err)
		os.Exit(1)
	}
	slog.Info("database connected")

	queries := db.New(pool)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogMethod:   true,
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogRemoteIP: true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			attrs := []any{
				"method", v.Method,
				"uri", v.URI,
				"status", v.Status,
				"latency_ms", v.Latency.Milliseconds(),
				"remote_ip", v.RemoteIP,
			}
			if v.Error != nil {
				attrs = append(attrs, "error", v.Error.Error())
				slog.Error("request", attrs...)
			} else {
				slog.Info("request", attrs...)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())
	e.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// Routes
	e.GET("/health", healthHandler)

	api := e.Group("/api/v1")
	if err := auth.RegisterRoutes(api, queries, cfg); err != nil {
		slog.Error("failed to register auth routes", "error", err)
		os.Exit(1)
	}
	if err := user.RegisterAdminRoutes(api, queries, cfg); err != nil {
		slog.Error("failed to register admin routes", "error", err)
		os.Exit(1)
	}

	// Server
	addr := ":" + cfg.Port
	slog.Info("starting server", "addr", addr)

	go func() {
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(shutCtx); err != nil {
		slog.Error("shutdown error", "error", err)
	}
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
