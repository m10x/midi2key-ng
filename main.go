package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var header = []string{"note", "hotkey", "description", "new2"}
var data = [][]string{header}

var strNoDevice = "No Device Found"
var strStartListen = "Start Listen"
var strStopListen = "Stop Listen"
var comboSelect *widget.Select
var comboHotkey *widget.Select
var btnListen *widget.Button
var btnAddRow *widget.Button
var btnDeleteRow *widget.Button
var btnEditRow *widget.Button
var selectedCell widget.TableCellID

func refreshDevices() {
	devices := getInputPorts()
	if len(devices) == 0 {
		devices = append(devices, strNoDevice)
	}
	comboSelect.Options = devices
	comboSelect.SetSelectedIndex(0)
}

func main() {
	a := app.NewWithID("de.m10x.midi2key-ng")
	w := a.NewWindow("midi2key-ng")
	w.Resize(fyne.NewSize(600, 400))

	hello := widget.NewLabel("Hello! :)")

	comboSelect = widget.NewSelect([]string{""}, func(value string) {
		fmt.Sprintln("Selected midi device " + value)
		if comboSelect.Selected != strNoDevice {
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
			startListen(comboSelect.Selected)
			btnListen.Text = strStopListen
			btnListen.Refresh()
			btnRefresh.Disable()
			comboSelect.Disable()
			btnAddRow.Disable()
			btnDeleteRow.Disable()
			btnEditRow.Disable()

		} else {
			stopListen()
			btnListen.Text = strStartListen
			btnListen.Refresh()
			btnRefresh.Enable()
			comboSelect.Enable()
			btnAddRow.Enable()
			btnDeleteRow.Enable()
			btnEditRow.Enable()
		}
	})

	hBoxSelect := container.NewHBox(btnRefresh, btnListen)

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

	btnAddRow = widget.NewButton("Add Row", func() {
		dataLen := strconv.Itoa(len(data))
		data = append(data, []string{dataLen, dataLen, dataLen, dataLen})
		table.Refresh()
	})
	btnDeleteRow = widget.NewButton("Delete Row", func() {
		tmpData := [][]string{header}
		for i, x := range data {
			if i != selectedCell.Row && i != 0 { // Dont apped header again, dont append row to delete
				tmpData = append(tmpData, x)
			}
		}
		data = tmpData
		table.Refresh()
	})
	btnEditRow = widget.NewButton("Edit Row", func() {
		if selectedCell.Row == 0 {
			return
		}
		btnAddRow.Disable()
		btnDeleteRow.Disable()
		btnEditRow.Disable()
		comboSelect.Disable()
		btnRefresh.Disable()
		btnListen.Disable()

		rowToEdit := selectedCell.Row
		popupEdit := a.NewWindow("Edit Row")
		popupEdit.Resize(fyne.NewSize(400, 300))
		popupEdit.SetOnClosed(func() {
			btnAddRow.Enable()
			btnDeleteRow.Enable()
			btnEditRow.Enable()
			btnListen.Enable()
			if btnListen.Text == strStartListen {
				comboSelect.Enable()
				btnRefresh.Enable()
			}
		})

		lblNote := widget.NewLabel("Note:")
		btnNote := widget.NewButton("Press this Button", nil)
		lblDescription := widget.NewLabel("Description:")
		entryDescription := widget.NewEntry()
		lblHotkey := widget.NewLabel("Hotkey:")
		entryHotkey := widget.NewEntry()
		comboHotkey = widget.NewSelect([]string{"Command Line Command", "Keypress (combo)", "Audio Control"}, func(value string) {
			if comboHotkey.SelectedIndex() == -1 {
				entryHotkey.Disable()
			} else {
				entryHotkey.Enable()
				// TODO: Add Popup to speficy the wanted hotkey. Eg. if Audio Control: Popup to choose if Mute, Volume Up, Volume Down
			}
		})
		comboHotkey.SetSelectedIndex(0)

		btnSave := widget.NewButton("Save", func() {
			data[rowToEdit][0] = btnNote.Text
			data[rowToEdit][1] = entryHotkey.Text
			data[rowToEdit][2] = entryDescription.Text
			table.Refresh()
			popupEdit.Close()
		})
		btnCancel := widget.NewButton("Cancel", func() {
			popupEdit.Close()
		})

		btnNote.Text = data[rowToEdit][0]
		entryHotkey.Text = data[rowToEdit][1]
		entryDescription.Text = data[rowToEdit][2]

		popupEdit.SetContent(container.NewVBox(container.New(layout.NewFormLayout(), lblNote, btnNote, lblHotkey, container.NewVBox(comboHotkey, entryHotkey), lblDescription, entryDescription), container.NewCenter(container.NewHBox(btnSave, btnCancel))))
		popupEdit.Show()
	})

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow)

	w.SetContent(container.NewBorder(
		container.NewBorder(nil, nil, hello, hBoxSelect, comboSelect), hBoxTable, nil, nil,
		table,
	))

	refreshDevices()

	w.ShowAndRun()
}
