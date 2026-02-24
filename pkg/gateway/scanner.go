package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Scanner struct {
	Gateway *Gateway
}

func NewScanner(gw *Gateway) *Scanner {
	return &Scanner{Gateway: gw}
}

func (s *Scanner) Start() {
	fmt.Println("[SCANNER] Auto-Detector started...")
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for range ticker.C {
			s.ScanAll()
		}
	}()
}

func (s *Scanner) ScanAll() {
	for externalAddr, zarAddr := range s.Gateway.ExternalReceivers {
		// Logic: If externalAddr starts with BTC, use Blockstream API
		// If starts with SOL, use Helius, etc.
		if len(externalAddr) > 4 && externalAddr[:3] == "BTC" {
			s.checkBitcoinDeposit(externalAddr, zarAddr)
		}
		if len(externalAddr) > 4 && externalAddr[:3] == "SOL" {
			s.checkSolanaDeposit(externalAddr, zarAddr)
		}
	}
}

// checkSolanaDeposit simulates monitoring the Solana network via Helius
func (s *Scanner) checkSolanaDeposit(solAddr string, zarAddr string) {
	// In production, we would use: https://mainnet.helius-rpc.com/?api-key=<TOKEN>
	// For now, we simulate a detected deposit to show the flow.
	fmt.Printf("[SCANNER] Monitoring Solana Address for user %s...\n", zarAddr)
}


// checkBitcoinDeposit uses Blockstream.info free API (no key)
func (s *Scanner) checkBitcoinDeposit(btcAddr string, zarAddr string) {
	// Note: In simulation, btcAddr is "BTC-RECV-..."
	// In production, this would be a real Bitcoin address like "bc1..."
	url := fmt.Sprintf("https://blockstream.info/api/address/%s/utxo", btcAddr)
	
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var utxos []map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&utxos)

	for _, utxo := range utxos {
		val := utxo["value"].(float64) / 100000000.0 // Satoshis to BTC
		fmt.Printf("[SCANNER] REAL BTC DEPOSIT DETECTED: %f BTC to %s\n", val, btcAddr)
		
		// 1. Process the deposit (Add to mempool)
		s.Gateway.ProcessExternalDeposit("BTC", btcAddr, val)

		// 2. Auto-Mine a new block to "finalize" the ZAR minting
		// In a real network, this would happen when the next miner finds a block.
		// For a single-node setup, we trigger it automatically for UX.
		fmt.Println("[SCANNER] Auto-Mining ZAR block to finalize payout...")
		s.Gateway.Chain.MinePendingTransactions("GATEWAY_RESERVE", "0xSTAKER", "0xTREASURY")
	}
}

