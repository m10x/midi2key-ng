package pkgMidi

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/widget"
	"github.com/go-vgo/robotgo"
	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"github.com/m10x/midi2key-ng/pkg/pkgUtils"
	"gitlab.com/gomidi/midi/v2"
)

var wayland bool
var checkOnce bool

func checkWayland() {
	// check if wayland is used, check only once
	if !checkOnce && runtime.GOOS == "linux" {
		value, exists := os.LookupEnv("XDG_SESSION_TYPE")
		if exists {
			if value == "wayland" {
				wayland = true
				log.Println("Detected Wayland")
			}
		} else {
			log.Println("XDG_SESSION_TYPE env does not exist")
		}
	}
	checkOnce = true
}

func keyTap(payload string) {
	if wayland {
		//usText := "Hello, Worldy! 123 $%^"
		// does not work properly
		//pkgCmd.ExeCmd("ydotool type '" + usText + "'")
	} else {
		payloadArr := strings.SplitN(payload, ",", 2)
		if len(payloadArr) == 1 {
			robotgo.KeyTap(payloadArr[0])
		} else if len(payloadArr) > 1 {
			robotgo.KeyTap(payloadArr[0], strings.Split(payloadArr[1], ","))
		} else {
			log.Printf("%s is no valid Keypress command\n", payload)
		}
	}
}

func keyDown(payload string) {
	payloadArr := strings.SplitN(payload, ",", 2)
	if len(payloadArr) == 1 {
		robotgo.KeyDown(payloadArr[0])
	} else if len(payloadArr) > 1 {
		for i, _ := range payloadArr {
			robotgo.KeyDown(payloadArr[i])
			log.Println(payloadArr[i] + " down")
		}

	} else {
		log.Printf("%s is no valid Keydown command\n", payload)
	}
}

func keyUp(payload string) {
	payloadArr := strings.SplitN(payload, ",", 2)
	if len(payloadArr) == 1 {
		robotgo.KeyUp(payloadArr[0])
	} else if len(payloadArr) > 1 {
		for i, _ := range payloadArr {
			robotgo.KeyUp(payloadArr[i])
			log.Println(payloadArr[i] + " up")
		}
	} else {
		log.Printf("%s is no valid Keydown command\n", payload)
	}
}

