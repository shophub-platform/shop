package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/shophub/shop/internal/config"
	"github.com/shophub/shop/internal/database"
	pgrepo "github.com/shophub/shop/internal/repository/postgres"
	redisrepo "github.com/shophub/shop/internal/repository/redis"
	"github.com/shophub/shop/internal/repository"
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

	bc := cfg.Blockchain
	if bc.SepoliaRPCURL == "" || bc.ShopWalletAddress == "" || bc.MockUSDTAddress == "" {
		logger.Fatal("SEPOLIA_RPC_URL, SHOP_WALLET_ADDRESS and MOCKUSDT_ADDRESS must be set")
	}

	var orderRepo repository.OrderRepository
	if cfg.Database.Type == "redis" {
		rdb, err := database.NewRedis(cfg.Redis)
		if err != nil {
			logger.Fatal("connect redis", zap.Error(err))
		}
		orderRepo = redisrepo.NewOrderRepository(rdb)
	} else {
		db, err := database.NewPostgres(cfg.Database)
		if err != nil {
			logger.Fatal("connect postgres", zap.Error(err))
		}
		orderRepo = pgrepo.NewOrderRepository(db)
	}

	listener := service.NewListener(
		bc.SepoliaRPCURL,
		bc.ShopWalletAddress,
		bc.MockUSDTAddress,
		bc.ListenerBackendURL,
		bc.ListenerInternalKey,
		orderRepo,
		logger,
	)

	ctx, cancel := context.WithCancel(context.Background())
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		cancel()
	}()

	if err := listener.Run(ctx); err != nil {
		logger.Fatal("listener error", zap.Error(err))
	}
}
