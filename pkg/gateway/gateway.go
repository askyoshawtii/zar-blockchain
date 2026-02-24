package gateway

import (
	"fmt"
	"time"
	"zar-blockchain/pkg/blockchain"
)


type Gateway struct {
	Chain            *blockchain.Chain
	Fee              float64
	Oracle           *PriceOracle
	ExternalReceivers map[string]string // Maps External Address -> User's ZAR Address
}

func NewGateway(chain *blockchain.Chain, fee float64) *Gateway {
	return &Gateway{
		Chain:             chain,
		Fee:               fee,
		Oracle:            NewPriceOracle(),
		ExternalReceivers: make(map[string]string),
	}
}

// GenerateReceiver generates a "deposit address" for a specific chain (BTC, SOL, etc.)
// and links it to the user's ZAR (MetaMask) address.
func (g *Gateway) GenerateReceiver(externalChain string, zarAddress string) string {
	// In a real system, this would derive a real BTC/SOL address from a HD Wallet.
	// For now, we simulate a unique receiver address.
	receiverAddr := fmt.Sprintf("%s-RECV-%s", externalChain, zarAddress[2:8])
	g.ExternalReceivers[receiverAddr] = zarAddress
	fmt.Printf("[GATEWAY] Generated %s Receiver: %s for user %s\n", externalChain, receiverAddr, zarAddress)
	return receiverAddr
}

func (g *Gateway) ProcessExternalDeposit(externalChain string, receiverAddr string, amount float64) {
	zarAddress, ok := g.ExternalReceivers[receiverAddr]
	if !ok {
		fmt.Printf("[GATEWAY] Error: Unknown receiver address %s\n", receiverAddr)
		return
	}

	// Fetch live price
	coinIDMap := map[string]string{"BTC": "bitcoin", "SOL": "solana", "TRX": "tron"}
	usdPrice, err := g.Oracle.GetPrice(coinIDMap[externalChain], "usd")
	if err != nil {
		fmt.Printf("[GATEWAY] Price Error: %v. Using fallback price.\n", err)
		usdPrice = 50000.0 // Fallback
	}

	netAmount := (amount * usdPrice) * (1 - g.Fee)
	
	fmt.Printf("[GATEWAY] Detected %f %s deposit. Minting %f ZAR to %s\n", amount, externalChain, netAmount, zarAddress)

	tx := blockchain.Transaction{
		ID:       fmt.Sprintf("swap-%s-%d", externalChain, time.Now().Unix()),
		Sender:   "GATEWAY",
		Receiver: zarAddress,
		Amount:   netAmount,
	}
	g.Chain.Mempool = append(g.Chain.Mempool, tx)
}