func doHotkey(lblOutput *widget.Label, mapKeys map[uint8]KeyStruct, ch uint8, key uint8, val uint16, keyState bool) midi.Message {
	var ok bool
	var curVel uint8

	checkWayland()

	if mapKeys[key].Key == "" {
		log.Printf("doHotkey: Key %v isn't assigned yet\n", key)
		return nil
	}

	if curVel, ok = mapCurrentVelocity[key]; !ok {
		log.Printf("doHotkey: Key %v isn't assigned to a Current Velocity\n", key)
		return nil
	}

	var newVolume, device, payload string
	log.Println("Payload: " + mapKeys[key].Payload)
	switch {
	case strings.HasPrefix(mapKeys[key].Payload, "Audio:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Audio:"))
		payloadArr := strings.SplitN(payload, ":", 3)
		device = strings.TrimSpace(payloadArr[0] + ":" + payloadArr[1])
		if device == "App: Focused Application" {
			log.Println("Checking the Focused Application")
			focusedWindowPID, err := pkgCmd.GetFocusedWindowPID()
			if err != nil {
				log.Println("Error getting the focused window PID:" + err.Error())
				return nil
			}
			log.Println("The Focused Application's PID is " + focusedWindowPID)
			focusedWindowAudioSink, err := pkgCmd.GetPulseAudioSinkDescriptionByPID(focusedWindowPID)
			if err != nil {
				log.Println("Error getting the audio sink of the focused window:" + err.Error())
				return nil
			}
			device = "App: " + focusedWindowAudioSink
		}
		log.Println("Device: " + device)
		action := strings.TrimSpace(payloadArr[2])

		if mapKeys[key].MidiType == MIDI_KNOB && mapKeys[key].Special && val > uint16(mapKeys[key].Velocity) {
			if strings.Contains(action, "+") {
				action = strings.Replace(action, "+", "-", -1)
			} else {
				action = strings.Replace(action, "-", "+", -1)
			}
		}

		if mapKeys[key].MidiType == MIDI_SLIDER && mapKeys[key].Special && strings.Contains(action, "=") {
			percent := (float32(val) * 100) / float32(mapKeys[key].Velocity)
			log.Printf("%f = (%d * 100) / %d\n", percent, val, mapKeys[key].Velocity)

			action = pkgUtils.ReplaceStringInBetween(action, "=", "%", fmt.Sprintf("%d", int(percent)))
			log.Println("Replaced Volume with ", action)

		}

		switch {
		case strings.HasPrefix(device, "In:"):
			for _, x := range pkgCmd.GetSources() {
				if x.Description == device {
					switch {
					case action == "(Un)Mute":
						pkgCmd.ExeCmd("pactl set-source-mute " + x.Name + " toggle")
					case strings.Contains(action, "+"), strings.Contains(action, "-"):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						if !(strings.Contains(action, "+") && x.Volume >= 100) { // Dont increase further if volume is already >= 100%
							pkgCmd.ExeCmd("pactl set-source-volume " + x.Name + " " + actionTrimmed)
						}
						newVolume, _ = pkgCmd.ExeCmd("pactl get-source-volume " + x.Name)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-source-volume " + x.Name + " " + actionTrimmed)
						newVolume, _ = pkgCmd.ExeCmd("pactl get-source-volume " + x.Name)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, payloadArr)
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
						if !(strings.Contains(action, "+") && x.Volume >= 100) { // Dont increase further if volume is already >= 100%
							pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
						}
						newVolume, _ = pkgCmd.ExeCmd("pactl get-sink-volume " + x.Name)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
						newVolume, _ = pkgCmd.ExeCmd("pactl get-sink-volume " + x.Name)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, payloadArr)
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
						if !(strings.Contains(action, "+") && x.Volume >= 100) { // Dont increase further if volume is already >= 100%
							pkgCmd.ExeCmd("pactl set-sink-input-volume " + x.Index + " " + actionTrimmed)
						}
						newVolume = strconv.Itoa(pkgCmd.GetAppVolume(device))
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-sink-input-volume " + x.Index + " " + actionTrimmed)
					default:
						log.Printf("Audio action %s is unknown (%s)\n", action, payloadArr)
					}
				}
			}
		default:
			log.Printf("Device-Type %s is unknown\n", device)
		}
	case strings.HasPrefix(mapKeys[key].Payload, "Keypress:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Keypress:"))
		if mapKeys[key].Held {
			log.Println("Held is on")
			if keyState == true { //Is it a keydown event?
				keyDown(payload)
			} else { //If not do a keyUp instead
				keyUp(payload)
			}

		} else {
			log.Println("Held is off")
			if keyState == true {
				keyTap(payload)
			}
		}
	case strings.HasPrefix(mapKeys[key].Payload, "Write:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Write:"))
		payload = strings.ReplaceAll(payload, "'", "'\\''")
		pkgCmd.ExeCmd(`echo type '` + payload + `' | dotoolc`)
		//REMOVE robotgo.TypeStr(payload)
	default:
		payload = mapKeys[key].Payload
		stdout, err := pkgCmd.ExeCmd(payload)

		if err != nil {
			log.Println("Error cmd.Output" + err.Error())
			break
		}

		log.Println("Output: " + string(stdout))
	}

	log.Printf("HOTKEY: %s\n", payload)
	var msg midi.Message

	vel := uint8(mapKeys[key].Velocity)
	if mapKeys[key].MidiType == MIDI_BUTTON && mapKeys[key].Special {
		if uint16(curVel) == mapKeys[key].Velocity {
			log.Printf("Set vel 0\n")
			vel = 0
		}
	}
	mapCurrentVelocity[key] = vel
	switch mapKeys[key].MidiType {
	case MIDI_BUTTON:
		log.Printf("Simulate Note on, ch=%d, key=%d, vel=%d\n", ch, key, vel)
		msg = midi.NoteOn(ch, key, vel)
		lblOutput.Text = payload
		lblOutput.Refresh()
	// Behringer X-Touch Mini https://stackoverflow.com/a/49740979
	case MIDI_KNOB:
		if strings.Contains(newVolume, "/") {
			newVolume = pkgUtils.GetStringInBetween(newVolume, " / ", "%")
		}
		key += 32
		vel = 33
		// is it used to control volume?
		if newVolume != "" {
			i, err := strconv.Atoi(strings.TrimSpace(newVolume))
			if err != nil {
				log.Print(err.Error())
			}
			log.Print("i", i)
			vel += uint8(i / 10)
			log.Print("Velocity", vel)
		} else { // not used for volume control, show the velocity used
			if val <= 30 {
				vel = 54
			} else {
				vel = 49
			}
		}
		log.Printf("Simulate Controlchange, ch=%d, cc=%d, vel=%d\n", ch, key, vel)
		msg = midi.ControlChange(ch, key, vel)
		log.Print("newVolume", newVolume)
		lblOutput.Text = device + newVolume
		lblOutput.Refresh()
	}

	return msg
}
