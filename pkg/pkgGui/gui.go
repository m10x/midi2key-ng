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
	COLUMN_SPECIAL     = 4

	COLUMN_COUNT = 5
)

var (
	strNoDevice    = "No Device Found"
	strStartListen = "Start Listen"
	strStopListen  = "Stop Listen"

	data    = [][]string{}
	mapKeys map[uint8]pkgMidi.KeyStruct

	comboSelect    *widget.Select
	comboPayload   *widget.Select
	btnListen      *widget.Button
	btnNote        *widget.Button
	btnAddRow      *widget.Button
	btnDeleteRow   *widget.Button
	btnEditRow     *widget.Button
	lblOutput      *widget.Label
	checkSpecial   *widget.Check
	entryVelocity  *widget.Entry
	selectedCell   widget.TableCellID
	popupPayload   *widget.PopUp
	table          *widget.Table
	menuItemListen *fyne.MenuItem
	btnRefresh     *widget.Button
	menuTray       *fyne.Menu
	desk           desktop.App
	a              fyne.App
	w              fyne.Window
)

func Startup(versionTool string) {
	a = app.NewWithID("de.m10x.midi2key-ng")
	w = a.NewWindow("midi2key-ng " + versionTool)
	w.Resize(fyne.NewSize(1050, 400))

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

	table = widget.NewTableWithHeaders(
		func() (int, int) {
			return len(data), COLUMN_COUNT
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("tmp")
		},
		func(i widget.TableCellID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(data[i.Row][i.Col])
		})

	table.OnSelected = func(id widget.TableCellID) {
		table.Select(widget.TableCellID{
			Row: id.Row,
			Col: 0,
		})
		log.Println("Selected Cell Col", id.Col, "Row", id.Row)
		selectedCell = id // wichtig für deleteRow und editRow
	}

	table.SetColumnWidth(0, 39)
	table.SetColumnWidth(1, 480)
	table.SetColumnWidth(2, 335)
	table.SetColumnWidth(3, 70)
	table.SetColumnWidth(4, 60)

	table.CreateHeader = headerCreate
	table.UpdateHeader = headerUpdate

	btnAddRow = widget.NewButton("Add Row", func() {
		data = append(data, []string{"-", "-", "-", "0", "false"})
		table.Refresh()
		setPreferences(versionTool)
	})
	btnDeleteRow = widget.NewButton("Delete Row", func() {
		tmpData := [][]string{}
		for i, x := range data {
			if i != selectedCell.Row { // dont append row to delete
				tmpData = append(tmpData, x)
			} else {
				log.Println(i, x)
			}
		}
		data = tmpData
		table.Refresh()
		setPreferences(versionTool)
	})
	btnEditRow = widget.NewButton("Edit Row", func() {
		rowToEdit := selectedCell.Row
		popupEdit := widget.NewModalPopUp(nil, w.Canvas())

		lblNote := widget.NewLabel("Note:")
		strNoButton := "Press this Button"
		strSpecialDisabled := "Waiting for key..."
		btnNote = widget.NewButton(strNoButton, func() {
			btnNote.Text = "Listening for Input..."
			btnNote.Disable()
			btnNote.Text = pkgMidi.GetOneInput(comboSelect.Selected)
			btnNote.Enable()

			configureCheckSpecial(strSpecialDisabled)
			if len(btnNote.Text) < 6 {
				switch btnNote.Text[:1] {
				case pkgMidi.MIDI_BUTTON:
					entryVelocity.Text = "255"
				case pkgMidi.MIDI_KNOB:
					entryVelocity.Text = "33"
				case pkgMidi.MIDI_SLIDER:
					entryVelocity.Text = "16256"
				default:
					log.Printf("ERROR widget.NewButton: No valid Midi Type: %s\n", btnNote.Text)
				}
				entryVelocity.Refresh()
			}
		})
		lblDescription := widget.NewLabel("Description:")
		entryDescription := widget.NewEntry()
		lblPayload := widget.NewLabel("Payload:")
		entryPayload := widget.NewEntry()
		comboPayload = widget.NewSelect([]string{"Command Line Command", "Keypress (combo)", "Write String", "Audio Control"}, func(value string) {
			// TODO: Add Popup to speficy the wanted Payload. Eg. if Audio Control: Popup to choose if Mute, Volume Up, Volume Down
			if comboPayload.SelectedIndex() == 0 {
				entryPayload.Text = "Replace with Command"
				entryPayload.Refresh()
			} else if comboPayload.SelectedIndex() == 1 {
				entryPayload.Text = "Keypress: a,ctrl,alt,cmd"
				entryPayload.Refresh()
			} else if comboPayload.SelectedIndex() == 2 {
				entryPayload.Text = "Write: example"
				entryPayload.Refresh()
			} else if comboPayload.SelectedIndex() == 3 {
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
				btnSavePayload := widget.NewButton("Save", func() {
					entryPayload.Text = "Audio: " + comboDevice.Selected + ": " + comboSound.Selected
					entryPayload.Refresh()
					popupPayload.Hide()
				})
				btnCancelPayload := widget.NewButton("Cancel", func() {
					popupPayload.Hide()
				})
				popupPayload = widget.NewModalPopUp(container.NewVBox(comboDevice, comboSound, container.NewHBox(btnSavePayload, btnCancelPayload)), popupEdit.Canvas)
				popupPayload.Show()
			}
		})
		lblVelocity := widget.NewLabel("Velocity:")
		entryVelocity = widget.NewEntry()
		lblToggle := widget.NewLabel("Special:")
		checkSpecial = widget.NewCheck(strSpecialDisabled, nil)
		if btnNote.Text == strNoButton {
			checkSpecial.Disable()
		}

		btnSave := widget.NewButton("Save", func() {
			data[rowToEdit][COLUMN_KEY] = btnNote.Text
			data[rowToEdit][COLUMN_PAYLOAD] = entryPayload.Text
			data[rowToEdit][COLUMN_DESCRIPTION] = entryDescription.Text
			data[rowToEdit][COLUMN_VELOCITY] = entryVelocity.Text
			if checkSpecial.Checked {
				data[rowToEdit][COLUMN_SPECIAL] = "true"
			} else {
				data[rowToEdit][COLUMN_SPECIAL] = "false"
			}
			table.Refresh()
			setPreferences(versionTool)
			popupEdit.Hide()
		})
		btnCancel := widget.NewButton("Cancel", func() {
			popupEdit.Hide()
		})

		btnNote.Text = data[rowToEdit][COLUMN_KEY]
		configureCheckSpecial(strSpecialDisabled)
		entryPayload.Text = data[rowToEdit][COLUMN_PAYLOAD]
		entryDescription.Text = data[rowToEdit][COLUMN_DESCRIPTION]
		entryVelocity.Text = data[rowToEdit][COLUMN_VELOCITY]
		checkSpecial.Checked = data[rowToEdit][COLUMN_SPECIAL] == "true"

		popupEdit.Content = container.NewVBox(container.New(layout.NewFormLayout(), lblNote, btnNote, lblPayload, container.NewVBox(comboPayload, entryPayload), lblDescription, entryDescription, lblVelocity, entryVelocity, lblToggle, checkSpecial), container.NewCenter(container.NewHBox(btnSave, btnCancel)))
		popupEdit.Resize(fyne.NewSize(400, 200))
		popupEdit.Show()
	})

	lblOutput = widget.NewLabel("")

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow, lblOutput)

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

	getPreferences(versionTool)

	w.ShowAndRun()
}

