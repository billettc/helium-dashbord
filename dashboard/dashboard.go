package dashboard

import (
	"fmt"
	"time"

	"github.com/billettc/helium-dashbord/helium"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Dashboard struct {
	app *tview.Application
}

func NewDashboard(addresses []string) *Dashboard {

	app := tview.NewApplication()

	table := tview.NewTable()
	table.SetBorders(false)

	table.SetCell(0, 0, tview.NewTableCell("Hotspot Name").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))
	table.SetCell(0, 1, tview.NewTableCell("last 24h").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, 2, tview.NewTableCell("last 7 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, 3, tview.NewTableCell("last 30 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, 4, tview.NewTableCell("Address").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))
	table.SetCell(0, 5, tview.NewTableCell("Owner").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))

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
			time.Sleep(60 * time.Second)
		}
	}()

	app.SetRoot(table, true)

	return &Dashboard{
		app: app,
	}
}

func (d *Dashboard) Run() error {
	return d.app.Run()
}

func refresh(addresses []string, table *tview.Table) error {

	for i, address := range addresses {
		row := i + 1
		hotspotResponse, err := helium.GetHotspot(address)
		if err != nil {
			return fmt.Errorf("getting hotspot: %s: %w", address, err)
		}

		h := hotspotResponse.Hotspot

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
	rewardResponse, err := helium.GetReward(address, max, min)
	if err != nil {
		return nil, fmt.Errorf("reward 24h: %s: %w", address, err)
	}
	return tview.NewTableCell(fmt.Sprintf("%f", rewardResponse.Data.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight), nil
}
