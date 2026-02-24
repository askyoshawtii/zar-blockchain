package main

import (
	"fmt"
	"zar-blockchain/pkg/blockchain"
	"zar-blockchain/pkg/rpc"
	"zar-blockchain/pkg/wallet"
)


func main() {
	fmt.Println("Starting ZAR Blockchain Node...")

	// Initialize Chain
	chain := blockchain.NewChain(2)
	fmt.Printf("Genesis Block Mined: %s\n", chain.Blocks[0].Hash)


	// Start RPC Server for MetaMask
	rpcServer := rpc.NewRPCServer(chain, 8545)
	rpcServer.Start()

	// Update DuckDNS (Token needs to be set via environment variable)
	// utils.UpdateDuckDNS("zar-chain", os.Getenv("DUCKDNS_TOKEN"))


	// Create a wallet
	w, _ := wallet.NewWallet()

	fmt.Printf("Node Wallet Address: %s\n", w.Address)

	// Public Treasury & Staker Addresses (Example addresses)
	treasuryAddr := "0xTreasuryFundAddress1234567890abcdef"
	stakerAddr := "0xStakerAddress1234567890abcdef"

	// Simple simulation: Mine 2 blocks
	fmt.Println("Mining Block 1...")
	chain.MinePendingTransactions(w.Address, stakerAddr, treasuryAddr)
	fmt.Printf("Block 1 Added. Hash: %s\n", chain.GetLatestBlock().Hash)

	fmt.Println("Mining Block 2...")
	chain.MinePendingTransactions(w.Address, stakerAddr, treasuryAddr)
	fmt.Printf("Block 2 Added. Hash: %s\n", chain.GetLatestBlock().Hash)


	fmt.Printf("\nBlockchain Length: %d\n", len(chain.Blocks))
	for _, block := range chain.Blocks {
		receiver := "None"
		if len(block.Transactions) > 0 {
			receiver = block.Transactions[0].Receiver
		}
		fmt.Printf("Block %d: %s | Reward for: %s\n", block.Index, block.Hash, receiver)
	}

	fmt.Println("\nNode is running. Press Ctrl+C to stop.")
	select {}
}

