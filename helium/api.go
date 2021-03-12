package helium

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var httpClient = &http.Client{Timeout: 10 * time.Second}

func GetReward(ctx context.Context, address string, days int, callback func(*Reward, error)) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for c := ticker; ; <-c.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		max := time.Now()
		min := time.Now().AddDate(0, 0, days)

		url := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s/rewards/sum?max_time=%s&min_time=%s", address, max.Format("2006-01-02T15:04:05-0700"), min.Format("2006-01-02T15:04:05-0700"))

		var response *RewardResponse
		err := getJson(url, &response)

		callback(response.Reward, err)
	}
}

func GetHotspot(ctx context.Context, address string, callback func(hotspot *Hotspot, err error)) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for c := ticker; ; <-c.C {
		select {
		case <-ctx.Done():
			return
		default:
		}

		url := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s", address)

		response := new(HotspotResponse)
		err := getJson(url, &response)

		callback(response.Hotspot, err)
	}
}

func getJson(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %s: %w", url, err)
	}
	defer r.Body.Close()

	if r.StatusCode != http.StatusOK {
		return fmt.Errorf("http client %d: %s for %v", r.StatusCode, r.Status, url)
	}

	return json.NewDecoder(r.Body).Decode(target)
}
