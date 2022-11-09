package pkgMidi

import (
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"

	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
	//_ "gitlab.com/gomidi/midi/v2/drivers/portmididrv" // autoregisters driver
)

const (
	MIDI_BUTTON = 0
	MIDI_KNOB   = 1
	MIDI_SLIDER = 2
)

var (
	errMidiInAlsa             = "message queue limit reached"
	stop               func() = nil
	out                drivers.Out
	mapHotkeys         map[uint8]string
	mapVelocity        map[uint8]uint8
	mapCurrentVelocity map[uint8]uint8
	mapToggle          map[uint8]string
)

type KeyStruct struct {
	MidiType      int // unnötig, kann entfernt werden...
	Key           string
	HotkeyPayload string
	Velocity      string
	Toggle        bool
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
			returnVal = "B" + strconv.Itoa(int(key))
			log.Printf("starting note %s (int:%v) on channel %v\n", midi.Note(key), key, ch)
			m.Unlock()
		case msg.GetControlChange(&ch, &cc, &val):
			m.Lock()
			returnVal = "K" + strconv.Itoa(int(cc)) // use cc instead of key as identifier
			log.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)

			/* not needed as this doesn't effect the lightning of the control
			msg = midi.NoteOn(ch, cc, 60)
			err := send(msg)
			if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
				log.Printf("ERROR send: %s\n", err)
			}
			*/

			m.Unlock()
		case msg.GetPitchBend(&ch, &rel, &abs):
			m.Lock()
			returnVal = "S" + strconv.Itoa(int(ch)) // use ch instead of key as identifier
			log.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)

			/* Not needed as slider has no lightning
			msg = midi.Pitchbend(ch, rel)
			err := send(msg)
			if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
				log.Printf("ERROR send: %s\n", err)
			}
			*/

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

func doHotkey(ch uint8, key uint8, midiType int) midi.Message {
	var ok bool
	var vel, curVel uint8
	var hotkey string
	if hotkey, ok = mapHotkeys[key]; !ok {
		log.Printf("DoHotkey: Key %v isn't assigned to a Hotkey\n", key)
		return nil
	}
	if vel, ok = mapVelocity[key]; !ok {
		log.Printf("DoHotkey: Key %v isn't assigned to a Velocity\n", key)
		return nil
	}
	if curVel, ok = mapCurrentVelocity[key]; !ok {
		log.Printf("DoHotkey: Key %v isn't assigned to a Current Velocity\n", key)
		return nil
	}

	log.Println(mapHotkeys)
	log.Println("KEy", key)

	switch {
	case strings.HasPrefix(hotkey, "Audio:"):
		hotkey = strings.TrimSpace(strings.TrimPrefix(hotkey, "Audio:"))
		hotkeyArr := strings.SplitN(hotkey, ":", 3)
		device := strings.TrimSpace(hotkeyArr[0] + ":" + hotkeyArr[1])
		action := strings.TrimSpace(hotkeyArr[2])
		switch {
		case strings.HasPrefix(device, "In:"):
			for _, x := range pkgCmd.GetSources() {
				if x.Description == device {
					switch {
					case action == "(Un)Mute":
						pkgCmd.ExeCmd("pactl set-source-mute " + x.Name + " toggle")
					case strings.Contains(action, "+"), strings.Contains(action, "-"):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						pkgCmd.ExeCmd("pactl set-source-volume " + x.Name + " " + actionTrimmed)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-source-volume " + x.Name + " " + actionTrimmed)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, hotkeyArr)
					}
				}
			}
		case strings.HasPrefix(device, "Out:"):
			for _, x := range pkgCmd.GetSinks() {
				if x.Description == device {
					switch {
					case action == "(Un)Mute":
						pkgCmd.ExeCmd("pactl set-sink-mute " + x.Name + " toggle")
					case strings.Contains(action, "+"), strings.Contains(action, "-"):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, hotkeyArr)
					}
				}
			}
		case strings.HasPrefix(device, "App:"):
			for _, x := range pkgCmd.GetSinkInputs() {
				if x.Description == device {
					switch {
					case action == "(Un)Mute":
						pkgCmd.ExeCmd("pactl set-sink-input-mute " + x.Index + " toggle")
					case strings.Contains(action, "+"), strings.Contains(action, "-"):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						pkgCmd.ExeCmd("pactl set-sink-input-volume " + x.Index + " " + actionTrimmed)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-sink-input-volume " + x.Index + " " + actionTrimmed)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, hotkeyArr)
					}
				}
			}
		default:
			log.Printf("Device-Type %s is unkown\n", device)
		}
	case strings.HasPrefix(hotkey, "Keypress:"):
		val := strings.TrimSpace(strings.TrimPrefix(hotkey, "Keypress:"))
		valArr := strings.SplitN(val, ",", 2)

		if len(valArr) == 1 {
			robotgo.KeyTap(valArr[0])
		} else if len(valArr) > 1 {
			robotgo.KeyTap(valArr[0], strings.Split(valArr[1], ","))
		} else {
			log.Printf("%s is no valid Keypress command\n", hotkey)
		}

	case strings.HasPrefix(hotkey, "Write:"):
		val := strings.TrimSpace(strings.TrimPrefix(hotkey, "Write:"))
		robotgo.TypeStr(val)
	default:
		stdout, err := pkgCmd.ExeCmd(hotkey)

		if err != nil {
			log.Println("Error cmd.Output" + err.Error())
			break
		}

		log.Println("Output: " + string(stdout))
	}

	log.Printf("HOTKEY: %s\n", hotkey)
	var msg midi.Message
	if mapToggle[key] == "true" {
		if curVel == vel {
			vel = 0
		}
	}
	mapCurrentVelocity[key] = vel
	if midiType == MIDI_BUTTON { // Others arent supported yet
		msg = midi.NoteOn(ch, key, vel)
	}

	return msg
}

func selectCell(table *widget.Table, data [][]string, pressedKey uint8) {
	for i, x := range data {
		if x[0][1:] == strconv.Itoa(int((pressedKey))) {
			log.Println("found!\n\n")
			table.Select(widget.TableCellID{
				Row: i,
				Col: 0})
			break
		}
	}
}

func StartListen(table *widget.Table, data [][]string, device string, newMapHotkeys map[uint8]string, newMapVelocity map[uint8]uint8, newMapToogle map[uint8]string) string {
	mapHotkeys = newMapHotkeys
	mapVelocity = newMapVelocity
	mapToggle = newMapToogle

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
			msg = doHotkey(ch, key, MIDI_BUTTON)
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}
			if mapToggle[key] == "false" {
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
			//log.Printf("ending note %s (int:%v) on channel %v\n", midi.Note(key), key, ch)
		case msg.GetControlChange(&ch, &cc, &val):
			log.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)
			selectCell(table, data, cc)       // use cc instead of key as reference
			msg = doHotkey(ch, cc, MIDI_KNOB) // use cc instead of key as reference
			if msg != nil {
				err := send(msg)
				if err != nil && !strings.Contains(err.Error(), errMidiInAlsa) {
					log.Printf("ERROR send: %s\n", err)
				}
			}
		case msg.GetPitchBend(&ch, &rel, &abs):
			log.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)
			selectCell(table, data, ch)         // use ch instead of key as reference
			msg = doHotkey(ch, ch, MIDI_SLIDER) // use ch instead of key as reference
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
