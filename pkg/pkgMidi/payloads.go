package pkgMidi

import (
	"log"
	"strings"

	"github.com/go-vgo/robotgo"
	"github.com/m10x/midi2key-ng/pkg/pkgCmd"
	"gitlab.com/gomidi/midi/v2"
)

func doHotkey(mapKeys map[uint8]KeyStruct, ch uint8, key uint8, midiType string) midi.Message {
	var ok bool
	var vel, curVel uint8

	if mapKeys[key].Key == "" {
		log.Printf("doHotkey: Key %v isn't assigned yet\n", key)
		return nil
	}

	if curVel, ok = mapCurrentVelocity[key]; !ok {
		log.Printf("doHotkey: Key %v isn't assigned to a Current Velocity\n", key)
		return nil
	}

	switch {
	case strings.HasPrefix(mapKeys[key].Payload, "Audio:"):
		payload := strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Audio:"))
		payloadArr := strings.SplitN(payload, ":", 3)
		device := strings.TrimSpace(payloadArr[0] + ":" + payloadArr[1])
		action := strings.TrimSpace(payloadArr[2])
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
						pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
					case strings.Contains(action, "="):
						actionTrimmed := strings.TrimSpace(strings.TrimPrefix(action, "Volume"))
						actionTrimmed = strings.TrimSpace(strings.TrimPrefix(actionTrimmed, "="))
						pkgCmd.ExeCmd("pactl set-sink-volume " + x.Name + " " + actionTrimmed)
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
						pkgCmd.ExeCmd("pactl set-sink-input-volume " + x.Index + " " + actionTrimmed)
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
			log.Printf("Device-Type %s is unkown\n", device)
		}
	case strings.HasPrefix(mapKeys[key].Payload, "Keypress:"):
		payload := strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Keypress:"))
		payloadArr := strings.SplitN(payload, ",", 2)

		if len(payloadArr) == 1 {
			robotgo.KeyTap(payloadArr[0])
		} else if len(payloadArr) > 1 {
			robotgo.KeyTap(payloadArr[0], strings.Split(payloadArr[1], ","))
		} else {
			log.Printf("%s is no valid Keypress command\n", mapKeys[key].Payload)
		}

	case strings.HasPrefix(mapKeys[key].Payload, "Write:"):
		payload := strings.TrimSpace(strings.TrimPrefix(mapKeys[key].Payload, "Write:"))
		robotgo.TypeStr(payload)
	default:
		stdout, err := pkgCmd.ExeCmd(mapKeys[key].Payload)

		if err != nil {
			log.Println("Error cmd.Output" + err.Error())
			break
		}

		log.Println("Output: " + string(stdout))
	}

	log.Printf("HOTKEY: %s\n", mapKeys[key].Payload)
	var msg midi.Message
	if mapKeys[key].Toggle {
		if curVel == mapKeys[key].Velocity {
			vel = 0
		}
	}
	mapCurrentVelocity[key] = vel
	if midiType == MIDI_BUTTON { // Others arent supported yet
		msg = midi.NoteOn(ch, key, vel)
	}

	return msg
}
