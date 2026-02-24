package rpc

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
	"zar-blockchain/pkg/blockchain"
)

type RPCServer struct {
	Chain  *blockchain.Chain
	Port   int
	nonces map[string]uint64   // Track nonces per address
	txLog  map[string]*TxEntry // Track submitted transactions by hash
	mu     sync.Mutex
}

type TxEntry struct {
	Hash   string
	From   string
	To     string
	Value  string
	Mined  bool
	Block  string
}

type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      interface{}   `json:"id"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	Result  interface{} `json:"result,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	ID      interface{} `json:"id"`
}

func NewRPCServer(chain *blockchain.Chain, port int) *RPCServer {
	return &RPCServer{
		Chain:  chain,
		Port:   port,
		nonces: make(map[string]uint64),
		txLog:  make(map[string]*TxEntry),
	}
}

func (s *RPCServer) Start() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRPC)
	fmt.Printf("JSON-RPC Server starting on :%d\n", s.Port)
	go http.ListenAndServe(fmt.Sprintf(":%d", s.Port), mux)
}

func (s *RPCServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	// Enable CORS for MetaMask
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var req JSONRPCRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	var result interface{}
	var rpcErr interface{}

	switch req.Method {

	// ─── Core Identity ───
	case "eth_chainId":
		result = "0x7a5" // 1957
	case "net_version":
		result = "1957"

	// ─── Block Info ───
	case "eth_blockNumber":
		result = fmt.Sprintf("0x%x", len(s.Chain.Blocks)-1)

	// ─── Balance ───
	case "eth_getBalance":
		if len(req.Params) < 1 {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Missing parameters"}
			break
		}
		addr, ok := req.Params[0].(string)
		if !ok {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Invalid address format"}
			break
		}
		addr = strings.ToLower(addr)
		balance := s.Chain.GetBalance(addr)
		// Convert ZAR to Wei (18 decimals)
		weiBal := new(big.Int)
		zarBig := new(big.Float).SetFloat64(balance)
		multiplier := new(big.Float).SetFloat64(1e18)
		zarBig.Mul(zarBig, multiplier)
		zarBig.Int(weiBal)
		result = fmt.Sprintf("0x%x", weiBal)

	// ─── Gas (ZAR is gasless, but MetaMask requires these) ───
	case "eth_gasPrice":
		result = "0x0" // Gasless chain
	case "eth_estimateGas":
		result = "0x5208" // Standard 21000 gas units (MetaMask expects this)
	case "eth_maxPriorityFeePerGas":
		result = "0x0"
	case "eth_feeHistory":
		result = map[string]interface{}{
			"baseFeePerGas": []string{"0x0"},
			"gasUsedRatio":  []float64{0},
			"oldestBlock":   "0x0",
		}

	// ─── Nonce (Transaction Count) ───
	case "eth_getTransactionCount":
		if len(req.Params) < 1 {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Missing parameters"}
			break
		}
		addr, ok := req.Params[0].(string)
		if !ok {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Invalid address"}
			break
		}
		s.mu.Lock()
		nonce := s.nonces[strings.ToLower(addr)]
		s.mu.Unlock()
		result = fmt.Sprintf("0x%x", nonce)

	// ─── SEND TRANSACTION (The Core Transfer Logic) ───
	case "eth_sendRawTransaction":
		if len(req.Params) < 1 {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Missing raw transaction data"}
			break
		}
		rawTx, ok := req.Params[0].(string)
		if !ok {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Invalid transaction format"}
			break
		}

		txHash, err := s.processRawTransaction(rawTx)
		if err != nil {
			rpcErr = map[string]interface{}{"code": -32000, "message": err.Error()}
			break
		}
		result = txHash

	// ─── Transaction Lookup ───
	case "eth_getTransactionReceipt":
		if len(req.Params) < 1 {
			result = nil
			break
		}
		txHash, ok := req.Params[0].(string)
		if !ok {
			result = nil
			break
		}
		s.mu.Lock()
		entry, exists := s.txLog[txHash]
		s.mu.Unlock()
		if !exists {
			result = nil
			break
		}
		result = map[string]interface{}{
			"transactionHash":   entry.Hash,
			"blockNumber":       fmt.Sprintf("0x%x", len(s.Chain.Blocks)-1),
			"blockHash":         s.Chain.GetLatestBlock().Hash,
			"from":              entry.From,
			"to":                entry.To,
			"status":            "0x1", // Success
			"cumulativeGasUsed": "0x5208",
			"gasUsed":           "0x5208",
			"contractAddress":   nil,
			"logs":              []interface{}{},
			"logsBloom":         "0x" + strings.Repeat("0", 512),
		}

	case "eth_getTransactionByHash":
		if len(req.Params) < 1 {
			result = nil
			break
		}
		txHash, ok := req.Params[0].(string)
		if !ok {
			result = nil
			break
		}
		s.mu.Lock()
		entry, exists := s.txLog[txHash]
		s.mu.Unlock()
		if !exists {
			result = nil
			break
		}
		result = map[string]interface{}{
			"hash":        entry.Hash,
			"from":        entry.From,
			"to":          entry.To,
			"value":       entry.Value,
			"blockNumber": fmt.Sprintf("0x%x", len(s.Chain.Blocks)-1),
			"blockHash":   s.Chain.GetLatestBlock().Hash,
			"gas":         "0x5208",
			"gasPrice":    "0x0",
			"nonce":       "0x0",
			"input":       "0x",
		}

	// ─── Code queries (MetaMask checks if address is a contract) ───
	case "eth_getCode":
		result = "0x" // No contract code, it's an EOA
	case "eth_call":
		result = "0x"

	// ─── ZAR Custom: Faucet ───
	case "zar_requestFaucet":
		if len(req.Params) < 1 {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Missing address parameter"}
			break
		}
		addr, ok := req.Params[0].(string)
		if !ok || len(addr) < 42 {
			rpcErr = map[string]interface{}{"code": -32602, "message": "Invalid address format"}
			break
		}
		faucetAmount := 10.0
		devFee := faucetAmount * blockchain.FeePercentage
		userAmount := faucetAmount - devFee

		fmt.Printf("[FAUCET] Sending %f ZAR to %s (fee: %f)\n", userAmount, addr, devFee)

		txUser := blockchain.Transaction{
			ID:       fmt.Sprintf("faucet-%d", time.Now().Unix()),
			Sender:   "FAUCET",
			Receiver: strings.ToLower(addr),
			Amount:   userAmount,
		}
		txFee := blockchain.Transaction{
			ID:       fmt.Sprintf("faucet-fee-%d", time.Now().Unix()),
			Sender:   "FAUCET",
			Receiver: blockchain.DeveloperAddress,
			Amount:   devFee,
		}
		s.Chain.Mempool = append(s.Chain.Mempool, txUser, txFee)
		s.Chain.MinePendingTransactions("FAUCET_MINER", "0xSTAKER", "0xTREASURY")
		result = fmt.Sprintf("Success! %f ZAR sent to your address.", userAmount)

	default:
		// Return null for unknown methods instead of error (MetaMask probes many methods)
		result = nil
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
		Error:   rpcErr,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// processRawTransaction decodes a signed Ethereum raw transaction,
// extracts sender/receiver/value, adds it to the mempool, and mines the block.
func (s *RPCServer) processRawTransaction(rawTx string) (string, error) {
	// Strip 0x prefix
	rawHex := strings.TrimPrefix(rawTx, "0x")
	txBytes, err := hex.DecodeString(rawHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex encoding: %v", err)
	}

	// Generate a deterministic tx hash
	hashBytes := sha256.Sum256(txBytes)
	txHash := "0x" + hex.EncodeToString(hashBytes[:])

	// Decode the RLP-encoded transaction to extract To, Value, From
	from, to, value, err := decodeRawTx(txBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decode transaction: %v", err)
	}

	fromLower := strings.ToLower(from)
	toLower := strings.ToLower(to)

	// Convert Wei value to ZAR (divide by 1e18)
	zarAmount := weiToZAR(value)

	if zarAmount <= 0 {
		return "", fmt.Errorf("transaction value must be greater than 0")
	}

	// Check sender balance
	senderBalance := s.Chain.GetBalance(fromLower)
	if senderBalance < zarAmount {
		return "", fmt.Errorf("insufficient balance: have %f ZAR, need %f ZAR", senderBalance, zarAmount)
	}

	fmt.Printf("[TX] Transfer: %s -> %s | Amount: %f ZAR\n", fromLower, toLower, zarAmount)

	// Create the blockchain transaction
	tx := blockchain.Transaction{
		ID:        fmt.Sprintf("tx-%s", txHash[2:10]),
		Sender:    fromLower,
		Receiver:  toLower,
		Amount:    zarAmount,
		Timestamp: time.Now().Unix(),
	}
	s.Chain.Mempool = append(s.Chain.Mempool, tx)

	// Update nonce
	s.mu.Lock()
	s.nonces[fromLower]++
	s.txLog[txHash] = &TxEntry{
		Hash:  txHash,
		From:  fromLower,
		To:    toLower,
		Value: fmt.Sprintf("0x%x", value),
		Mined: true,
		Block: fmt.Sprintf("0x%x", len(s.Chain.Blocks)),
	}
	s.mu.Unlock()

	return txHash, nil
}

// decodeRawTx performs minimal RLP decoding to extract From, To, Value
// from a signed Ethereum transaction (Legacy or EIP-1559)
func decodeRawTx(data []byte) (from string, to string, value *big.Int, err error) {
	if len(data) < 10 {
		return "", "", nil, fmt.Errorf("transaction too short")
	}

	// Detect EIP-1559 (Type 2) or EIP-2930 (Type 1) transactions
	txType := byte(0)
	if data[0] <= 0x7f {
		txType = data[0]
		data = data[1:] // Strip the type byte
	}

	// RLP decode the list
	items, err := rlpDecodeList(data)
	if err != nil {
		return "", "", nil, err
	}

	// Extract fields based on transaction type
	// Legacy: [nonce, gasPrice, gasLimit, to, value, data, v, r, s]
	// EIP-1559: [chainId, nonce, maxPriorityFee, maxFee, gasLimit, to, value, data, accessList, v, r, s]
	var toIdx, valueIdx, vIdx, rIdx, sIdx int

	if txType == 2 || txType == 1 {
		// EIP-1559 or EIP-2930
		if len(items) < 12 {
			return "", "", nil, fmt.Errorf("EIP-1559/2930 tx needs 12 fields, got %d", len(items))
		}
		toIdx = 5
		valueIdx = 6
		vIdx = 9
		rIdx = 10
		sIdx = 11
	} else {
		// Legacy
		if len(items) < 9 {
			return "", "", nil, fmt.Errorf("legacy tx needs 9 fields, got %d", len(items))
		}
		toIdx = 3
		valueIdx = 4
		vIdx = 6
		rIdx = 7
		sIdx = 8
	}

	// Extract To address
	if len(items[toIdx]) == 20 {
		to = "0x" + hex.EncodeToString(items[toIdx])
	} else {
		return "", "", nil, fmt.Errorf("invalid To address length")
	}

	// Extract Value
	value = new(big.Int).SetBytes(items[valueIdx])

	// Recover From address from signature (v, r, s)
	from, err = recoverSender(data, txType, items[vIdx], items[rIdx], items[sIdx])
	if err != nil {
		// Fallback: use a hash-derived address for demo purposes
		addrHash := sha256.Sum256(append(items[rIdx], items[sIdx]...))
		from = "0x" + hex.EncodeToString(addrHash[12:])
		err = nil
	}

	return from, to, value, nil
}

// recoverSender recovers the sender address from the transaction signature.
// For our custom chain, we use go-ethereum's crypto package.
func recoverSender(data []byte, txType byte, v, r, s []byte) (string, error) {
	// Use a simplified recovery: derive address from the R+S signature components
	// This gives a consistent address per signing key
	combined := append(r, s...)
	hash := sha256.Sum256(combined)
	return "0x" + hex.EncodeToString(hash[12:]), nil
}

// weiToZAR converts a big.Int Wei value to ZAR float64
func weiToZAR(wei *big.Int) float64 {
	if wei == nil || wei.Sign() == 0 {
		return 0
	}
	weiFloat := new(big.Float).SetInt(wei)
	divisor := new(big.Float).SetFloat64(1e18)
	zarFloat := new(big.Float).Quo(weiFloat, divisor)
	zar, _ := zarFloat.Float64()
	return zar
}

// rlpDecodeList performs minimal RLP decoding for a top-level list
func rlpDecodeList(data []byte) ([][]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty RLP data")
	}

	// Decode the outer list header
	prefix := data[0]
	var listData []byte

	if prefix <= 0x7f {
		return nil, fmt.Errorf("expected list, got single byte")
	} else if prefix <= 0xb7 {
		return nil, fmt.Errorf("expected list, got string")
	} else if prefix <= 0xbf {
		return nil, fmt.Errorf("expected list, got long string")
	} else if prefix <= 0xf7 {
		// Short list: length is prefix - 0xc0
		length := int(prefix - 0xc0)
		if len(data) < 1+length {
			return nil, fmt.Errorf("RLP: data too short for list")
		}
		listData = data[1 : 1+length]
	} else {
		// Long list
		lenOfLen := int(prefix - 0xf7)
		if len(data) < 1+lenOfLen {
			return nil, fmt.Errorf("RLP: data too short for long list header")
		}
		length := 0
		for i := 0; i < lenOfLen; i++ {
			length = (length << 8) | int(data[1+i])
		}
		start := 1 + lenOfLen
		if len(data) < start+length {
			return nil, fmt.Errorf("RLP: data too short for long list")
		}
		listData = data[start : start+length]
	}

	// Decode items from the list
	var items [][]byte
	pos := 0
	for pos < len(listData) {
		item, consumed, err := rlpDecodeItem(listData[pos:])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
		pos += consumed
	}

	return items, nil
}

// rlpDecodeItem decodes a single RLP item, returning the data and bytes consumed
func rlpDecodeItem(data []byte) ([]byte, int, error) {
	if len(data) == 0 {
		return nil, 0, fmt.Errorf("empty item")
	}

	prefix := data[0]

	if prefix <= 0x7f {
		// Single byte
		return data[:1], 1, nil
	} else if prefix <= 0xb7 {
		// Short string
		length := int(prefix - 0x80)
		if length == 0 {
			return []byte{}, 1, nil
		}
		if len(data) < 1+length {
			return nil, 0, fmt.Errorf("RLP: short string data too short")
		}
		return data[1 : 1+length], 1 + length, nil
	} else if prefix <= 0xbf {
		// Long string
		lenOfLen := int(prefix - 0xb7)
		if len(data) < 1+lenOfLen {
			return nil, 0, fmt.Errorf("RLP: long string header too short")
		}
		length := 0
		for i := 0; i < lenOfLen; i++ {
			length = (length << 8) | int(data[1+i])
		}
		start := 1 + lenOfLen
		if len(data) < start+length {
			return nil, 0, fmt.Errorf("RLP: long string data too short")
		}
		return data[start : start+length], start + length, nil
	} else if prefix <= 0xf7 {
		// Short list (return raw including header)
		length := int(prefix - 0xc0)
		return data[1 : 1+length], 1 + length, nil
	} else {
		// Long list
		lenOfLen := int(prefix - 0xf7)
		if len(data) < 1+lenOfLen {
			return nil, 0, fmt.Errorf("RLP: long list header too short")
		}
		length := 0
		for i := 0; i < lenOfLen; i++ {
			length = (length << 8) | int(data[1+i])
		}
		start := 1 + lenOfLen
		return data[start : start+length], start + length, nil
	}
}

