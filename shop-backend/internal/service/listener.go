package service

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/shophub/shop/internal/model"
	"github.com/shophub/shop/internal/repository"
)

const (
	// keccak256("Transfer(address,address,uint256)")
	transferEventTopic = "0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"
	pollInterval       = 15 * time.Second
	lookbackBlocks     = 5  // start a few blocks back to catch very recent txs on startup
	maxBlockRange      = 9  // Alchemy Free tier: max 10 blocks per eth_getLogs request
)

type Listener struct {
	rpcURL       string
	shopWallet   string
	contractAddr string
	backendURL   string
	internalKey  string
	orderRepo    repository.OrderRepository
	logger       *zap.Logger
	lastBlock    uint64
	httpClient   *http.Client
}

func NewListener(
	rpcURL, shopWallet, contractAddr, backendURL, internalKey string,
	orderRepo repository.OrderRepository,
	logger *zap.Logger,
) *Listener {
	// wss:// → https:// — listener polls via HTTP, no WebSocket needed
	rpcURL = strings.Replace(rpcURL, "wss://", "https://", 1)
	return &Listener{
		rpcURL:       rpcURL,
		shopWallet:   strings.ToLower(shopWallet),
		contractAddr: strings.ToLower(contractAddr),
		backendURL:   strings.TrimRight(backendURL, "/"),
		internalKey:  internalKey,
		orderRepo:    orderRepo,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		logger:       logger,
	}
}

func (l *Listener) Run(ctx context.Context) error {
	l.logger.Info("blockchain listener starting",
		zap.String("contract", l.contractAddr),
		zap.String("shop_wallet", l.shopWallet),
	)

	currentBlock, err := l.getBlockNumber()
	if err != nil {
		return fmt.Errorf("get initial block number: %w", err)
	}
	if currentBlock > lookbackBlocks {
		l.lastBlock = currentBlock - lookbackBlocks
	}
	l.logger.Info("scanning from block", zap.Uint64("block", l.lastBlock))

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.logger.Info("blockchain listener stopped")
			return nil
		case <-ticker.C:
			if err := l.poll(ctx); err != nil {
				l.logger.Warn("poll failed", zap.Error(err))
			}
		}
	}
}

func (l *Listener) poll(ctx context.Context) error {
	current, err := l.getBlockNumber()
	if err != nil {
		return err
	}
	if current <= l.lastBlock {
		return nil
	}

	// Cap to maxBlockRange per request (Alchemy Free tier limit).
	toBlock := current
	if toBlock > l.lastBlock+maxBlockRange {
		toBlock = l.lastBlock + maxBlockRange
	}

	from := fmt.Sprintf("0x%x", l.lastBlock+1)
	to := fmt.Sprintf("0x%x", toBlock)

	logs, err := l.getLogs(from, to)
	if err != nil {
		return err
	}

	if len(logs) > 0 {
		l.logger.Info("transfer events found",
			zap.Uint64("from_block", l.lastBlock+1),
			zap.Uint64("to_block", current),
			zap.Int("count", len(logs)),
		)
	}

	for _, log := range logs {
		l.processLog(ctx, log)
	}
	l.lastBlock = toBlock
	return nil
}

type ethLog struct {
	Address         string   `json:"address"`
	Topics          []string `json:"topics"`
	Data            string   `json:"data"`
	TransactionHash string   `json:"transactionHash"`
}

