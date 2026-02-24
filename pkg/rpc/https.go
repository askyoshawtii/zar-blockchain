package rpc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/caddyserver/certmagic"
	"github.com/libdns/duckdns"
)

// StartTLS initializes the RPC server with automatic SSL via CertMagic and DuckDNS
func (s *RPCServer) StartTLS(domain, token string) {
	fmt.Printf("[SSL] Activating Automated SSL for %s...\n", domain)

	// Configure DuckDNS provider for DNS-01 challenge (bypasses port 80 restrictions)
	provider := &duckdns.Provider{
		APIToken: token,
	}

	// Configure CertMagic
	certmagic.DefaultACME.Agreed = true
	certmagic.DefaultACME.Email = "admin@zar-chain.org" // Used for SSL registration
	certmagic.DefaultACME.DNS01Solver = &certmagic.DNS01Solver{
		DNSManager: certmagic.DNSManager{
			DNSProvider: provider,
		},
	}

	magic := certmagic.NewDefault()
	err := magic.ManageSync(context.Background(), []string{domain})
	if err != nil {
		fmt.Printf("[SSL] Certificate Management Error: %v\n", err)
		return
	}

	// Create the secure server
	tlsConfig := magic.TLSConfig()
	srv := &http.Server{
		Addr:      fmt.Sprintf(":%d", s.Port),
		TLSConfig: tlsConfig,
		Handler:   http.HandlerFunc(s.handleRPC),
	}

	fmt.Printf("[SSL] JSON-RPC Secure Server (HTTPS) starting on :%d\n", s.Port)
	go func() {
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			fmt.Printf("[SSL] Server Error: %v\n", err)
		}
	}()
}
