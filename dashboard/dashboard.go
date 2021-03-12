package dashboard

import (
	"context"
	"fmt"
	"sync"

	"github.com/billettc/helium-dashbord/helium"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var logo = ",--.  ,--.,------.   \n|  '--'  ||  .-.  \\  \n|  .--.  ||  |  \\  : \n|  |  |  ||  '--'  / \n`--'  `--'`-------'  "

type Dashboard struct {
	app *tview.Application

	addresses []string
	table     *tview.Table
	lock      sync.Mutex

	rows     map[string]int
	hotspots map[string]*helium.Hotspot
	rewards  map[string]*helium.Rewards
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
	flex := tview.NewFlex()
	pages := tview.NewPages().
		AddPage("main", flex, true, true)

	// Create the layout.
	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSelectable(true, false)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'i' {
			box := tview.NewBox().
				SetBorder(true).
				SetTitle("Centered Box")

			box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					pages.RemovePage("modal")
				}
				return event
			})

			app.SetFocus(box)
			pages.AddPage("modal", modal(box, 40, 10), true, true)
		}
		return event
	})

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

	header := tview.NewFlex()
	header.AddItem(buildMenu(app), 0, 4, false)
	header.AddItem(tview.NewTextView().SetText(logo).SetTextAlign(tview.AlignRight), 0, 1, false)

	table.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	footer := tview.NewFlex().SetBorder(false)
	flex.AddItem(header, 0, 1, false)
	flex.AddItem(table, 0, 4, false).SetBorder(true)
	flex.AddItem(footer, 0, 1, false)

	flex.SetDirection(tview.FlexRow)
	flex.SetBorder(false)

	app.SetRoot(pages, true).SetFocus(table)
	return &Dashboard{
		app: app,

		addresses: addresses,
		table:     table,
		rewards:   map[string]*helium.Rewards{},
		hotspots:  map[string]*helium.Hotspot{},
		rows:      map[string]int{},
	}
}

func modal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, false).
			AddItem(nil, 0, 1, false), width, 1, false).
		AddItem(nil, 0, 1, false)
}

func buildMenu(app *tview.Application) *tview.List {
	return tview.NewList().
		AddItem("List item 1", "Some explanatory text", 'a', nil).
		AddItem("List item 2", "Some explanatory text", 'b', nil).
		AddItem("Quit", "Press to exit", 'q', func() {
			app.Stop()
		})
}

func (d *Dashboard) Run() error {
	err := d.loadData(context.TODO())
	if err != nil {
		return err
	}
	return d.app.Run()
}

func (d *Dashboard) hotspotChange(address string) {
	hotspot := d.hotspots[address]
	rewards := d.rewards[address]
	row := d.rows[address]
	d.app.QueueUpdateDraw(func() {
		d.table.SetCell(row, columnHotpotName, tview.NewTableCell(hotspot.Name).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		d.table.SetCell(row, columnHotspotAddress, tview.NewTableCell(address).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))
		d.table.SetCell(row, columnHotspotOwner, tview.NewTableCell(hotspot.Owner).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignLeft))

		cell := tview.NewTableCell(fmt.Sprintf("%f", rewards.Day1.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
		d.table.SetCell(row, columnLast24h, cell)

		cell = tview.NewTableCell(fmt.Sprintf("%f", rewards.Day7.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
		d.table.SetCell(row, columnLast7d, cell)

		cell = tview.NewTableCell(fmt.Sprintf("%f", rewards.Day30.Total)).SetTextColor(tcell.ColorWhite).SetAlign(tview.AlignRight)
		d.table.SetCell(row, columnlast30d, cell)
	})
}

func (d *Dashboard) loadData(ctx context.Context) error {
	for i, address := range d.addresses {
		d.rows[address] = i + 1
		d.hotspots[address] = &helium.Hotspot{}
		d.rewards[address] = &helium.Rewards{
			Day1:  &helium.Reward{},
			Day7:  &helium.Reward{},
			Day30: &helium.Reward{},
		}

		go func(address string) {
			helium.GetHotspot(ctx, address, func(h *helium.Hotspot, err error) {
				if err != nil {
					panic(fmt.Errorf("get hotspot: %w", err))
				}
				d.hotspots[address] = h
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -1, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 24h: %s: %w", address, err))
				}
				d.rewards[address].Day1 = reward
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -7, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 7d: %s: %w", address, err))
				}
				d.rewards[address].Day7 = reward
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -30, func(reward *helium.Reward, err error) {
				if err != nil {
					panic(fmt.Errorf("reward 30d: %s: %w", address, err))
				}
				d.rewards[address].Day30 = reward
				d.hotspotChange(address)
			})
		}(address)
	}
	return nil
}
