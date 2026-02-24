package main

import (
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




	// Start RPC Server for MetaMask
	rpcServer := rpc.NewRPCServer(chain, 8545)
	rpcServer.Start()

	// Update DuckDNS (Token and Domain needs to be set via environment variables)
	if os.Getenv("DUCKDNS_TOKEN") != "" && os.Getenv("DUCKDNS_DOMAIN") != "" {
		utils.UpdateDuckDNS(os.Getenv("DUCKDNS_DOMAIN"), os.Getenv("DUCKDNS_TOKEN"))
	} else {
		fmt.Println("[WARN] DuckDNS Domain or Token not set. Remote access might be unstable.")
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


	// Demonstration: Universal Gateway
	gw := gateway.NewGateway(chain, 0.01) // 1% Gateway Fee
	
	// Start the Auto-Detector (Scanner)
	scanner := gateway.NewScanner(gw)
	scanner.Start()

	fmt.Println("\n--- Universal Gateway Activated ---")
	// 1. User generates a BTC receiver address
	btcReceiver := gw.GenerateReceiver("BTC", myMetaMaskAddr)
	fmt.Printf("Send Bitcoin to: %s to receive ZAR Coin!\n", btcReceiver)

	fmt.Println("\nNode is running. Press Ctrl+C to stop.")
	select {}
}




