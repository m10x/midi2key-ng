package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/itchyny/volume-go"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var stop func() = nil
var out drivers.Out
var mapHotkeys map[uint8]string
var mapVelocity map[uint8]uint8
var mapCurrentVelocity map[uint8]uint8

var errMidiInAlsa = "MidiInAlsa: message queue limit reached!!"

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
			stop()
		default:
			fmt.Printf("received unsupported %s\n", msg)
			m.Lock()
			returnVal = "received unsupported" + msg.String()
			m.Unlock()
			stop()
		}
	}, midi.UseSysEx())

	if err != nil && err.Error() != errMidiInAlsa {
		fmt.Printf("ERROR midi.ListenTo: %s\n", err)
		stop()
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
		switch {
		case strings.HasPrefix(hotkey, "(Un)Mute"):
			isMuted, err := volume.GetMuted()
			if err != nil {
				fmt.Printf("ERROR volume.GetMuted: %s\n", err)
			}
			if isMuted {
				err = volume.Unmute()
			} else {
				err = volume.Mute()
			}
			if err != nil {
				fmt.Printf("ERROR volume.(Un)Mute: %s\n", err)
			}
		case strings.HasPrefix(hotkey, "+"), strings.HasPrefix(hotkey, "-"):
			val := strings.TrimSuffix(hotkey, "%")
			val = strings.TrimSpace(val)
			diff, err := strconv.Atoi(val)
			if err != nil {
				fmt.Printf("ERROR strconv.Atoi: %s\n", err)
			}
			volume.IncreaseVolume(diff)
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
		var cmd *exec.Cmd
		args := strings.SplitN(hotkey, " ", 2)
		if len(args) == 1 {
			cmd = exec.Command(args[0])
		} else if len(args) == 2 {
			cmd = exec.Command(args[0], args[1])
		} else {
			fmt.Printf("%s is no valid command\n", hotkey)
		}
		stdout, err := cmd.Output()

		if err != nil {
			fmt.Println(err.Error())
			break
		}

		fmt.Println(string(stdout))
	}

	if curVel == vel {
		vel = 0
	}
	mapCurrentVelocity[key] = vel

	fmt.Printf("HOTKEY: %s\n", hotkey)

	msg := midi.NoteOn(ch, key, vel)
	return msg
}

func startListen(device string, newMapHotkeys map[uint8]string, newMapVelocity map[uint8]uint8) string {
	mapHotkeys = newMapHotkeys
	mapVelocity = newMapVelocity

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
			msg = doHotkey(ch, key)
			if msg != nil {
				err := send(msg)
				if err != nil && err.Error() != errMidiInAlsa {
					fmt.Printf("ERROR send: %s\n", err)
				}
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

func alteMain() {
	defer midi.CloseDriver()

	rtmididrv.New() // Not needed, but rtmididrv needs to be called, so the import doesn't get removed

	if len(os.Args) == 2 && os.Args[1] == "list" {
		fmt.Printf("MIDI IN Ports\n")
		fmt.Println(midi.GetInPorts())
		fmt.Printf("\n\nMIDI OUT Ports\n")
		fmt.Println(midi.GetOutPorts())
		fmt.Printf("\n\n")

	}

	inPort := "X-TOUCH MINI"
	in, err := midi.FindInPort(inPort)
	if err != nil {
		fmt.Println("can't find " + inPort)
		return
	}

	// prepare to send ----------
	outPort := inPort
	out, err := midi.FindOutPort(outPort)
	if err != nil {
		fmt.Println("can't find " + outPort)
		return
	}

	send, err := midi.SendTo(out)
	if err != nil {
		fmt.Printf("ERROR midi.SendTo: %s\n", err)
		return
	}
	// --------------------------

	// listen ----------------------
	stop, err := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		var bt []byte
		var ch, key, vel, cc, val uint8
		var rel int16
		var abs uint16
		switch {
		case msg.GetSysEx(&bt):
			fmt.Printf("got sysex: % X\n", bt)
		case msg.GetNoteStart(&ch, &key, &vel):
			fmt.Printf("starting note %s on channel %v with velocity %v\n", midi.Note(key), ch, vel)
			msg := midi.NoteOn(ch, key, vel)
			send(msg)
		case msg.GetNoteEnd(&ch, &key):
			fmt.Printf("ending note %s (int:%v) on channel %v\n", midi.Note(key), key, ch)
		case msg.GetControlChange(&ch, &cc, &val):
			fmt.Printf("control change %v %q channel: %v value: %v\n", cc, midi.ControlChangeName[cc], ch, val)
			msg := midi.NoteOn(ch, cc, 60)
			send(msg)
		case msg.GetPitchBend(&ch, &rel, &abs):
			fmt.Printf("pitch bend on channel %v: value: %v (rel) %v (abs)\n", ch, rel, abs)
			msg := midi.Pitchbend(ch, rel)
			send(msg)
		default:
			fmt.Printf("received unsupported %s\n", msg)
		}
	}, midi.UseSysEx())

	if err != nil {
		fmt.Printf("ERROR midi.ListenTo: %s\n", err)
		return
	}

	time.Sleep(time.Second * 5)

	stop()
	// ---------------------------------

}
