package gateway

import (
	"fmt"
	"strings"
	"sync"
	"time"
	"zar-blockchain/pkg/blockchain"
)

type BridgeOrder struct {
	ID             string  `json:"id"`
	Chain          string  `json:"chain"`          // BTC, ETH, SOL
	DepositAddress string  `json:"depositAddress"` // Address user sends crypto to
	ZARAddress     string  `json:"zarAddress"`      // User's MetaMask address
	Status         string  `json:"status"`          // pending, completed, expired
	AmountIn       float64 `json:"amountIn"`        // External crypto amount
	AmountOut      float64 `json:"amountOut"`       // ZAR amount minted
	CreatedAt      int64   `json:"createdAt"`
}

type Gateway struct {
	Chain             *blockchain.Chain
	Fee               float64
	Oracle            *PriceOracle
	ExternalReceivers map[string]string       // Maps Deposit Address -> User's ZAR Address
	BridgeOrders      map[string]*BridgeOrder // Maps Order ID -> BridgeOrder
	mu                sync.Mutex
}

var SupportedChains = map[string]string{
	"BTC":   "bitcoin",
	"ETH":   "ethereum",
	"SOL":   "solana",
	"TRX":   "tron",
	"BNB":   "binancecoin",
	"LTC":   "litecoin",
	"DOGE":  "dogecoin",
	"MATIC": "matic-network",
	"XMR":   "monero",
	"XRP":   "ripple",
	"ADA":   "cardano",
	"PEPE":  "pepe",
	"CELO":  "celo",
	"AVAX":  "avalanche-2",
	"DOT":   "polkadot",
	"LINK":  "chainlink",
	"SHIB":  "shiba-inu",
	"UNI":   "uniswap",
	"APT":   "aptos",
	"SUI":   "sui",
	"NEAR":  "near",
	"FTM":   "fantom",
	"ATOM":  "cosmos",
	"OP":    "optimism",
	"ARB":   "arbitrum",
}

func NewGateway(chain *blockchain.Chain, fee float64) *Gateway {
	return &Gateway{
		Chain:             chain,
		Fee:               fee,
		Oracle:            NewPriceOracle(),
		ExternalReceivers: make(map[string]string),
		BridgeOrders:      make(map[string]*BridgeOrder),
	}
}

// GenerateReceiver generates a "deposit address" for a specific chain (BTC, ETH, SOL, etc.)
// and links it to the user's ZAR (MetaMask) address. Returns the order ID.
func (g *Gateway) GenerateReceiver(externalChain string, zarAddress string) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	chain := strings.ToUpper(externalChain)
	orderID := fmt.Sprintf("bridge-%s-%d", chain, time.Now().UnixNano())
	
	// Simulate a deposit address (in production: derive from HD wallet)
	depositAddr := fmt.Sprintf("%s-RECV-%s-%d", chain, zarAddress[2:8], time.Now().Unix()%10000)
	
	g.ExternalReceivers[depositAddr] = zarAddress
	
	order := &BridgeOrder{
		ID:             orderID,
		Chain:          chain,
		DepositAddress: depositAddr,
		ZARAddress:     zarAddress,
		Status:         "pending",
		CreatedAt:      time.Now().Unix(),
	}
	g.BridgeOrders[orderID] = order

	fmt.Printf("[BRIDGE] New %s bridge order: %s -> %s\n", chain, depositAddr, zarAddress)
	return orderID
}

// GetBridgeOrder returns a bridge order by ID
func (g *Gateway) GetBridgeOrder(orderID string) *BridgeOrder {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.BridgeOrders[orderID]
}

// GetLiveRate returns the current price of a crypto in USD
func (g *Gateway) GetLiveRate(chain string) (float64, error) {
	coinID, ok := SupportedChains[strings.ToUpper(chain)]
	if !ok {
		return 0, fmt.Errorf("unsupported chain: %s", chain)
	}
	return g.Oracle.GetPrice(coinID, "usd")
}

func (g *Gateway) ProcessExternalDeposit(externalChain string, receiverAddr string, amount float64) {
	zarAddress, ok := g.ExternalReceivers[receiverAddr]
	if !ok {
		fmt.Printf("[GATEWAY] Error: Unknown receiver address %s\n", receiverAddr)
		return
	}

	// Fetch live price
	coinID, ok := SupportedChains[strings.ToUpper(externalChain)]
	if !ok {
		fmt.Printf("[GATEWAY] Unsupported chain: %s\n", externalChain)
		return
	}
	
	usdPrice, err := g.Oracle.GetPrice(coinID, "usd")
	if err != nil {
		fmt.Printf("[GATEWAY] Price Error: %v. Using fallback.\n", err)
		usdPrice = 50000.0
	}

	grossAmount := amount * usdPrice
	bridgeFee := grossAmount * g.Fee
	devFee := grossAmount * blockchain.FeePercentage
	netAmount := grossAmount - bridgeFee - devFee

	fmt.Printf("[BRIDGE] %f %s ($%.2f) -> %f ZAR to %s (fee: $%.2f, dev: $%.4f)\n",
		amount, externalChain, grossAmount, netAmount, zarAddress, bridgeFee, devFee)

	// Main payout
	tx := blockchain.Transaction{
		ID:       fmt.Sprintf("bridge-%s-%d", externalChain, time.Now().Unix()),
		Sender:   "BRIDGE",
		Receiver: zarAddress,
		Amount:   netAmount,
	}
	// Developer fee
	txDev := blockchain.Transaction{
		ID:       fmt.Sprintf("bridge-dev-%d", time.Now().Unix()),
		Sender:   "BRIDGE",
		Receiver: blockchain.DeveloperAddress,
		Amount:   devFee,
	}
	g.Chain.Mempool = append(g.Chain.Mempool, tx, txDev)

	// Update bridge order status
	g.mu.Lock()
	for _, order := range g.BridgeOrders {
		if order.DepositAddress == receiverAddr && order.Status == "pending" {
			order.Status = "completed"
			order.AmountIn = amount
			order.AmountOut = netAmount
			break
		}
	}
	g.mu.Unlock()
}


