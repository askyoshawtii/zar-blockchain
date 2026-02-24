package gateway

import (
	"fmt"
	"zar-blockchain/pkg/blockchain"
)

type Gateway struct {
	Chain *blockchain.Chain
	Fee   float64 // Fee percentage (e.g., 0.01 for 1%)
}

func NewGateway(chain *blockchain.Chain, fee float64) *Gateway {
	return &Gateway{Chain: chain, Fee: fee}
}

// DepositNotification simulates a deposit being detected on an external chain.
// externalChain: "BTC", "SOL", "TRX", etc.
// externalAmount: Amount deposited in the external currency.
// targetZARAddress: The user's MetaMask address to receive ZAR.
func (g *Gateway) ProcessExternalDeposit(externalChain string, externalAmount float64, exchangeRate float64, targetZARAddress string) {
	fmt.Printf("[GATEWAY] Deposit detected on %s: %f units\n", externalChain, externalAmount)
	
	// Calculation: (Amount * Rate) - Fee
	rawTotal := externalAmount * exchangeRate
	feeAmount := rawTotal * g.Fee
	netAmount := rawTotal - feeAmount

	fmt.Printf("[GATEWAY] Converting to %f ZAR (Fee: %f)\n", netAmount, feeAmount)

	// In a real implementation, this would trigger a new block or transaction
	tx := blockchain.Transaction{
		ID:       fmt.Sprintf("gw-%s-%f", externalChain, externalAmount),
		Sender:   "GATEWAY_RESERVE",
		Receiver: targetZARAddress,
		Amount:   netAmount,
	}

	g.Chain.Mempool = append(g.Chain.Mempool, tx)
	fmt.Printf("[GATEWAY] Minting %f ZAR to %s\n", netAmount, targetZARAddress)
}
