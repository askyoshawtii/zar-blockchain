package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"zar-blockchain/pkg/blockchain"
)


type RPCServer struct {
	Chain *blockchain.Chain
	Port  int
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
	return &RPCServer{Chain: chain, Port: port}
}

func (s *RPCServer) Start() {
	http.HandleFunc("/", s.handleRPC)
	fmt.Printf("JSON-RPC Server starting on :%d\n", s.Port)
	go http.ListenAndServe(fmt.Sprintf(":%d", s.Port), nil)
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
	case "eth_chainId":
		result = "0x7a5" // 1957 in hex
	case "eth_blockNumber":
		result = fmt.Sprintf("0x%x", len(s.Chain.Blocks)-1)

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
		balance := s.Chain.Balances[addr]
		// Convert ZAR to Wei-ish (18 decimals)
		// 1 ZAR = 10^18 units
		weiBalance := int64(balance * 1e18)
		result = fmt.Sprintf("0x%x", weiBalance)
	case "net_version":
		result = "1957"
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

		fmt.Printf("[FAUCET] Sending %f ZAR to %s and %f ZAR fee to developer\n", userAmount, addr, devFee)
		
		txUser := blockchain.Transaction{
			ID:       fmt.Sprintf("faucet-%d", time.Now().Unix()),
			Sender:   "FAUCET",
			Receiver: addr,
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
		result = fmt.Sprintf("Success! %f ZAR sent to your address (%f ZAR dev fee applied).", userAmount, devFee)
	default:
		rpcErr = map[string]interface{}{
			"code":    -32601,
			"message": "Method not found",
		}
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
