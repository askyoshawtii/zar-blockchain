package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/crypto"
)


type Wallet struct {
	PrivateKey *ecdsa.PrivateKey
	PublicKey  []byte
	Address    string
}

func NewWallet() (*Wallet, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	publicKey := crypto.FromECDSAPub(&privateKey.PublicKey)
	address := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()

	return &Wallet{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		Address:    address,
	}, nil
}

func (w *Wallet) Sign(data []byte) (string, error) {
	hash := crypto.Keccak256Hash(data)
	signature, err := crypto.Sign(hash.Bytes(), w.PrivateKey)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(signature), nil
}

func VerifySignature(address string, data []byte, sigHex string) bool {
	sig, err := hex.DecodeString(sigHex)
	if err != nil {
		return false
	}

	hash := crypto.Keccak256Hash(data)
	pubKey, err := crypto.SigToPub(hash.Bytes(), sig)
	if err != nil {
		return false
	}

	recoveredAddr := crypto.PubkeyToAddress(*pubKey).Hex()
	return recoveredAddr == address
}
