package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/shophub/shop/internal/config"
	"github.com/shophub/shop/internal/handler"
	"github.com/shophub/shop/internal/middleware"
)

func main() {
	// Učitaj .env fajl ako postoji (za lokalni development)
	// U Kubernetes-u će env varijable doći iz ConfigMap/Secret
	_ = godotenv.Load()

	// Inicijalizacija konfiguracije
	cfg := config.Load()

	// Inicijalizacija loggera (zap je strukturirani logger, kao Logback u Javi)
	logger, _ := zap.NewProduction()
	if cfg.Server.Env == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// Inicijalizacija router-a
	// chi je kao Spring MVC DispatcherServlet - prima sve zahteve i rutira ih
	r := chi.NewRouter()

	// Middleware stack - izvršava se po redu za svaki zahtev
	// Kao FilterChain u Spring Security / Spring MVC
	r.Use(chimiddleware.RequestID) // Dodaje X-Request-ID header
	r.Use(chimiddleware.Recoverer) // Kao @ExceptionHandler za panike
	r.Use(middleware.RequestLogger(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"}, // Angular dev server
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	// Registracija handler-a
	// Kao @RequestMapping u Spring-u
	healthHandler := handler.NewHealthHandler()
	r.Get("/health", healthHandler.Health)

	// TODO: Ovde će biti registrovani ostali ruter-i:
	// r.Mount("/api/v1/items", itemsRouter(logger))
	// r.Mount("/api/v1/cart", cartRouter(logger))
	// r.Mount("/api/v1/orders", ordersRouter(logger))

	// Pokretanje HTTP servera
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown - kao @PreDestroy ili SmartLifecycle u Spring-u
	// Server čeka da završi aktivne zahteve pre nego što se ugasi
	go func() {
		logger.Info("starting shop service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	// Čekaj OS signal za gašenje (SIGTERM u K8s, Ctrl+C lokalno)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited cleanly")
}
