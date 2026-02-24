package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

const DeveloperAddress = "0xA048F7cfFb548B05eA90ab94962ED0e9A7fC865b"
const FeePercentage = 0.0001 // 0.01%

type Chain struct {
	Blocks     []*Block           `json:"blocks"`
	Difficulty int                `json:"difficulty"`
	Mempool    []Transaction      `json:"mempool"`
	Balances   map[string]float64 `json:"balances"`
	mu         sync.Mutex
}


func NewChain(difficulty int) *Chain {
	genesisBlock := NewBlock(0, "0", []Transaction{}, difficulty)
	// genesisBlock.Mine() // Avoid mining on init to keep it fast if needed
	return &Chain{
		Blocks:     []*Block{genesisBlock},
		Difficulty: difficulty,
		Balances:   make(map[string]float64),
	}
}


func (c *Chain) GetLatestBlock() *Block {
	return c.Blocks[len(c.Blocks)-1]
}

func (c *Chain) AddBlock(block *Block) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	latest := c.GetLatestBlock()
	if block.PrevHash != latest.Hash {
		return errors.New("invalid previous hash")
	}

	if !block.ValidateHash() {
		return errors.New("invalid block hash or difficulty")
	}

	// Update Balances
	for _, tx := range block.Transactions {
		isSystemRoute := (tx.Sender == "SYSTEM" || tx.Sender == "FAUCET" || tx.Sender == DeveloperAddress)
		
		if tx.Sender != "SYSTEM" {
			c.Balances[tx.Sender] -= tx.Amount
		}

		if isSystemRoute {
			c.Balances[tx.Receiver] += tx.Amount
		} else {
			// Apply 0.01% developer fee to regular user transactions
			fee := tx.Amount * FeePercentage
			netAmount := tx.Amount - fee
			
			c.Balances[tx.Receiver] += netAmount
			c.Balances[DeveloperAddress] += fee
		}
	}

	c.Blocks = append(c.Blocks, block)
	return nil
}


func (c *Chain) MinePendingTransactions(minerAddress string, stakerAddress string, treasuryAddress string) {
	// Total Reward: 10 ZAR
	totalReward := 10.0
	devFee := totalReward * FeePercentage
	remainingReward := totalReward - devFee
	
	minerReward := remainingReward * 0.60
	stakerReward := remainingReward * 0.30
	treasuryReward := remainingReward * 0.10

	rewards := []Transaction{
		{ID: fmt.Sprintf("miner-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: minerAddress, Amount: minerReward},
		{ID: fmt.Sprintf("staker-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: stakerAddress, Amount: stakerReward},
		{ID: fmt.Sprintf("treasury-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: treasuryAddress, Amount: treasuryReward},
		{ID: fmt.Sprintf("dev-fee-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: DeveloperAddress, Amount: devFee},
	}
	
	txs := append(c.Mempool, rewards...)
	c.Mempool = []Transaction{}

	newBlock := NewBlock(int64(len(c.Blocks)), c.GetLatestBlock().Hash, txs, c.Difficulty)
	newBlock.Mine()
	c.AddBlock(newBlock)

	// Adjust difficulty every 10 blocks (for testing) or 100 for production
	if len(c.Blocks)%10 == 0 {
		c.AdjustDifficulty()
	}

	c.SaveToFile()
}

func (c *Chain) AdjustDifficulty() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Simple logic: increase difficulty as height grows
	// To make it professional, compare actual mine time vs target time
	c.Difficulty++
	fmt.Printf("[NETWORK] Difficulty increased to: %d\n", c.Difficulty)
}


func (c *Chain) SaveToFile() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile("chaindata.json", data, 0644)
}

func LoadChain(difficulty int) *Chain {
	data, err := os.ReadFile("chaindata.json")
	if err != nil {
		// If file doesn't exist, return a new chain
		return NewChain(difficulty)
	}

	var chain Chain
	if err := json.Unmarshal(data, &chain); err != nil {
		return NewChain(difficulty)
	}

	return &chain
}

