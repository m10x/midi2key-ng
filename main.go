package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var header = []string{"note", "hotkey", "description", "new2"}
var data = [][]string{header}

var strNoDevice = "No Device Found"
var strStartListen = "Start Listen"
var strStopListen = "Stop Listen"
var midiDevice = strNoDevice
var combo *widget.Select
var btnListen *widget.Button
var selectedCell widget.TableCellID

func refreshDevices() {
	devices := getInputPorts()
	if len(devices) == 0 {
		devices = append(devices, strNoDevice)
	}
	combo.Options = devices
	combo.SetSelectedIndex(0)
}

func main() {
	a := app.NewWithID("de.m10x.midi2key-ng")
	w := a.NewWindow("midi2key-ng")
	w.Resize(fyne.NewSize(600, 400))

	hello := widget.NewLabel("Hello! :)")

	combo = widget.NewSelect([]string{""}, func(value string) {
		midiDevice = value
		fmt.Sprintln("Selected midi device " + midiDevice)
		if combo.Selected != strNoDevice {
			btnListen.Enable()
		} else {
			btnListen.Disable()
		}
	})

	btnRefresh := widget.NewButton("Refresh Devices", func() {
		refreshDevices()
	})

	btnListen = widget.NewButton(strStartListen, func() {
		if btnListen.Text == strStartListen {
			startListen(midiDevice)
			btnListen.Text = strStopListen
			btnListen.Refresh()
			btnRefresh.Disable()
			combo.Disable()

		} else {
			stopListen()
			btnListen.Text = strStartListen
			btnListen.Refresh()
			btnRefresh.Enable()
			combo.Enable()
		}
	})

	hBoxSelect := container.NewHBox(hello, combo, btnRefresh, btnListen)

	table := widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("wide content")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
		})

	table.OnSelected = func(id widget.TableCellID) {
		selectedCell = id
		fmt.Println("Selected Cell Col", selectedCell.Col, "Row", selectedCell.Row)
	}

	btnAddRow := widget.NewButton("Add Row", func() {
		dataLen := strconv.Itoa(len(data))
		data = append(data, []string{dataLen, dataLen, dataLen, dataLen})
		table.Refresh()
	})
	btnEditRow := widget.NewButton("Edit Row", nil)
	btnDeleteRow := widget.NewButton("Delete Row", func() {
		tmpData := [][]string{header}
		for i, x := range data {
			if i != selectedCell.Row && i != 0 { // Dont apped header again, dont append row to delete
				tmpData = append(tmpData, x)
			}
		}
		data = tmpData
		table.Refresh()
	})

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow)

	w.SetContent(container.NewBorder(
		hBoxSelect, hBoxTable, nil, nil,
		table,
	))

	refreshDevices()

	w.ShowAndRun()
}
