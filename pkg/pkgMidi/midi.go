package pkgMidi

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"

	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver, default driver, runs solid
)

const (
	MIDI_BUTTON = "B"
	MIDI_KNOB   = "K"
	MIDI_SLIDER = "S"
)

var (
	errMidiInAlsa             = "message queue limit reached"
	stop               func() = nil
	out                drivers.Out
	mapCurrentVelocity map[uint8]uint8
)

type KeyStruct struct {
	MidiType string
	Key      string
	Payload  string
	Velocity uint16
	Special  bool
	Held     bool
}

func GetInputPorts() []string {
	var ports []string

	log.Println(midi.GetInPorts().String())
	portArr := strings.Split(midi.GetInPorts().String(), "]")
	for i := len(portArr) - 1; i > 0; i-- {
		port := strings.Split(portArr[i], ":")[0]
		ports = append(ports, strings.TrimSpace(port))
	}

	log.Println(ports)

	return ports
}

func GetOneInput(device string) string {
	// prepare to listen ---------
	inPort := device
	in, err := midi.FindInPort(inPort)
	if err != nil {
		log.Println("can't find " + inPort)
		return "can't find " + inPort
	}

	returnVal := ""
	var m sync.Mutex

	stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel, cc, val uint8
		var rel int16
		var abs uint16
		switch {
		case msg.GetSysEx(&bt):
			log.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel), msg.GetNoteOn(&ch, &key, &vel):
			m.Lock()
			returnVal = MIDI_BUTTON + strconv.Itoa(int(key))
			log.Printf("starting note %s (int:%v) on channel %v\n", midi.Note(key), key, ch)
			m.Unlock()
		case msg.GetControlChange(&ch, &cc, &val):
			m.Lock()
			returnVal = MIDI_KNOB + strconv.Itoa(int(cc)) // use cc instead of key as identifier
			log.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)
			m.Unlock()
		case msg.GetPitchBend(&ch, &rel, &abs):
			m.Lock()
			returnVal = MIDI_SLIDER + strconv.Itoa(int(ch)) // use ch instead of key as identifier
			log.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)
			m.Unlock()
		default:
			log.Printf("received unsupported %s\n", msg)
			m.Lock()
			returnVal = "received unsupported" + msg.String()
			m.Unlock()
		}
	}, midi.UseSysEx())

	if err != nil && err.Error() != errMidiInAlsa {
		log.Printf("ERROR midi.ListenTo: %s\n", err)
		return "ERROR midi.ListenTo: " + err.Error()
	}

	for i := 0; i < 100; i++ {
		m.Lock()
		if returnVal != "" {
			m.Unlock()
			break
		}
		m.Unlock()
		time.Sleep(time.Millisecond * 50)
	}

	in.Close()
	if returnVal != "" {
		return returnVal
	} else {
		return "Nothing received"
	}
}

func selectCell(table *widget.Table, data [][]string, pressedKey uint8) {
	for i, x := range data {
		if x[0][1:] == strconv.Itoa(int((pressedKey))) {
			table.Select(widget.TableCellID{
				Row: i,
				Col: 0,
			})
			break
		}
	}
}

func StartListen(table *widget.Table, lblOutput *widget.Label, data [][]string, device string, mapKeys map[uint8]KeyStruct) string {

	// prepare to listen ---------
	inPort := device
	in, err := midi.FindInPort(inPort)
	if err != nil {
		log.Println("can't find " + inPort)
		return "can't find " + inPort
	}

	// prepare to send ----------
	outPort := device
	out, err = midi.FindOutPort(outPort)
	if err != nil {
		log.Printf("ERROR midi.FindOutPort: %s\n", err)
		return "ERROR midi.FindOutPort:  " + err.Error()
	}

	send, err := midi.SendTo(out)
	if err != nil {
		log.Printf("ERROR midi.SendTo: %s\n", err)
		return "ERROR midi.SendTo: " + err.Error()
	}

	// turn all lights off
	mapCurrentVelocity = make(map[uint8]uint8)
	for i := 0; i < 255; i++ {
		msg := midi.NoteOn(0, uint8(i), 0)
		err := send(msg)
		if err != nil {
			log.Printf("ERROR send: %s\n", err)
		}
		mapCurrentVelocity[uint8(i)] = 0
	}
	for i := 48; i < 56; i++ {
		msg := midi.ControlChange(0, uint8(i), 0)
		err := send(msg)
		if err != nil {
			log.Printf("ERROR send: %s\n", err)
		}
	}

	msg := midi.NoteOn(0, uint8(37), 255)
	err = send(msg)
	if err != nil {
		log.Printf("ERROR send: %s\n", err)
	}

	// listen ----------------------
	stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel, cc, val uint8
		var rel int16
		var abs uint16
		switch {
		case msg.GetSysEx(&bt):
			log.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			log.Printf("starting note %s (int: %v) on channel %v with velocity %v\n", midi.Note(key), key, ch, vel)
			selectCell(table, data, key)
			msg = doHotkey(lblOutput, mapKeys, ch, key, uint16(vel), true)
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}
			if !mapKeys[key].Special {
				go func(ch uint8, key uint8) {
					time.Sleep(200 * time.Millisecond)
					msg = midi.NoteOn(ch, key, 0)
					if msg != nil {
						err := send(msg)
						if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
							log.Printf("ERROR send: %s\n", err)
						}
					}
				}(ch, key)
			}
		case msg.GetNoteEnd(&ch, &key):
			// Not Used Currently, because a command would be run twice
			/*msg = doHotkey(lblOutput, mapKeys, ch, key, uint16(vel), false)
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}*/
		case msg.GetControlChange(&ch, &cc, &val):
			log.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)
			selectCell(table, data, cc)                                    // use cc instead of key as reference
			msg = doHotkey(lblOutput, mapKeys, ch, cc, uint16(val), false) // use cc instead of key as reference
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}
		case msg.GetPitchBend(&ch, &rel, &abs):
			log.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)
			selectCell(table, data, ch)                            // use ch instead of key as reference
			msg = doHotkey(lblOutput, mapKeys, ch, ch, abs, false) // use ch instead of key as reference
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}
		default:
			log.Printf("received unsupported %s\n", msg)
		}
	}, midi.UseSysEx())

	if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
		log.Printf("ERROR midi.ListenTo: %s\n", err)
		return "ERROR midi.ListenTo: " + err.Error()
	}

	return ""
}

func StopListen() string {
	if out != nil {
		out.Close()
	} else {
		log.Println("out is nil")
	}
	if stop != nil {
		stop()
	} else {
		log.Println("stop is nil")
	}
	return ""
}
