package blockchain

import (
	"fmt"
	"strings"
)

func (b *Block) Mine() {
	target := strings.Repeat("0", b.Difficulty)
	for {
		b.Hash = b.CalculateHash()
		if strings.HasPrefix(b.Hash, target) {
			fmt.Printf("\rBlock Mined! Hash: %s\n", b.Hash)
			break
		}
		b.Nonce++
	}
}

func (b *Block) ValidateHash() bool {
	target := strings.Repeat("0", b.Difficulty)
	return b.Hash == b.CalculateHash() && strings.HasPrefix(b.Hash, target)
}

// SelectValidator chooses a validator from a list based on their stake.
// Simple version: Choose based on stake weight.
func SelectValidator(stakes map[string]float64) string {
	var totalStake float64
	for _, stake := range stakes {
		totalStake += stake
	}

	if totalStake == 0 {
		return ""
	}

	// Pseudo-random selection (deterministic for demo purposes)
	// In a real blockchain, this would use a verifiable random function (VRF)
	var winner string
	var currentRank float64
	for addr, stake := range stakes {
		if stake > currentRank {
			currentRank = stake
			winner = addr
		}
	}
	return winner
}

