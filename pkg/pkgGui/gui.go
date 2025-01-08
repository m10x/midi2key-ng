package pkgGui

import (
	"crypto/ed25519"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/fynelabs/fyneselfupdate"
	"github.com/fynelabs/selfupdate"
	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"github.com/m10x/midi2key-ng/pkg/pkgControllers"
	"github.com/m10x/midi2key-ng/pkg/pkgMidi"
)

const (
	COLUMN_KEY         = 0
	COLUMN_PAYLOAD     = 1
	COLUMN_DESCRIPTION = 2
	COLUMN_VELOCITY    = 3
	COLUMN_SPECIAL     = 4
	COLUMN_COUNT       = 5
	MAX_LOG_LINES      = 100
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
	btnMoveRowUp   *widget.Button
	btnMoveRowDown *widget.Button
	btnShowLog     *widget.Button
	btnLoadConfig  *widget.Button
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
	popupConfig    *widget.PopUp
	popupLog       *widget.PopUp
	entryLog       *widget.Entry
)

func enableRowButtons() {
	if selectedCell.Row >= 0 && selectedCell.Row < len(data) {
		btnDeleteRow.Enable()
		btnEditRow.Enable()
	} else {
		btnDeleteRow.Disable()
		btnEditRow.Disable()
	}
	if selectedCell.Row > 0 {
		btnMoveRowUp.Enable()
	} else {
		btnMoveRowUp.Disable()
	}
	if selectedCell.Row < len(data)-1 {
		btnMoveRowDown.Enable()
	} else {
		btnMoveRowDown.Disable()
	}
}

func selfManage(a fyne.App, w fyne.Window, sourceURL string) {
	// Used `selfupdatectl create-keys` followed by `selfupdatectl print-key`
	publicKey := ed25519.PublicKey{92, 160, 144, 239, 198, 220, 223, 157, 245, 210, 226, 218, 96, 33, 135, 235, 59, 40, 171, 175, 247, 183, 212, 247, 115, 23, 226, 247, 239, 148, 90, 54}
	httpSource := selfupdate.NewHTTPSource(nil, sourceURL)
	log.Println("Checking for new version")
	config := fyneselfupdate.NewConfigWithTimeout(a, w, time.Duration(1)*time.Minute,
		httpSource,
		selfupdate.Schedule{FetchOnStart: true, Interval: time.Hour * time.Duration(24)}, // Checking for binary update on start and every 24 hours
		publicKey)

	selfupdate.LogError = log.Printf
	selfupdate.LogInfo = log.Printf
	selfupdate.LogDebug = log.Printf

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}
}

type CustomLogWriter struct {
	lines    []string
	maxLines int
	mu       sync.RWMutex
}

func NewCustomLogWriter(maxLines int) *CustomLogWriter {
	return &CustomLogWriter{
		lines:    make([]string, 0, maxLines),
		maxLines: maxLines,
	}
}

func (w *CustomLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	if len(w.lines) >= w.maxLines {
		w.lines = w.lines[1:]
	}
	w.lines = append(w.lines, strings.TrimSpace(string(p)))
	w.mu.Unlock()

	return len(p), nil
}

func (w *CustomLogWriter) GetLogs() string {
	w.mu.RLock()
	text := strings.Join(w.lines, "\n")
	w.mu.RUnlock()
	return text
}