func configureCheckSpecial(strSpecialDisabled string) {
	if len(btnNote.Text) > 6 {
		checkSpecial.Disable()
		checkSpecial.Text = strSpecialDisabled
		checkSpecial.Refresh()
	} else {
		checkSpecial.Enable()
		switch btnNote.Text[:1] {
		case pkgMidi.MIDI_BUTTON:
			checkSpecial.Text = "Toggle LED"
		case pkgMidi.MIDI_KNOB:
			checkSpecial.Text = "Left = Inverse"
		case pkgMidi.MIDI_SLIDER:
			checkSpecial.Text = "Absolute Control"
		default:
			checkSpecial.Text = "Err: Unkown Midi Type"
			checkSpecial.Disable()
		}
		checkSpecial.Refresh()
	}
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
		var midiType string
		switch data[i][COLUMN_KEY][:1] {
		case pkgMidi.MIDI_BUTTON:
			midiType = pkgMidi.MIDI_BUTTON
		case pkgMidi.MIDI_KNOB:
			midiType = pkgMidi.MIDI_KNOB
		case pkgMidi.MIDI_SLIDER:
			midiType = pkgMidi.MIDI_SLIDER
		default:
			log.Println("ERROR fillMapKeys: Unknown midiType", data[i][COLUMN_KEY])
		}

		keyId, err := strconv.Atoi(data[i][COLUMN_KEY][1:]) // first char is midiType
		if err != nil {
			log.Printf("ERROR fillMapKeys: strconv.Atoi 1: %s\n", err)
			continue
		}

		vel, err := strconv.Atoi(data[i][COLUMN_VELOCITY]) // first char is midiType
		if err != nil {
			log.Printf("ERROR fillMapKeys: strconv.Atoi 2:%s\n", err)
			continue
		}

		mapKeys[uint8(keyId)] = pkgMidi.KeyStruct{
			MidiType: midiType,
			Key:      data[i][COLUMN_KEY],
			Payload:  data[i][COLUMN_PAYLOAD],
			Velocity: uint16(vel),
			Special:  data[i][COLUMN_SPECIAL] == "true",
		}
	}
}

func listen() {
	if btnListen.Text == strStartListen {
		fillMapKeys()
		pkgMidi.StartListen(table, lblOutput, data, comboSelect.Selected, mapKeys)
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

type ActiveHeader struct {
	widget.Label
	OnTapped func()
}

func headerCreate() fyne.CanvasObject {
	h := &ActiveHeader{}
	h.ExtendBaseWidget(h)
	h.SetText("000")
	return h
}

func headerUpdate(id widget.TableCellID, o fyne.CanvasObject) {
	header := o.(*ActiveHeader)
	header.TextStyle.Bold = true
	switch id.Col {
	case -1:
		header.SetText(strconv.Itoa(id.Row + 1))
	case 0:
		header.SetText("Key")
	case 1:
		header.SetText("Payload")
	case 2:
		header.SetText("Description")
	case 3:
		header.SetText("Velocity")
	case 4:
		header.SetText("Special")
	}

	// header.OnTapped = func() {
	// 	fmt.Printf("Header %d tapped\n", id.Col)
	// }
}
