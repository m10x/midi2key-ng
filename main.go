package main

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var data = [][]string{[]string{"note", "hotkey", "description", "new2"},
	[]string{"A1", "top right", "new", "new2"},
	[]string{"B2", "bottom right", "newx", "newx2"},
	[]string{"C3", "bottom right", "newx", "newx2"},
	[]string{"D4", "bottom right", "newx", "newx2"}}

var strNoDevice = "No Device Found"
var strStartListen = "Start Listen"
var strStopListen = "Stop Listen"
var midiDevice = strNoDevice
var combo *widget.Select
var btnListen *widget.Button

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
			btnRefresh.Disable()
			combo.Disable()
		} else {
			stopListen()
			btnListen.Text = strStartListen
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

	w.SetContent(container.NewVBox(
		hBoxSelect,
		table,
	))

	refreshDevices()

	w.ShowAndRun()
}
