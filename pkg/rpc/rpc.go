package rpc

import (
	"encoding/json"
	"fmt"
	"net/http"
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
		result = "0x7A5" // Chain ID 1957 (ZAR in hex-ish / arbitrary)
	case "eth_blockNumber":
		result = fmt.Sprintf("0x%x", len(s.Chain.Blocks)-1)
	case "eth_getBalance":
		// Simple mock for now - returns 100 ZAR for any address
		result = "0x56bc75e2d63100000" // 100 * 10^18 (100 ZAR in Wei)
	case "net_version":
		result = "1957"
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
