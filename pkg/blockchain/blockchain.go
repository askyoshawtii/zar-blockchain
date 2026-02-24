package blockchain

import (
	"errors"
	"fmt"
	"sync"
)

type Chain struct {
	Blocks     []*Block
	Difficulty int
	Mempool    []Transaction
	mu         sync.Mutex
}

func NewChain(difficulty int) *Chain {
	genesisBlock := NewBlock(0, "0", []Transaction{}, difficulty)
	genesisBlock.Mine()
	return &Chain{
		Blocks:     []*Block{genesisBlock},
		Difficulty: difficulty,
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

	c.Blocks = append(c.Blocks, block)
	return nil
}

func (c *Chain) MinePendingTransactions(minerAddress string, stakerAddress string, treasuryAddress string) {
	// Total Reward: 60 ZAR
	totalReward := 60.0
	minerReward := totalReward * 0.60
	stakerReward := totalReward * 0.30
	treasuryReward := totalReward * 0.10

	rewards := []Transaction{
		{ID: fmt.Sprintf("miner-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: minerAddress, Amount: minerReward},
		{ID: fmt.Sprintf("staker-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: stakerAddress, Amount: stakerReward},
		{ID: fmt.Sprintf("treasury-reward-%d", len(c.Blocks)), Sender: "SYSTEM", Receiver: treasuryAddress, Amount: treasuryReward},
	}
	
	txs := append(c.Mempool, rewards...)
	c.Mempool = []Transaction{}

	newBlock := NewBlock(int64(len(c.Blocks)), c.GetLatestBlock().Hash, txs, c.Difficulty)
	newBlock.Mine()
	c.AddBlock(newBlock)
}