func (l *Listener) processLog(ctx context.Context, log ethLog) {
	if len(log.Topics) < 3 {
		return
	}

	fromAddr := topicToAddress(log.Topics[1])
	toAddr := topicToAddress(log.Topics[2])

	if !strings.EqualFold(toAddr, l.shopWallet) {
		return
	}

	valueWei := hexDataToInt(log.Data)
	l.logger.Info("transfer to shop wallet detected",
		zap.String("from", fromAddr),
		zap.String("txHash", log.TransactionHash),
	)

	status := model.OrderStatusPendingPayment
	orders, _, err := l.orderRepo.List(ctx, repository.OrderFilter{
		Status:   &status,
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		l.logger.Error("list pending orders", zap.Error(err))
		return
	}

	for _, order := range orders {
		if order.WalletFrom == nil || !strings.EqualFold(*order.WalletFrom, fromAddr) {
			continue
		}
		if !valueCoversTotal(valueWei, order.Total) {
			continue
		}

		if err := l.confirmOrder(order.ID.String(), log.TransactionHash); err != nil {
			l.logger.Error("confirm order via backend",
				zap.Error(err),
				zap.String("orderID", order.ID.String()),
			)
		} else {
			l.logger.Info("order confirmed by listener",
				zap.String("orderID", order.ID.String()),
				zap.String("txHash", log.TransactionHash),
			)
		}
		return
	}

	l.logger.Warn("no matching PENDING order for transfer",
		zap.String("from", fromAddr),
		zap.String("txHash", log.TransactionHash),
	)
}

func (l *Listener) confirmOrder(orderID, txHash string) error {
	body, _ := json.Marshal(map[string]string{"txHash": txHash})
	url := fmt.Sprintf("%s/api/v1/internal/orders/%s/confirm", l.backendURL, orderID)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Internal-Key", l.internalKey)

	resp, err := l.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("backend returned HTTP %d", resp.StatusCode)
	}
	return nil
}

// --- JSON-RPC helpers ---

type rpcRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type rpcResponse struct {
	Result json.RawMessage `json:"result"`
	Error  *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (l *Listener) jsonRPC(method string, params any) (json.RawMessage, error) {
	body, err := json.Marshal(rpcRequest{JSONRPC: "2.0", Method: method, Params: params, ID: 1})
	if err != nil {
		return nil, err
	}
	resp, err := l.httpClient.Post(l.rpcURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, fmt.Errorf("rpc: %s", rpcResp.Error.Message)
	}
	return rpcResp.Result, nil
}

func (l *Listener) getBlockNumber() (uint64, error) {
	result, err := l.jsonRPC("eth_blockNumber", []any{})
	if err != nil {
		return 0, err
	}
	var hexBlock string
	if err := json.Unmarshal(result, &hexBlock); err != nil {
		return 0, err
	}
	return parseHexUint64(hexBlock)
}

func (l *Listener) getLogs(fromBlock, toBlock string) ([]ethLog, error) {
	// topics[2] matches `to` address (padded to 32 bytes)
	walletTopic := addressToTopic(l.shopWallet)
	params := []any{
		map[string]any{
			"address":   l.contractAddr,
			"topics":    []any{transferEventTopic, nil, walletTopic},
			"fromBlock": fromBlock,
			"toBlock":   toBlock,
		},
	}
	result, err := l.jsonRPC("eth_getLogs", params)
	if err != nil {
		return nil, err
	}
	var logs []ethLog
	if err := json.Unmarshal(result, &logs); err != nil {
		return nil, err
	}
	return logs, nil
}

// --- address / value helpers ---

func topicToAddress(topic string) string {
	topic = strings.TrimPrefix(topic, "0x")
	if len(topic) < 40 {
		return ""
	}
	return "0x" + strings.ToLower(topic[len(topic)-40:])
}

func addressToTopic(addr string) string {
	addr = strings.ToLower(strings.TrimPrefix(addr, "0x"))
	return "0x" + strings.Repeat("0", 64-len(addr)) + addr
}

func hexDataToInt(hexStr string) *big.Int {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	b, _ := hex.DecodeString(hexStr)
	return new(big.Int).SetBytes(b)
}

func parseHexUint64(hexStr string) (uint64, error) {
	hexStr = strings.TrimPrefix(hexStr, "0x")
	return strconv.ParseUint(hexStr, 16, 64)
}

// valueCoversTotal returns true when valueWei (18-decimal token) is >= order total.
// Allows 1% downward tolerance for floating-point rounding in the frontend.
func valueCoversTotal(valueWei *big.Int, total float64) bool {
	totalF := new(big.Float).SetFloat64(total)
	ten18 := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	totalF.Mul(totalF, ten18)

	expectedWei, _ := totalF.Int(nil)

	// minRequired = expectedWei * 99 / 100
	minRequired := new(big.Int).Mul(expectedWei, big.NewInt(99))
	minRequired.Div(minRequired, big.NewInt(100))

	return valueWei.Cmp(minRequired) >= 0
}
