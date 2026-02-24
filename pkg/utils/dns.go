package utils

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func UpdateDuckDNS(domain, token string) {
	url := fmt.Sprintf("https://www.duckdns.org/update?domains=%s&token=%s&ip=", domain, token)
	
	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("DuckDNS Update Error: %v\n", err)
				continue
			}
			body, _ := io.ReadAll(resp.Body)
			fmt.Printf("DuckDNS Update: %s (Status: %d)\n", string(body), resp.StatusCode)
			resp.Body.Close()
		}
	}()
}
