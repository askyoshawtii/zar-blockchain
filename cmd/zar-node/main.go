package main

import (
	"fmt"
	"os"
	"time"
	"zar-blockchain/pkg/blockchain"
	"zar-blockchain/pkg/gateway"
	"zar-blockchain/pkg/rpc"
	"zar-blockchain/pkg/utils"

	"zar-blockchain/pkg/wallet"
)






func main() {
	fmt.Println("Starting ZAR Blockchain Node...")

	// Initialize Chain (Load from disk if exists)
	chain := blockchain.LoadChain(2)
	fmt.Printf("Current Blockchain Height: %d\n", len(chain.Blocks))
	fmt.Printf("Latest Block Hash: %s\n", chain.GetLatestBlock().Hash)

	// Automated Port Forwarding (UPnP)
	utils.SetupUPnP(8545)

	// Initialize Universal Gateway (Bridge)
	gw := gateway.NewGateway(chain, 0.01) // 1% Bridge Fee

	// Start RPC Server for MetaMask + Bridge
	rpcServer := rpc.NewRPCServer(chain, gw, 8545)

	domain := os.Getenv("DUCKDNS_DOMAIN")
	token := os.Getenv("DUCKDNS_TOKEN")

	if domain != "" && token != "" {
		fmt.Println("[SSL] DuckDNS credentials found. Activating Automated SSL (HTTPS)...")
		fqdn := domain + ".duckdns.org"
		rpcServer.StartTLS(fqdn, token)
		
		// Also start the IP update heartbeat
		go utils.UpdateDuckDNS(domain, token)
	} else {
		fmt.Println("[RPC] Starting standard HTTP server (No SSL credentials found).")
		rpcServer.Start()
	}





	// Create a wallet
	w, _ := wallet.NewWallet()

	fmt.Printf("Node Wallet Address: %s\n", w.Address)

	// YOUR METAMASK ADDRESS
	myMetaMaskAddr := "0xA048F7cfFb548B05eA90ab94962ED0e9A7fC865b" 

	// Public Treasury & Staker Addresses
	treasuryAddr := "0xTreasuryFundAddress1234567890abcdef"
	stakerAddr := "0xStakerAddress1234567890abcdef"

	// START REAL CONTINUOUS MINING
	fmt.Println("\n[MINER] Starting Background Mining Loop...")
	go func() {
		for {
			fmt.Println("[MINER] Mining next block...")
			chain.MinePendingTransactions(myMetaMaskAddr, stakerAddr, treasuryAddr)
			fmt.Printf("[MINER] Block Mined! Height: %d | Hash: %s\n", len(chain.Blocks), chain.GetLatestBlock().Hash)
			
			// Wait 10 seconds between blocks (adjust for difficulty)
			time.Sleep(10 * time.Second)
		}
	}()


	// Start the Auto-Detector (Scanner)
	scanner := gateway.NewScanner(gw)
	scanner.Start()

	fmt.Println("\n--- Universal Gateway (Bridge) Activated ---")
	fmt.Println("Supported Chains: BTC, ETH, SOL, TRX, BNB")
	btcReceiver := gw.GenerateReceiver("BTC", myMetaMaskAddr)
	fmt.Printf("Send Bitcoin to: %s to receive ZAR Coin!\n", btcReceiver)

	fmt.Println("\nNode is running. Press Ctrl+C to stop.")
	select {}
}





