package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

var stop func() = nil
var out drivers.Out
var keyMap map[uint8]int

func initialize() string {
	rtmididrv.New() // Not needed, but rtmididrv needs to be called, so the import doesn't get removed
	return ""
}

func closeDriver() {
	stopListen()
	midi.CloseDriver()
}

func getInputPorts() string {
	return midi.GetInPorts().String()
}

func startListen(device string) string {
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
		fmt.Println("can't find " + outPort)
		return "can't find " + outPort
	}

	send, err := midi.SendTo(out)
	if err != nil {
		fmt.Printf("ERROR midi.SendTo: %s\n", err)
		return "ERROR midi.SendTo: " + err.Error()
	}

	// turn all lights off
	keyMap = make(map[uint8]int)
	for i := 0; i < 255; i++ {
		msg := midi.NoteOn(0, uint8(i), 0)
		err := send(msg)
		if err != nil {
			fmt.Printf("ERROR send: %s\n", err)
		}
	}

	msg := midi.NoteOn(0, uint8(37), 255)
	err = send(msg)
	if err != nil {
		fmt.Printf("ERROR send: %s\n", err)
	}

	errMidiInAlsa := "MidiInAlsa: message queue limit reached!!"

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
			state, ok := keyMap[key]
			if !ok || state == 0 {
				vel = 255
				keyMap[key] = 1
			} else {
				vel = 0
				keyMap[key] = 0
			}
			msg = midi.NoteOn(ch, key, vel)
			err := send(msg)
			if err != nil && err.Error() != errMidiInAlsa {
				fmt.Printf("ERROR send: %s\n", err)
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

	if err != nil {
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
