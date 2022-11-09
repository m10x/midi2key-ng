package pkgGui

import (
	"log"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"github.com/m10x/midi2key-ng/pkg/pkgMidi"
)

const (
	COLUMN_KEY         = 0
	COLUMN_PAYLOAD     = 1
	COLUMN_DESCRIPTION = 2
	COLUMN_VELOCITY    = 3
	COLUMN_TOGGLE      = 4
)

var (
	strNoDevice    = "No Device Found"
	strStartListen = "Start Listen"
	strStopListen  = "Stop Listen"

	header  = []string{"key", "payload", "description", "velocity", "toggle"}
	data    = [][]string{header}
	mapKeys map[uint8]pkgMidi.KeyStruct

	comboSelect    *widget.Select
	comboHotkey    *widget.Select
	btnListen      *widget.Button
	btnNote        *widget.Button
	btnAddRow      *widget.Button
	btnDeleteRow   *widget.Button
	btnEditRow     *widget.Button
	selectedCell   widget.TableCellID
	popupHotkey    *widget.PopUp
	table          *widget.Table
	menuItemListen *fyne.MenuItem
	btnRefresh     *widget.Button
	menuTray       *fyne.Menu
	desk           desktop.App
	a              fyne.App
)

func Startup(versionTool string, versionPref int) {
	a = app.NewWithID("de.m10x.midi2key-ng")
	w := a.NewWindow("midi2key-ng " + versionTool)
	w.Resize(fyne.NewSize(1000, 400))

	mapKeys = make(map[uint8]pkgMidi.KeyStruct)

	hello := widget.NewLabel("Hello! :)")

	comboSelect = widget.NewSelect([]string{""}, func(value string) {
		log.Println("Selected midi device " + value)
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
		log.Println("Selected Cell Col", selectedCell.Col, "Row", selectedCell.Row)
	}

	table.SetColumnWidth(0, 39)
	table.SetColumnWidth(1, 480)
	table.SetColumnWidth(2, 335)
	table.SetColumnWidth(3, 70)
	table.SetColumnWidth(4, 60)

	btnAddRow = widget.NewButton("Add Row", func() {
		data = append(data, []string{"-", "-", "-", "255", "false"})
		table.Refresh()
		setPreferences(versionPref)
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
		setPreferences(versionPref)
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
			btnNote.Text = pkgMidi.GetOneInput(comboSelect.Selected)
			btnNote.Enable()
		})
		lblDescription := widget.NewLabel("Description:")
		entryDescription := widget.NewEntry()
		lblHotkey := widget.NewLabel("Hotkey type:")
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
				devices := []string{}
				for _, sink := range pkgCmd.GetSinks() {
					devices = append(devices, sink.Description)
				}
				for _, source := range pkgCmd.GetSources() {
					devices = append(devices, source.Description)
				}
				for _, inputSink := range pkgCmd.GetSinkInputs() {
					devices = append(devices, inputSink.Description)
				}
				comboDevice := widget.NewSelect(devices, nil)
				comboSound := widget.NewSelect([]string{"(Un)Mute", "Volume +10%", "Volume -10%", "Volume =50%"}, nil)
				btnSaveHotkey := widget.NewButton("Save", func() {
					entryHotkey.Text = "Audio: " + comboDevice.Selected + ": " + comboSound.Selected
					entryHotkey.Refresh()
					popupHotkey.Hide()
				})
				btnCancelHotkey := widget.NewButton("Cancel", func() {
					popupHotkey.Hide()
				})
				popupHotkey = widget.NewModalPopUp(container.NewVBox(comboDevice, comboSound, container.NewHBox(btnSaveHotkey, btnCancelHotkey)), popupEdit.Canvas)
				popupHotkey.Show()
			}
		})
		lblVelocity := widget.NewLabel("Velocity:")
		entryVelocity := widget.NewEntry()
		lblToggle := widget.NewLabel("Toggle:")
		checkToggle := widget.NewCheck("Toggle LED", nil)

		btnSave := widget.NewButton("Save", func() {

			if comboHotkey.SelectedIndex() > -1 { // only add hotkey if a hotkeytype was selected
				data[rowToEdit][0] = btnNote.Text
				data[rowToEdit][1] = entryHotkey.Text
				data[rowToEdit][2] = entryDescription.Text
				data[rowToEdit][3] = entryVelocity.Text
				if checkToggle.Checked {
					data[rowToEdit][4] = "true"
				} else {
					data[rowToEdit][4] = "false"
				}
				table.Refresh()
				setPreferences(versionPref)
			}
			popupEdit.Hide()
		})
		btnCancel := widget.NewButton("Cancel", func() {
			popupEdit.Hide()
		})

		btnNote.Text = data[rowToEdit][0]
		entryHotkey.Text = data[rowToEdit][1]
		entryDescription.Text = data[rowToEdit][2]
		entryVelocity.Text = data[rowToEdit][3]
		if data[rowToEdit][4] == "true" {
			checkToggle.Checked = true
		}

		popupEdit.Content = container.NewVBox(container.New(layout.NewFormLayout(), lblNote, btnNote, lblHotkey, container.NewVBox(comboHotkey, entryHotkey), lblDescription, entryDescription, lblVelocity, entryVelocity, lblToggle, checkToggle), container.NewCenter(container.NewHBox(btnSave, btnCancel)))
		popupEdit.Resize(fyne.NewSize(400, 200))
		popupEdit.Show()
	})

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow)

	w.SetContent(container.NewBorder(
		container.NewBorder(nil, nil, hello, hBoxSelect, comboSelect), hBoxTable, nil, nil,
		table,
	))

	refreshDevices()

	var ok bool
	if desk, ok = a.(desktop.App); ok {
		menuItemListen = fyne.NewMenuItem(strStartListen, listen)
		menuTray = fyne.NewMenu("midi2key-ng "+versionTool,
			fyne.NewMenuItem("Show", func() {
				w.Show()
			}), menuItemListen)
		desk.SetSystemTrayMenu(menuTray)
		desk.SetSystemTrayIcon(resourceMidiOffPng)
	}

	getPreferences(versionPref)

	w.ShowAndRun()
}

