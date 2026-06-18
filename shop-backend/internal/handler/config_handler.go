package handler

import (
	"net/http"

	"github.com/shophub/shop/internal/config"
	"github.com/shophub/shop/pkg/response"
)

type ConfigHandler struct {
	cfg *config.Config
}

func NewConfigHandler(cfg *config.Config) *ConfigHandler {
	return &ConfigHandler{cfg: cfg}
}

type shopConfig struct {
	WalletAddress   string `json:"walletAddress"`
	ContractAddress string `json:"contractAddress"`
}

// GetConfig returns public blockchain configuration needed by the frontend.
func (h *ConfigHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	response.Ok(w, shopConfig{
		WalletAddress:   h.cfg.Blockchain.ShopWalletAddress,
		ContractAddress: h.cfg.Blockchain.MockUSDTAddress,
	})
}
