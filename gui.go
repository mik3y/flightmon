package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	lru "github.com/hashicorp/golang-lru"
	pb "github.com/mik3y/goadsb/proto"
	"github.com/rivo/tview"
)

var updatedTextColor = tcell.ColorWhite
var normalTextColor = tcell.ColorGray

func getHeaderText(cache *lru.Cache) string {
	return fmt.Sprintf("Flightmon v%s · %d tracked", Version, cache.Len())
}

func updateRow(table *tview.Table, row int, positionData *pb.PositionUpdate) {
	var cell *tview.TableCell
	colNum := 0

	cell, colNum = table.GetCell(row, colNum), colNum+1
	cell.SetText(*positionData.IcaoId)

	now := int64(time.Now().UnixNano() / int64(time.Millisecond))
	ageSeconds := (now - *positionData.Timestamp) / 1000
	cell, colNum = table.GetCell(row, colNum), colNum+1
	cell.SetText(fmt.Sprint(ageSeconds))

	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Callsign != nil {
		if cell.Text != *positionData.Callsign {
			cell.SetText(*positionData.Callsign).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Lat != nil {
		s := fmt.Sprint(*positionData.Lat)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Lng != nil {
		s := fmt.Sprint(*positionData.Lng)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Altitude != nil {
		s := fmt.Sprint(*positionData.Altitude)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.GroundSpeed != nil {
		s := fmt.Sprint(*positionData.GroundSpeed)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Track != nil {
		s := fmt.Sprint(*positionData.Track)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.VerticalRate != nil {
		s := fmt.Sprint(*positionData.VerticalRate)
		if cell.Text != s {
			cell.SetText(s).SetTextColor(updatedTextColor)
		}
	}
	cell, colNum = table.GetCell(row, colNum), colNum+1
	if positionData.Squawk != nil {
		if cell.Text != *positionData.Squawk {
			cell.SetText(*positionData.Squawk).SetTextColor(updatedTextColor)
		}
	}
}

func clearUpdatedColors(table *tview.Table, row int) {
	for c := 0; c < table.GetColumnCount(); c++ {
		table.GetCell(row, c).SetTextColor(normalTextColor)
	}
}

func updateTable(cache *lru.Cache, table *tview.Table, header *tview.TextView) {
	// Clear any previously-updated cells.
	for r := 1; r < table.GetRowCount(); r++ {
		clearUpdatedColors(table, r)
	}

	// Pass 1: Refresh all existing rows.
	updatedMap := make(map[string]bool)
	for r := table.GetRowCount(); r >= 1; r-- {
		icaoID := table.GetCell(r, 0).Text
		updatedMap[icaoID] = true
		positionDataPtr, ok := cache.Peek(icaoID)
		if !ok {
			table.RemoveRow(r)
		} else {
			positionData := positionDataPtr.(*pb.PositionUpdate)
			updateRow(table, r, positionData)
		}
	}

	// Pass 2: Add any new rows.
	keyItems := cache.Keys()
	keys := make([]string, len(keyItems))
	for i, v := range keyItems {
		keys[i] = fmt.Sprint(v)
	}

	for idx := range keys {
		icaoID := keys[idx]
		if updatedMap[icaoID] {
			continue
		}
		positionDataPtr, ok := cache.Peek(icaoID)
		if !ok {
			continue
		}
		positionData := positionDataPtr.(*pb.PositionUpdate)
		rowNum := table.GetRowCount()
		for c := 0; c < table.GetColumnCount(); c++ {
			table.SetCell(rowNum, c, tview.NewTableCell("").SetTextColor(updatedTextColor))
		}
		updateRow(table, rowNum, positionData)
	}

	header.SetText(getHeaderText(cache))
}

// ShowUI shows a curses-like GUI in the terminal.
func ShowUI(cache *lru.Cache, donec chan<- struct{}) {
	app := tview.NewApplication()

	header := tview.NewTextView().
		SetText(getHeaderText(cache)).
		SetTextColor(tcell.ColorGreen)

	table := tview.NewTable().
		SetBorders(true)

	grid := tview.NewGrid().
		SetRows(1, 0).
		SetBorders(false).
		AddItem(header, 0, 0, 1, 1, 0, 0, false).
		AddItem(table, 1, 0, 1, 1, 0, 0, true)

	headers := strings.Split("ICAO Age Callsign Lat Lng Altitude GroundSpeed Track VerticalRate Squawk", " ")
	for h := 0; h < len(headers); h++ {
		table.SetCell(0, h,
			tview.NewTableCell(headers[h]).
				SetTextColor(tcell.ColorYellow).
				SetAlign(tview.AlignCenter))
	}

	table.SetFixed(1, 4)

	go func() {
		for {
			time.Sleep(time.Second)
			updateTable(cache, table, header)
			app.Draw()
		}
	}()

	if err := app.SetRoot(grid, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
	donec <- struct{}{}
}