func refreshDevices() {
	devices := pkgMidi.GetInputPorts()
	if len(devices) == 0 {
		devices = append(devices, strNoDevice)
	}
	comboSelect.Options = devices
	comboSelect.SetSelectedIndex(0)
}

func fillMapKeys() {
	for i := 1; i < len(data); i++ {
		var midiType int
		switch data[i][COLUMN_KEY][:1] {
		case "B":
			midiType = pkgMidi.MIDI_BUTTON
		case "K":
			midiType = pkgMidi.MIDI_KNOB
		case "S":
			midiType = pkgMidi.MIDI_SLIDER
		}

		keyId, err := strconv.Atoi(data[i][COLUMN_KEY][1:]) // first char is midiType
		if err != nil {
			log.Printf("ERROR fillMapKeys: strconv.Atoi 1: %s\n", err)
			continue
		}

		vel, err := strconv.Atoi(data[i][0][1:]) // first char is midiType
		if err != nil {
			log.Printf("ERROR fillMapKeys: strconv.Atoi 2:%s\n", err)
			continue
		}

		mapKeys[uint8(keyId)] = pkgMidi.KeyStruct{
			MidiType: midiType,
			Key:      data[i][COLUMN_KEY],
			Payload:  data[i][COLUMN_PAYLOAD],
			Velocity: uint8(vel),
			Toggle:   data[i][COLUMN_TOGGLE] == "true",
		}
	}
}

func listen() {
	if btnListen.Text == strStartListen {
		fillMapKeys()
		pkgMidi.StartListen(table, data, comboSelect.Selected, mapKeys)
		btnListen.Text = strStopListen
		btnListen.Refresh()
		menuItemListen.Label = strStopListen
		menuTray.Refresh()
		desk.SetSystemTrayIcon(resourceMidiOnPng)
		btnRefresh.Disable()
		comboSelect.Disable()
		btnAddRow.Disable()
		btnDeleteRow.Disable()
		btnEditRow.Disable()
	} else {
		pkgMidi.StopListen()
		btnListen.Text = strStartListen
		btnListen.Refresh()
		menuItemListen.Label = strStartListen
		menuTray.Refresh()
		desk.SetSystemTrayIcon(resourceMidiOffPng)
		btnRefresh.Enable()
		comboSelect.Enable()
		btnAddRow.Enable()
		btnDeleteRow.Enable()
		btnEditRow.Enable()
	}
}
