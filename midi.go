package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var stop func() = nil
var out drivers.Out
var mapHotkeys map[uint8]string
var mapVelocity map[uint8]uint8
var mapCurrentVelocity map[uint8]uint8
var mapToggle map[uint8]string

var errMidiInAlsa = "MidiInAlsa: message queue limit reached!!"

type keyStruct struct {
	key           string
	hotkeyPayload string
	velocity      string
	toggle        bool
}

func initialize() string {
	rtmididrv.New() // Not needed, but rtmididrv needs to be called, so the import doesn't get removed
	return ""
}

func closeDriver() {
	stopListen()
	midi.CloseDriver()
}

func getInputPorts() []string {
	var ports []string

	portArr := strings.Split(midi.GetInPorts().String(), "]")
	for i := len(portArr) - 1; i > 0; i-- {
		port := strings.Split(portArr[i], ":")[0]
		ports = append(ports, strings.TrimSpace(port))
	}

	fmt.Println(ports)

	return ports
}

func getOneInput(device string) string {
	// prepare to listen ---------
	inPort := device
	in, err := midi.FindInPort(inPort)
	if err != nil {
		fmt.Println("can't find " + inPort)
		return "can't find " + inPort
	}

	returnVal := ""
	var m sync.Mutex

	stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel uint8
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			m.Lock()
			returnVal = strconv.Itoa(int(key))
			fmt.Println(returnVal)
			m.Unlock()
		default:
			fmt.Printf("received unsupported %s\n", msg)
			m.Lock()
			returnVal = "received unsupported" + msg.String()
			m.Unlock()
		}
	}, midi.UseSysEx())

	if err != nil && err.Error() != errMidiInAlsa {
		fmt.Printf("ERROR midi.ListenTo: %s\n", err)
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

	stop()
	if returnVal != "" {
		return returnVal
	} else {
		return "Nothing received"
	}
}

func doHotkey(ch uint8, key uint8) midi.Message {
	var ok bool
	var vel, curVel uint8
	var hotkey string
	if hotkey, ok = mapHotkeys[key]; !ok {
		fmt.Printf("DoHotkey: Key %v isn't assigned to a Hotkey\n", key)
		return nil
	}
	if vel, ok = mapVelocity[key]; !ok {
		fmt.Printf("DoHotkey: Key %v isn't assigned to a Velocity\n", key)
		return nil
	}
	if curVel, ok = mapCurrentVelocity[key]; !ok {
		fmt.Printf("DoHotkey: Key %v isn't assigned to a Current Velocity\n", key)
		return nil
	}

	switch {
	case strings.HasPrefix(hotkey, "Audio:"):
		hotkey = strings.TrimSpace(strings.TrimPrefix(hotkey, "Audio:"))
		hotkeyArr := strings.SplitN(hotkey, ":", 2)
		device := strings.TrimSpace(hotkeyArr[0])
		action := strings.TrimSpace(hotkeyArr[1])
		switch {
		case action == "(Un)Mute":
			switch {
			case device == "Input":
				exeCmd("amixer set Capture toggle")
			case device == "Output":
				exeCmd("amixer set Master toggle")
			default:
				for _, x := range getInputSinks() {
					if x.name == device {
						exeCmd("pactl set-sink-input-mute " + x.index + " toggle")
					}
				}
			}
		case strings.Contains(action, "+"), strings.Contains(action, "-"):
			action = strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
			switch {
			case device == "Input":
				exeCmd("amixer set Capture " + action[1:] + action[0:1])
			case device == "Output":
				exeCmd("amixer set Master " + action[1:] + action[0:1])
			default:
				for _, x := range getInputSinks() {
					if x.name == device {
						exeCmd("pactl set-sink-input-volume " + x.index + " " + action)
					}
				}
			}
		case strings.Contains(action, "="):
			action = strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
			action = strings.TrimSpace(strings.TrimPrefix(action, "="))
			switch {
			case device == "Input":
				exeCmd("amixer set Capture " + action)
			case device == "Output":
				exeCmd("amixer set Master " + action)
			default:
				for _, x := range getInputSinks() {
					if x.name == device {
						exeCmd("pactl set-sink-input-volume " + x.index + " " + action)
					}
				}
			}
		default:
			fmt.Printf("%s is no valid Audio command\n", hotkey)
		}
	case strings.HasPrefix(hotkey, "Keypress:"):
		val := strings.TrimSpace(strings.TrimPrefix(hotkey, "Keypress:"))
		valArr := strings.SplitN(val, ",", 2)

		if len(valArr) == 1 {
			robotgo.KeyTap(valArr[0])
		} else if len(valArr) > 1 {
			robotgo.KeyTap(valArr[0], strings.Split(valArr[1], ","))
		} else {
			fmt.Printf("%s is no valid Keypress command\n", hotkey)
		}

	case strings.HasPrefix(hotkey, "Write:"):
		val := strings.TrimSpace(strings.TrimPrefix(hotkey, "Write:"))
		robotgo.TypeStr(val)
	default:
		stdout, err := exeCmd(hotkey)

		if err != nil {
			fmt.Println("Error cmd.Output" + err.Error())
			break
		}

		fmt.Println("Output: " + string(stdout))
	}

	fmt.Printf("HOTKEY: %s\n", hotkey)
	var msg midi.Message
	if mapToggle[key] == "true" {
		if curVel == vel {
			vel = 0
		}
	}
	mapCurrentVelocity[key] = vel
	msg = midi.NoteOn(ch, key, vel)

	return msg
}

