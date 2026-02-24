package utils

import (
	"fmt"
	"github.com/prestonTao/upnp"
)

// SetupUPnP attempts to automatically forward Port 8545 on the router
func SetupUPnP(port int) {
	fmt.Println("[NETWORK] Attempting to auto-map Port 8545 via UPnP...")
	
	mapping := new(upnp.Upnp)
	if err := mapping.AddPortMapping(port, port, "TCP"); err != nil {
		fmt.Printf("[NETWORK] UPnP Error: %v\n", err)
		fmt.Println("[NETWORK] Please ensure UPnP is enabled in your Huawei router settings.")
		return
	}
	
	fmt.Printf("[NETWORK] Successfully mapped Port %d to your computer! 🚀\n", port)
}
