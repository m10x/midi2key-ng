package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

var header = []string{"key", "hotkey", "description", "velocity"}
var data = [][]string{header}

var strNoDevice = "No Device Found"
var strStartListen = "Start Listen"
var strStopListen = "Stop Listen"
var comboSelect *widget.Select
var comboHotkey *widget.Select
var btnListen *widget.Button
var btnNote *widget.Button
var btnAddRow *widget.Button
var btnDeleteRow *widget.Button
var btnEditRow *widget.Button
var selectedCell widget.TableCellID
var popupHotkey *widget.PopUp
var table *widget.Table
var menuItemListen *fyne.MenuItem
var btnRefresh *widget.Button
var menuTray *fyne.Menu

func refreshDevices() {
	devices := getInputPorts()
	if len(devices) == 0 {
		devices = append(devices, strNoDevice)
	}
	comboSelect.Options = devices
	comboSelect.SetSelectedIndex(0)
}

func getMapHotkeys() map[uint8]string {
	m := make(map[uint8]string)

	for i := 1; i < len(data); i++ {
		key, err := strconv.Atoi(data[i][0])
		if err != nil {
			fmt.Printf("ERROR getMapHotkeys: %s\n", err)
		}
		m[uint8(key)] = data[i][1]
	}

	return m
}

func getMapVelocity() map[uint8]uint8 {
	m := make(map[uint8]uint8)

	for i := 1; i < len(data); i++ {
		key, err := strconv.Atoi(data[i][0])
		if err != nil {
			fmt.Printf("ERROR getMapVelocity: %s\n", err)
		}
		velocity, err := strconv.Atoi(data[i][3])
		if err != nil {
			fmt.Printf("ERROR getMapVelocity: %s\n", err)
		}
		m[uint8(key)] = uint8(velocity)
	}

	return m
}

func listen() {
	if btnListen.Text == strStartListen {
		startListen(comboSelect.Selected, getMapHotkeys(), getMapVelocity())
		btnListen.Text = strStopListen
		btnListen.Refresh()
		menuItemListen.Label = strStopListen
		menuTray.Refresh()
		btnRefresh.Disable()
		comboSelect.Disable()
		btnAddRow.Disable()
		btnDeleteRow.Disable()
		btnEditRow.Disable()

	} else {
		stopListen()
		btnListen.Text = strStartListen
		btnListen.Refresh()
		menuItemListen.Label = strStartListen
		menuTray.Refresh()
		btnRefresh.Enable()
		comboSelect.Enable()
		btnAddRow.Enable()
		btnDeleteRow.Enable()
		btnEditRow.Enable()
	}
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

	btnRefresh = widget.NewButton("Refresh Devices", func() {
		refreshDevices()
	})

	btnListen = widget.NewButton(strStartListen, listen)

	hBoxSelect := container.NewHBox(btnRefresh, btnListen)

	table = widget.NewTable(
		func() (int, int) {
			return len(data), len(data[0])
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("tmp")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
		})

	table.OnSelected = func(id widget.TableCellID) {
		selectedCell = id
		fmt.Println("Selected Cell Col", selectedCell.Col, "Row", selectedCell.Row)
	}

	table.SetColumnWidth(0, 39)
	table.SetColumnWidth(1, 240)
	table.SetColumnWidth(2, 240)
	table.SetColumnWidth(3, 70)

	btnAddRow = widget.NewButton("Add Row", func() {
		data = append(data, []string{"", "", "", "255"})
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

		rowToEdit := selectedCell.Row
		popupEdit := widget.NewModalPopUp(nil, w.Canvas())

		lblNote := widget.NewLabel("Note:")
		btnNote = widget.NewButton("Press this Button", func() {
			btnNote.Text = "Listening for Input..."
			btnNote.Disable()
			btnNote.Text = getOneInput(comboSelect.Selected)
			btnNote.Enable()
		})
		lblDescription := widget.NewLabel("Description:")
		entryDescription := widget.NewEntry()
		lblHotkey := widget.NewLabel("Hotkey:")
		entryHotkey := widget.NewEntry()
		comboHotkey = widget.NewSelect([]string{"Command Line Command", "Keypress (combo)", "Write String", "Audio Control"}, func(value string) {
			// TODO: Add Popup to speficy the wanted hotkey. Eg. if Audio Control: Popup to choose if Mute, Volume Up, Volume Down
			if comboHotkey.SelectedIndex() == 0 {
				entryHotkey.Text = "Replace with Command"
				entryHotkey.Refresh()
			} else if comboHotkey.SelectedIndex() == 1 {
				entryHotkey.Text = "Keypress: a,ctrl,alt,cmd"
				entryHotkey.Refresh()
			} else if comboHotkey.SelectedIndex() == 2 {
				entryHotkey.Text = "Write: example"
				entryHotkey.Refresh()
			} else if comboHotkey.SelectedIndex() == 3 {
				comboSound := widget.NewSelect([]string{"Audio: (Un)Mute", "Audio: +10%", "Audio: -10%"}, nil)
				btnSaveHotkey := widget.NewButton("Save", func() {
					entryHotkey.Text = comboSound.Selected
					entryHotkey.Refresh()
					popupHotkey.Hide()
				})
				btnCancelHotkey := widget.NewButton("Cancel", func() {
					popupHotkey.Hide()
				})
				popupHotkey = widget.NewModalPopUp(container.NewVBox(comboSound, container.NewHBox(btnSaveHotkey, btnCancelHotkey)), popupEdit.Canvas)
				popupHotkey.Show()
			}
		})
		lblVelocity := widget.NewLabel("Velocity:")
		entryVelocity := widget.NewEntry()

		btnSave := widget.NewButton("Save", func() {
			data[rowToEdit][0] = btnNote.Text
			data[rowToEdit][1] = entryHotkey.Text
			data[rowToEdit][2] = entryDescription.Text
			data[rowToEdit][3] = entryVelocity.Text
			table.Refresh()
			popupEdit.Hide()
		})
		btnCancel := widget.NewButton("Cancel", func() {
			popupEdit.Hide()
		})

		btnNote.Text = data[rowToEdit][0]
		entryHotkey.Text = data[rowToEdit][1]
		entryDescription.Text = data[rowToEdit][2]
		entryVelocity.Text = data[rowToEdit][3]

		popupEdit.Content = container.NewVBox(container.New(layout.NewFormLayout(), lblNote, btnNote, lblHotkey, container.NewVBox(comboHotkey, entryHotkey), lblDescription, entryDescription, lblVelocity, entryVelocity), container.NewCenter(container.NewHBox(btnSave, btnCancel)))
		popupEdit.Resize(fyne.NewSize(400, 200))
		popupEdit.Show()
	})

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow)

	w.SetContent(container.NewBorder(
		container.NewBorder(nil, nil, hello, hBoxSelect, comboSelect), hBoxTable, nil, nil,
		table,
	))

	refreshDevices()

	if desk, ok := a.(desktop.App); ok {
		menuItemListen = fyne.NewMenuItem(strStartListen, listen)
		menuTray = fyne.NewMenu("midi2key-ng",
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}), menuItemListen)
		desk.SetSystemTrayMenu(menuTray)
	}

	w.ShowAndRun()
}
