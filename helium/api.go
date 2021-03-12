package helium

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var webClient = &http.Client{Timeout: 10 * time.Second}

func GetReward(adress string, max time.Time, min time.Time) (*RewardResponse, error) {
	url := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s/rewards/sum?max_time=%s&min_time=%s", adress, max.Format("2006-01-02T15:04:05-0700"), min.Format("2006-01-02T15:04:05-0700"))
	var response *RewardResponse

	if err := getJson(url, &response); err != nil {
		return nil, fmt.Errorf("get json: %w", err)
	}

	return response, nil
}

func GetHotspot(address string) (*HotspotResponse, error) {
	url := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s", address)
	var response *HotspotResponse

	if err := getJson(url, &response); err != nil {
		return nil, fmt.Errorf("get json: %w", err)
	}

	return response, nil
}

func getJson(url string, target interface{}) error {
	r, err := webClient.Get(url)
	if err != nil {
		return fmt.Errorf("http get: %s: %w", url, err)
	}
	defer r.Body.Close()

	return json.NewDecoder(r.Body).Decode(target)
}
