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
	footer    *tview.Box
	lock      sync.Mutex

	rows         map[string]int
	addressAtRow map[int]string
	hotspots     map[string]*helium.Hotspot
	rewards      map[string]*helium.Rewards
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
	dashboard := &Dashboard{
		addresses:    addresses,
		addressAtRow: map[int]string{},
		rewards:      map[string]*helium.Rewards{},
		hotspots:     map[string]*helium.Hotspot{},
		rows:         map[string]int{},
	}

	dashboard.app = tview.NewApplication()

	flex := tview.NewFlex()
	pages := tview.NewPages().
		AddPage("main", flex, true, true)

	// Create the layout.
	table := tview.NewTable()
	dashboard.table = table

	table.SetBorders(false)
	table.SetSelectable(true, false)

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if row, _ := table.GetSelection(); row < 1 {
			return event
		}

		//hotspot := dashboard.hotspots[dashboard]
		if event.Rune() == 'i' {
			box := tview.NewBox().
				SetBorder(true).
				SetTitle("BOXO")

			box.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if event.Key() == tcell.KeyEsc {
					pages.RemovePage("modal")
				}
				return event
			})

			dashboard.app.SetFocus(box)
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
			dashboard.app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, true)
		}
	}).SetSelectedFunc(func(row int, column int) {
		table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		table.SetSelectable(false, false)
	})

	header := tview.NewFlex()
	header.AddItem(buildMenu(dashboard.app), 0, 4, false)
	header.AddItem(tview.NewTextView().SetText(logo).SetTextAlign(tview.AlignRight), 0, 1, false)

	table.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	dashboard.footer = tview.NewFlex().SetBorder(false)
	flex.AddItem(header, 0, 1, false)
	flex.AddItem(table, 0, 4, false).SetBorder(true)
	flex.AddItem(dashboard.footer, 0, 1, false)

	flex.SetDirection(tview.FlexRow)
	flex.SetBorder(false)

	dashboard.app.SetRoot(pages, true).SetFocus(table)

	return dashboard
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
	d.lock.Lock()
	defer d.lock.Unlock()

	hotspot, ok := d.hotspots[address]
	if !ok {
		return //
	}

	rewards := d.rewards[address]
	row, ok := d.rows[address]
	if !ok {
		row = d.table.GetRowCount()
		d.rows[address] = row
		d.addressAtRow[row] = address
	}

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

func (d *Dashboard) displayError(err error) {
	// do something
}

func (d *Dashboard) loadData(ctx context.Context) error {
	for _, address := range d.addresses {
		d.rewards[address] = &helium.Rewards{
			Day1:  &helium.Reward{},
			Day7:  &helium.Reward{},
			Day30: &helium.Reward{},
		}

		go func(address string) {
			helium.GetHotspot(ctx, address, func(h *helium.Hotspot, err error) {
				if err != nil {
					d.displayError(err)
					return
				}
				d.hotspots[address] = h
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -1, func(reward *helium.Reward, err error) {
				if err != nil {
					d.displayError(err)
					return
				}
				d.rewards[address].Day1 = reward
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -7, func(reward *helium.Reward, err error) {
				if err != nil {
					d.displayError(err)
					return
				}
				d.rewards[address].Day7 = reward
				d.hotspotChange(address)
			})
		}(address)

		go func(address string) {
			helium.GetReward(ctx, address, -30, func(reward *helium.Reward, err error) {
				if err != nil {
					d.displayError(err)
					return
				}
				d.rewards[address].Day30 = reward
				d.hotspotChange(address)
			})
		}(address)
	}
	return nil
}
