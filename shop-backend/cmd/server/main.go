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
	"github.com/shophub/shop/internal/database"
	"github.com/shophub/shop/internal/handler"
	"github.com/shophub/shop/internal/middleware"
	"github.com/shophub/shop/internal/repository"
	pgrepo "github.com/shophub/shop/internal/repository/postgres"
	redisrepo "github.com/shophub/shop/internal/repository/redis"
	"github.com/shophub/shop/internal/service"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	logger, _ := zap.NewProduction()
	if cfg.Server.Env == "development" {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	itemRepo, cartRepo, orderRepo := initRepositories(cfg, logger)

	// Services
	itemSvc := service.NewItemService(itemRepo)
	cartSvc := service.NewCartService(cartRepo, itemRepo)
	orderSvc := service.NewOrderService(orderRepo, cartRepo, itemRepo)

	// Handlers
	healthHandler := handler.NewHealthHandler()
	itemHandler := handler.NewItemHandler(itemSvc, logger)
	cartHandler := handler.NewCartHandler(cartSvc, logger)
	orderHandler := handler.NewOrderHandler(orderSvc, logger)

	r := chi.NewRouter()
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.RequestLogger(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", healthHandler.Health)

	auth := middleware.Authenticate(cfg.JWT.Secret)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/items", func(r chi.Router) {
			r.Get("/", itemHandler.List)
			r.Get("/{id}", itemHandler.GetByID)
			r.Group(func(r chi.Router) {
				r.Use(auth, middleware.RequireAdmin)
				r.Post("/", itemHandler.Create)
				r.Patch("/{id}", itemHandler.Update)
				r.Delete("/{id}", itemHandler.Delete)
			})
		})

		r.Route("/cart", func(r chi.Router) {
			r.Use(auth)
			r.Get("/", cartHandler.Get)
			r.Delete("/", cartHandler.Clear)
			r.Post("/items", cartHandler.AddItem)
			r.Put("/items/{itemId}", cartHandler.SetItemQuantity)
			r.Delete("/items/{itemId}", cartHandler.RemoveItem)
		})

		r.Route("/orders", func(r chi.Router) {
			r.Use(auth)
			r.Get("/", orderHandler.List)
			r.Post("/", orderHandler.Create)
			r.Get("/{id}", orderHandler.GetByID)
			r.Post("/{id}/confirm", orderHandler.ConfirmPayment)
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin)
				r.Patch("/{id}/status", orderHandler.UpdateStatus)
			})
		})
	})

	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("starting shop service",
			zap.String("port", cfg.Server.Port),
			zap.String("db_type", cfg.Database.Type),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

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

// initRepositories connects to the configured backend and returns the three repositories.
func initRepositories(cfg *config.Config, logger *zap.Logger) (
	repository.ItemRepository,
	repository.CartRepository,
	repository.OrderRepository,
) {
	if cfg.Database.Type == "redis" {
		rdb, err := database.NewRedis(cfg.Redis)
		if err != nil {
			logger.Fatal("failed to connect to redis", zap.Error(err))
		}
		logger.Info("connected to redis", zap.String("addr", cfg.Redis.Addr))
		return redisrepo.NewItemRepository(rdb),
			redisrepo.NewCartRepository(rdb),
			redisrepo.NewOrderRepository(rdb)
	}

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}
	logger.Info("connected to postgres and migrated")
	return pgrepo.NewItemRepository(db),
		pgrepo.NewCartRepository(db),
		pgrepo.NewOrderRepository(db)
}
