package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Index        int64         `json:"index"`
	Timestamp    int64         `json:"timestamp"`
	PrevHash     string        `json:"prev_hash"`
	Hash         string        `json:"hash"`
	Transactions []Transaction `json:"transactions"`
	Nonce        int64         `json:"nonce"`
	Difficulty   int           `json:"difficulty"`
	Validator    string        `json:"validator,omitempty"` // PoS Validator Address
	Signature    string        `json:"signature,omitempty"` // Validator Signature
}

type Transaction struct {
	ID        string  `json:"id"`
	Sender    string  `json:"sender"`
	Receiver  string  `json:"receiver"`
	Amount    float64 `json:"amount"`
	Timestamp int64   `json:"timestamp"`
	Signature string  `json:"signature"`
}

func (b *Block) CalculateHash() string {
	data, _ := json.Marshal(struct {
		Index        int64         `json:"index"`
		Timestamp    int64         `json:"timestamp"`
		PrevHash     string        `json:"prev_hash"`
		Transactions []Transaction `json:"transactions"`
		Nonce        int64         `json:"nonce"`
		Validator    string        `json:"validator"`
	}{
		Index:        b.Index,
		Timestamp:    b.Timestamp,
		PrevHash:     b.PrevHash,
		Transactions: b.Transactions,
		Nonce:        b.Nonce,
		Validator:    b.Validator,
	})
	hash := sha256.Sum256(data) // Keeping SHA256 for PoW mining (it's traditional)
	return fmt.Sprintf("%x", hash)
}


func NewBlock(index int64, prevHash string, txs []Transaction, difficulty int) *Block {
	b := &Block{
		Index:        index,
		Timestamp:    time.Now().Unix(),
		PrevHash:     prevHash,
		Transactions: txs,
		Difficulty:   difficulty,
	}
	b.Hash = b.CalculateHash()
	return b
}
