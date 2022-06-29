package main

import (
	"fmt"
	"os"
	"time"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers/rtmididrv" // autoregisters driver
)

func start() {
	defer midi.CloseDriver()

	drivvv, _ := rtmididrv.New()
	fmt.Println(drivvv)

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