func Startup(versionTool string) {
	a = app.NewWithID("de.m10x.midi2key-ng")
	app.SetMetadata(fyne.AppMetadata{
		ID:      "de.m10x.midi2key-ng",
		Name:    "midi2key-ng",
		Version: versionTool,
	})
	w := a.NewWindow(a.Metadata().Name + " " + versionTool)
	w.Resize(fyne.NewSize(1055, 400))

	selfManage(a, w, "https://github.com/m10x/midi2key-ng/releases/latest/download/midi2key-ng-{{.OS}}-{{.Arch}}")

	mapKeys = make(map[uint8]pkgMidi.KeyStruct)

	vboxConfigs := container.NewVBox()
	for _, controller := range pkgControllers.Controllers {
		vboxConfigs.Add(widget.NewButton(controller.Name, func() {
			data = controller.Data
		}))
	}

	btnCloseConfig := widget.NewButton("Close", func() {
		popupConfig.Hide()
	})

	entryCurrentConfig := widget.NewEntry()

	borderConfig := container.NewBorder(entryCurrentConfig, btnCloseConfig, nil, nil, vboxConfigs)

	popupConfig = widget.NewModalPopUp(borderConfig, w.Canvas())
	popupConfig.Resize(fyne.NewSize(800, 400))

	btnLoadConfig = widget.NewButton("Load Config", func() {
		entryCurrentConfig.SetText(pkgControllers.PrintCurrentConfigAsCode(data))
		popupConfig.Show()
	})

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
		selectedCell = id // important for deleteRow, editRow, moveRowUp, moveRowDown

		if btnListen.Text == strStartListen {
			enableRowButtons()
		}
	}

	table.SetColumnWidth(0, 39)
	table.SetColumnWidth(1, 475)
	table.SetColumnWidth(2, 360)
	table.SetColumnWidth(3, 55)
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
	btnDeleteRow.Disable()
	btnEditRow = widget.NewButton("Edit Row", func() {
		rowToEdit := selectedCell.Row
		popupEdit := widget.NewModalPopUp(nil, w.Canvas())

		lblNote := widget.NewLabel("Note:")
		strNoButton := "Press this Button"
		strSpecialDisabled := "Waiting for key..."
		btnNote = widget.NewButton(strNoButton, func() {
			btnNote.Text = "Listening for Input..."
			btnNote.Disable()
			newInput := pkgMidi.GetOneInput(comboSelect.Selected)
			// Check if input was already assigned
			for row := range data {
				if data[row][COLUMN_KEY] == newInput && newInput != "Nothing received" {
					btnNote.Text = "Input is already assigned"
					break
				}
				btnNote.Text = newInput
			}
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
		comboPayload = widget.NewSelect([]string{"Command Line Command", "Keypress (combo)", "Write String", "Audio Control", "Soundboard"}, func(value string) {
			switch comboPayload.SelectedIndex() {
			case 0:
				entryPayload.Text = "Cmd: echo type 'example' | dotoolc"
				entryPayload.Refresh()
			case 1:
				entryPayload.Text = "Keypress: ctrl+shift+a super+alt+altgr+1"
				entryPayload.Refresh()
			case 2:
				entryPayload.Text = "Write: example"
				entryPayload.Refresh()
			case 3:
				devices := []string{"App: Focused Application"}
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
			case 4:
				entryPayload.Text = "Sound: /tmp/foo.wav"
				entryPayload.Refresh()
			}
			log.Println(entryPayload.Text)
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
	btnEditRow.Disable()
	btnMoveRowUp = widget.NewButton("", func() {
		if selectedCell.Row > 0 {
			// swap rows
			data[selectedCell.Row], data[selectedCell.Row-1] = data[selectedCell.Row-1], data[selectedCell.Row]
			selectedCell.Row--

			// select row again
			table.Select(widget.TableCellID{
				Row: selectedCell.Row,
				Col: 0,
			})

			table.Refresh()
			setPreferences(versionTool)
		} else {
			log.Println("Cannot move up: Already at the top row")
		}
	})
	btnMoveRowUp.Icon = theme.Icon(theme.IconNameArrowDropUp)
	btnMoveRowUp.Disable()

	btnMoveRowDown = widget.NewButton("", func() {
		if selectedCell.Row < len(data)-1 {
			// swap rows
			data[selectedCell.Row], data[selectedCell.Row+1] = data[selectedCell.Row+1], data[selectedCell.Row]
			selectedCell.Row++

			// select row again
			table.Select(widget.TableCellID{
				Row: selectedCell.Row,
				Col: 0,
			})

			table.Refresh()
			setPreferences(versionTool)
		} else {
			log.Println("Cannot move down: Already at the bottom row")
		}
	})
	btnMoveRowDown.Icon = theme.Icon(theme.IconNameArrowDropDown)
	btnMoveRowDown.Disable()

	lblOutput = widget.NewLabel("")

	entryLog = widget.NewMultiLineEntry()
	entryLog.Wrapping = fyne.TextWrapWord

	entryLog.OnChanged = func(newMsg string) {
		// Update the cursor to the end of the text
		entryLog.CursorRow = len(entryLog.Text) - 1
	}

	logWriter := NewCustomLogWriter(MAX_LOG_LINES)
	multiWriter := io.MultiWriter(os.Stdout, logWriter)
	log.SetOutput(multiWriter)

	btnCopyLog := widget.NewButton("Copy Log to Clipboard", func() {
		w.Clipboard().SetContent(entryLog.Text)
	})
	btnRefreshLog := widget.NewButton("Refresh Log", func() {
		entryLog.SetText(logWriter.GetLogs())
		entryLog.CursorRow = len(entryLog.Text) - 1
	})
	btnCloseLog := widget.NewButton("Close Log", func() {
		popupLog.Hide()
	})
	logHBox := container.NewHBox(btnCopyLog, layout.NewSpacer(), widget.NewLabel(strconv.Itoa(int(MAX_LOG_LINES))+" last log lines"), layout.NewSpacer(), btnRefreshLog, btnCloseLog)

	logBorder := container.NewBorder(nil, logHBox, nil, nil, entryLog)
	popupLog = widget.NewModalPopUp(logBorder, w.Canvas())
	popupLog.Resize(fyne.NewSize(800, 400))

	btnShowLog = widget.NewButton("Show Log", func() {
		entryLog.SetText(logWriter.GetLogs())
		entryLog.CursorRow = len(entryLog.Text) - 1
		popupLog.Show()
	})

	hBoxTable := container.NewHBox(btnAddRow, btnEditRow, btnDeleteRow, btnMoveRowUp, btnMoveRowDown, lblOutput, layout.NewSpacer(), btnShowLog)

	w.SetContent(container.NewBorder(
		container.NewBorder(nil, nil, btnLoadConfig, hBoxSelect, comboSelect), hBoxTable, nil, nil,
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
	} else {
		log.Println("Can't create TrayMenu")
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
			checkSpecial.Text = "Err: Unknown Midi Type"
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
	for i := 0; i < len(data); i++ {
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
		btnMoveRowUp.Disable()
		btnMoveRowDown.Disable()
		btnLoadConfig.Disable()
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
		btnLoadConfig.Enable()
		enableRowButtons()
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
		header.SetText("Vel")
	case 4:
		header.SetText("Special")
	}

	// header.OnTapped = func() {
	// 	fmt.Printf("Header %d tapped\n", id.Col)
	// }
}
