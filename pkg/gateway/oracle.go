package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type PriceOracle struct {
	BaseURL string
}

func NewPriceOracle() *PriceOracle {
	return &PriceOracle{
		BaseURL: "https://api.coingecko.com/api/v3/simple/price",
	}
}

// GetPrice returns the price of a coin in USD (or ZAR)
// Example: GetPrice("bitcoin", "usd")
func (o *PriceOracle) GetPrice(coinID string, vsCurrency string) (float64, error) {
	url := fmt.Sprintf("%s?ids=%s&vs_currencies=%s", o.BaseURL, coinID, vsCurrency)
	
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result map[string]map[string]float64
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	price, ok := result[coinID][vsCurrency]
	if !ok {
		return 0, fmt.Errorf("price not found for %s in %s", coinID, vsCurrency)
	}

	return price, nil
}
