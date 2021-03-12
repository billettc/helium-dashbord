package dashboard

import (
	"context"
	"fmt"
	"time"

	"github.com/billettc/helium-dashbord/helium"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Dashboard struct {
	app *tview.Application
}

const (
	columnHotpotName = iota
	columnLast24h
	columnLast7d
	columnlast30d
	columnHotspotAddress
	columnHotspotOwner
)

func NewDashboard(addresses []string) *Dashboard {
	app := tview.NewApplication()
	// Create the layout.

	table := tview.NewTable()
	table.SetBorders(false)

	table.SetCell(0, columnHotpotName, tview.NewTableCell("Hotspot Name").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))
	table.SetCell(0, columnLast24h, tview.NewTableCell("last 24h").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, columnLast7d, tview.NewTableCell("last 7 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, columnlast30d, tview.NewTableCell("last 30 days").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignRight).SetExpansion(20))
	table.SetCell(0, columnHotspotAddress, tview.NewTableCell("Address").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))
	table.SetCell(0, columnHotspotOwner, tview.NewTableCell("Owner").SetTextColor(tcell.ColorYellow).SetAlign(tview.AlignLeft))

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
				if err := loadData(context.TODO(), addresses, table); err != nil {
					panic(err)
				}
			})
			time.Sleep(60 * time.Second)
		}
	}()

	header := tview.NewBox().SetTitle("Header").SetBorder(true)

	flex := tview.NewFlex()
	flex.AddItem(header, 0, 1, false)
	flex.AddItem(table, 0, 3, false).SetBorder(true)
	flex.AddItem(tview.NewBox().SetTitle("Footer").SetBorder(true), 0, 1, false)

	flex.SetDirection(tview.FlexRow)

	app.SetRoot(flex, true)
	return &Dashboard{
		app: app,
	}
}

func (d *Dashboard) Run() error {
	return d.app.Run()
}

func loadData(ctx context.Context, addresses []string, table *tview.Table) error {
	for i, address := range addresses {
		row := i + 1

		go func(row int, address string) {
			helium.GetHotspot(ctx, address, func(h *helium.Hotspot, err error) {
				table.SetCell(row, columnHotpotName, tview.NewTableCell(h.Name).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
				table.SetCell(row, columnHotspotAddress, tview.NewTableCell(h.Address).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
				table.SetCell(row, columnHotspotOwner, tview.NewTableCell(h.Owner).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
			})
		}(row, address)

		go func(row int, address string) {
			helium.GetReward(ctx, address, -1, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 24h: %s: %w", address, err))
				}
				cell := tview.NewTableCell(fmt.Sprintf("%f", reward.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
				table.SetCell(row, columnLast24h, cell)
			})
		}(row, address)

		go func(row int, address string) {
			helium.GetReward(ctx, address, -7, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 7d: %s: %w", address, err))
				}
				cell := tview.NewTableCell(fmt.Sprintf("%f", reward.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
				table.SetCell(row, columnLast7d, cell)
			})
		}(row, address)

		go func(row int, address string) {
			helium.GetReward(ctx, address, -30, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 30d: %s: %w", address, err))
				}
				cell := tview.NewTableCell(fmt.Sprintf("%f", reward.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
				table.SetCell(row, columnlast30d, cell)
			})
		}(row, address)
	}

	return nil
}
