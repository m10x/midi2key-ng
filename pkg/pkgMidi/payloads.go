package pkgMidi

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/widget"
	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"github.com/m10x/midi2key-ng/pkg/pkgUtils"
	"gitlab.com/gomidi/midi/v2"
)

func doHotkey(lblOutput *widget.Label, mapKeys map[uint8]KeyStruct, ch uint8, key uint8, val uint16, keyState bool) midi.Message {
	var ok bool
	var curVel uint8

	if mapKeys[key].Key == "" {
		log.Printf("doHotkey: Key %v isn't assigned yet\n", key)
		return nil
	}

	if curVel, ok = mapCurrentVelocity[key]; !ok {
		log.Printf("doHotkey: Key %v isn't assigned to a Current Velocity\n", key)
		return nil
	}

	log.Printf("Payload: %s\n", mapKeys[key].Payload)
	lblOutput.Text = mapKeys[key].Payload
	lblOutput.Refresh()

	var newVolume, device, payload string
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
		pkgCmd.ExeCmdRoutine(`echo key '` + payload + `' | dotoolc`)
	case strings.HasPrefix(mapKeys[key].Payload, "Write:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Write:"))
		payload = strings.ReplaceAll(payload, "'", "'\\''")
		pkgCmd.ExeCmdRoutine(`echo type '` + payload + `' | dotoolc`)
	case strings.HasPrefix(mapKeys[key].Payload, "Sound:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Sound:"))
		pkgCmd.ExeCmdRoutine(`paplay --device=soundboard_router -p ` + payload + ` --volume=32768`)
	case strings.HasPrefix(mapKeys[key].Payload, "Cmd:"):
		payload = strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Cmd:"))
		pkgCmd.ExeCmdRoutine(payload)
	default:
		log.Println("Unknown payload type " + mapKeys[key].Payload)
	}

	vel := uint8(mapKeys[key].Velocity)
	if mapKeys[key].MidiType == MIDI_BUTTON && mapKeys[key].Special {
		if uint16(curVel) == mapKeys[key].Velocity {
			log.Printf("Set vel 0\n")
			vel = 0
		}
	}
	mapCurrentVelocity[key] = vel

	var msg midi.Message
	switch mapKeys[key].MidiType {
	case MIDI_BUTTON:
		log.Printf("Simulate Note on, ch=%d, key=%d, vel=%d\n", ch, key, vel)
		msg = midi.NoteOn(ch, key, vel)
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
	}

	return msg
}
