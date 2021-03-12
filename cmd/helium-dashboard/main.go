package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var webClient = &http.Client{Timeout: 10 * time.Second}

func main() {

	app := tview.NewApplication()
	table := tview.NewTable()
	table.SetBorders(true)

	table.SetCell(0, 0, tview.NewTableCell("Hotspot Name").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 1, tview.NewTableCell("last 24h").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 2, tview.NewTableCell("last 7 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 3, tview.NewTableCell("last 30 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 4, tview.NewTableCell("Address").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))
	table.SetCell(0, 5, tview.NewTableCell("Owner").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignCenter))

	addresses := os.Args[1:]

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		table.SetSelectable(false, false)
	})

	go func() {
		for {
			app.QueueUpdateDraw(func() {
				if err := refresh(addresses, table); err != nil {
					panic(err)
				}
			})
			time.Sleep(30 * time.Second)
		}
	}()

	app.SetRoot(table, true).EnableMouse(true)
	if err := app.Run(); err != nil {
		panic(err)
	}
}

func refresh(addresses []string, table *tview.Table) error {

	for i, address := range addresses {
		row := i + 1
		hotspotResponse, err := getHotspot(address)
		if err != nil {
			return fmt.Errorf("getting hotspot: %s: %w", address, err)
		}

		h := hotspotResponse.Data

		table.SetCell(row, 0, tview.NewTableCell(h.Name).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

		cell, err := getRewardCell(address, time.Now(), time.Now().AddDate(0, 0, -1))
		if err != nil {
			return fmt.Errorf("reward 24h: %w", err)
		}
		table.SetCell(row, 1, cell)
		cell, err = getRewardCell(address, time.Now(), time.Now().AddDate(0, 0, -7))
		if err != nil {
			return fmt.Errorf("reward 7 days: %w", err)
		}
		table.SetCell(row, 2, cell)
		cell, err = getRewardCell(address, time.Now(), time.Now().AddDate(0, 0, -30))
		if err != nil {
			return fmt.Errorf("reward 30 days: %w", err)
		}
		table.SetCell(row, 3, cell)

		table.SetCell(row, 4, tview.NewTableCell(h.Address).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		table.SetCell(row, 5, tview.NewTableCell(h.Address).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
	}

	return nil
}

func getRewardCell(address string, max time.Time, min time.Time) (*tview.TableCell, error) {
	rewardResponse, err := getReward(address, max, min)
	if err != nil {
		return nil, fmt.Errorf("reward 24h: %s: %w", address, err)
	}
	return tview.NewTableCell(fmt.Sprintf("%f", rewardResponse.Data.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight), nil
}

func getReward(adress string, max time.Time, min time.Time) (*RewardResponse, error) {
	url := fmt.Sprintf("https://api.helium.io/v1/hotspots/%s/rewards/sum?max_time=%s&min_time=%s", adress, max.Format("2006-01-02T15:04:05-0700"), min.Format("2006-01-02T15:04:05-0700"))
	var response *RewardResponse

	if err := getJson(url, &response); err != nil {
		return nil, fmt.Errorf("get json: %w", err)
	}

	return response, nil
}

func getHotspot(address string) (*HotspotResponse, error) {
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

type HotspotResponse struct {
	Data struct {
		Lng            float64   `json:"lng"`
		Lat            float64   `json:"lat"`
		TimestampAdded time.Time `json:"timestamp_added"`
		Status         struct {
			Online      string   `json:"online"`
			ListenAddrs []string `json:"listen_addrs"`
			Height      int      `json:"height"`
		} `json:"status"`
		RewardScale      float64 `json:"reward_scale"`
		Owner            string  `json:"owner"`
		Nonce            int     `json:"nonce"`
		Name             string  `json:"name"`
		Location         string  `json:"location"`
		LastPocChallenge int     `json:"last_poc_challenge"`
		LastChangeBlock  int     `json:"last_change_block"`
		Geocode          struct {
			ShortStreet  string `json:"short_street"`
			ShortState   string `json:"short_state"`
			ShortCountry string `json:"short_country"`
			ShortCity    string `json:"short_city"`
			LongStreet   string `json:"long_street"`
			LongState    string `json:"long_state"`
			LongCountry  string `json:"long_country"`
			LongCity     string `json:"long_city"`
			CityID       string `json:"city_id"`
		} `json:"geocode"`
		BlockAdded int    `json:"block_added"`
		Block      int    `json:"block"`
		Address    string `json:"address"`
	} `json:"data"`
}

type RewardResponse struct {
	Meta struct {
		MinTime time.Time `json:"min_time"`
		MaxTime time.Time `json:"max_time"`
	} `json:"meta"`
	Data struct {
		Total  float64 `json:"total"`
		Sum    int     `json:"sum"`
		Stddev float64 `json:"stddev"`
		Min    float64 `json:"min"`
		Median float64 `json:"median"`
		Max    float64 `json:"max"`
		Avg    float64 `json:"avg"`
	} `json:"data"`
}
