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
	footer    *tview.Flex
	lock      sync.Mutex

	rows         map[string]int
	addressAtRow map[int]string
	hotspots     map[string]*helium.Hotspot
	rewards      map[string]*helium.Rewards
	pages        *tview.Pages
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
	dashboard.pages = tview.NewPages().
		AddPage("main", flex, true, true)

	// Create the layout.
	dashboard.table = buildTable(dashboard)

	header := tview.NewFlex()
	header.AddItem(buildMenu(dashboard.app), 0, 4, false)
	header.AddItem(tview.NewTextView().SetText(logo).SetTextAlign(tview.AlignRight), 0, 1, false)

	dashboard.table.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	dashboard.footer = tview.NewFlex()
	dashboard.footer.SetBorder(false)
	flex.AddItem(header, 0, 1, false)
	flex.AddItem(dashboard.table, 0, 4, false).SetBorder(true)
	flex.AddItem(dashboard.footer, 0, 1, false)

	flex.SetDirection(tview.FlexRow)
	flex.SetBorder(false)

	dashboard.app.SetRoot(dashboard.pages, true).SetFocus(dashboard.table)

	return dashboard
}

func buildTable(dashboard *Dashboard) *tview.Table {
	table := tview.NewTable()
	table.SetBorders(false)
	table.SetSelectable(true, false)

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

	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()
		if row < 1 {
			return event
		}

		address := dashboard.addressAtRow[row]
		hotspot := dashboard.hotspots[address]
		if event.Rune() == 'i' {
			dashboard.hotspotDetail(hotspot)
		}
		return event
	})
	return table
}

func (d *Dashboard) hotspotDetail(hotspot *helium.Hotspot) {
	detail := tview.NewFlex()
	detail.SetDirection(tview.FlexRow)

	//detail := tview.NewBox()

	detail.SetBorder(true).SetTitle(" " + hotspot.Name + " ").SetTitleColor(tcell.ColorYellow)
	detail.SetBorderPadding(1, 1, 1, 1)
	detail.SetBackgroundColor(tcell.ColorBlack)

	hack := tview.NewBox()

	hack.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			d.pages.RemovePage("modal")
			d.app.SetFocus(d.table)
		}
		return event
	})

	ownership := tview.NewFlex().
		AddItem(labelValue("Address", hotspot.Address, 10), 0, 1, false).
		AddItem(labelValue("Owner", hotspot.Owner, 10), 0, 1, false).
		SetDirection(tview.FlexColumn)

	detail.AddItem(ownership, 0, 1, false)

	//location := tview.NewFlex()
	//location.SetBorder(true).SetBorderPadding(1, 1, 1, 1)
	//location.
	//	AddItem(labelValue("lat", fmt.Sprintf("%f", hotspot.Lat), 5), 0, 1, false).
	//	AddItem(labelValue("long", fmt.Sprintf("%f", hotspot.Lng), 5), 0, 1, false)

	detail.AddItem(locationBox(hotspot), 15, 0, false)

	detail.AddItem(hack, 0, 1, false)

	d.app.SetFocus(hack)

	d.pages.AddPage("modal", detail, true, true)

}

func locationBox(hotspot *helium.Hotspot) tview.Primitive {
	return tview.NewBox().
		SetBorder(true).
		SetDrawFunc(func(screen tcell.Screen, x int, y int, width int, height int) (int, int, int, int) {
			// Draw a horizontal line across the middle of the box.
			//centerY := y + height/2
			//for cx := x + 1; cx < x+width-1; cx++ {
			//	screen.SetContent(cx, centerY, tview.BoxDrawingsLightHorizontal, nil, tcell.StyleDefault.Foreground(tcell.ColorWhite))
			//}

			// Write some text along the horizontal line.
			tview.Print(screen, "Lat:", x+1, y+1, width-2, tview.AlignLeft, tcell.ColorYellow)
			tview.Print(screen, fmt.Sprintf("%f", hotspot.Lat), x+1+5, y+1, width-2, tview.AlignLeft, tcell.ColorWhite)

			tview.Print(screen, "Log:", x+1, y+2, width-2, tview.AlignLeft, tcell.ColorYellow)
			tview.Print(screen, fmt.Sprintf("%f", hotspot.Lng), x+1+5, y+2, width-2, tview.AlignLeft, tcell.ColorWhite)

			// Space for other content.
			return x + 1, y + 1, 15, 5
		})
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

func labelValue(label, value string, labelSize int) tview.Primitive {
	flex := tview.NewFlex().
		AddItem(tview.NewTextView().SetText(label+":").SetTextColor(tcell.ColorYellow), labelSize, 0, false).
		AddItem(tview.NewTextView().SetText(value).SetTextAlign(tview.AlignLeft), 0, 1, false)

	flex.SetDirection(tview.FlexColumn)

	return flex
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
	d.footer.AddItem(tview.NewTextView().SetText(err.Error()).SetTextAlign(tview.AlignCenter), 0, 1, false)
}

func (d *Dashboard) loadData(ctx context.Context) error {
	for _, address := range d.addresses {
		ctx, cancel := context.WithCancel(ctx)

		d.rewards[address] = &helium.Rewards{
			Day1:  &helium.Reward{},
			Day7:  &helium.Reward{},
			Day30: &helium.Reward{},
		}

		go func(address string) {
			helium.GetHotspot(ctx, address, func(h *helium.Hotspot, err error) {
				if err != nil {
					d.displayError(err)
					cancel()
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