func startListen(device string, newMapHotkeys map[uint8]string, newMapVelocity map[uint8]uint8, newMapToogle map[uint8]string) string {
	mapHotkeys = newMapHotkeys
	mapVelocity = newMapVelocity
	mapToggle = newMapToogle

	// prepare to listen ---------
	inPort := device
	in, err := midi.FindInPort(inPort)
	if err != nil {
		fmt.Println("can't find " + inPort)
		return "can't find " + inPort
	}

	// prepare to send ----------
	outPort := device
	out, err = midi.FindOutPort(outPort)
	if err != nil {
		fmt.Printf("ERROR midi.FindOutPort: %s\n", err)
		return "ERROR midi.FindOutPort:  " + err.Error()
	}

	send, err := midi.SendTo(out)
	if err != nil {
		fmt.Printf("ERROR midi.SendTo: %s\n", err)
		return "ERROR midi.SendTo: " + err.Error()
	}

	// turn all lights off
	mapCurrentVelocity = make(map[uint8]uint8)
	for i := 0; i < 255; i++ {
		msg := midi.NoteOn(0, uint8(i), 0)
		err := send(msg)
		if err != nil {
			fmt.Printf("ERROR send: %s\n", err)
		}
		mapCurrentVelocity[uint8(i)] = 0
	}

	msg := midi.NoteOn(0, uint8(37), 255)
	err = send(msg)
	if err != nil {
		fmt.Printf("ERROR send: %s\n", err)
	}

	// listen ----------------------
	stop, err = midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel, cc, val uint8
		var rel int16
		var abs uint16
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s (int: %v) on channel %v with velocity %v\n", midi.Note(key), key, ch, vel)
			selectCell(key)
			msg = doHotkey(ch, key)
			if msg != nil {
				err := send(msg)
				if err != nil && err.Error() != errMidiInAlsa {
					fmt.Printf("ERROR send: %s\n", err)
				}
			}
			if mapToggle[key] == "false" {
				go func(ch uint8, key uint8) {
					time.Sleep(200 * time.Millisecond)
					msg = midi.NoteOn(ch, key, 0)
					if msg != nil {
						err := send(msg)
						if err != nil && err.Error() != errMidiInAlsa {
							fmt.Printf("ERROR send: %s\n", err)
						}
					}
				}(ch, key)
			}
		case msg.GetNoteEnd(&ch, &key):
			//fmt.Printf("ending note %s (int:%v) on channel %v\n", midi.Note(key), key, ch)
		case msg.GetControlChange(&ch, &cc, &val):
			fmt.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)
			/* not needed as this doesn't effect the lightning of the control
			msg = midi.NoteOn(ch, cc, 60)
			err := send(msg)
			if err != nil && err.Error() != errMidiInAlsa {
				fmt.Printf("ERROR send: %s\n", err)
			}
			*/
		case msg.GetPitchBend(&ch, &rel, &abs):
			fmt.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)
			/* Not needed as slider has no lightning
			msg = midi.Pitchbend(ch, rel)
			err := send(msg)
			if err != nil && err.Error() != errMidiInAlsa {
				fmt.Printf("ERROR send: %s\n", err)
			}
			*/
		default:
			fmt.Printf("received unsupported %s\n", msg)
		}
	}, midi.UseSysEx())

	if err != nil && err.Error() != errMidiInAlsa {
		fmt.Printf("ERROR midi.ListenTo: %s\n", err)
		return "ERROR midi.ListenTo: " + err.Error()
	}

	return ""
}

func stopListen() string {
	if out != nil {
		out.Close()
	} else {
		fmt.Println("out is nil")
	}
	if stop != nil {
		stop()
	} else {
		fmt.Println("stop is nil")
	}
	return ""
}
