package pkgCmd

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

const (
	DEV_SINK       = 0
	DEV_SOURCE     = 1
	DEV_SINK_INPUT = 2
)

type ApplicationSinkStruct struct {
	Index       string // Sink Input #
	Name        string // sinkinput: sink: Name
	Description string // sinkinput: application.name, sink: Description

	mute    bool // Mute
	volume  int  // Volume (in %)
	devType int  // Sink, Source or Sinkinput? Ist das nötig? Hole es aktuell aus dem String
}

// https://stackoverflow.com/a/20438245
func ExeCmd(cmd string) ([]byte, error) {
	log.Printf("command is %s\n", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	out, err := exec.Command(head, parts...).Output()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func GetSinks() []ApplicationSinkStruct {
	out, err := ExeCmd("pactl list sinks")
	if err != nil {
		log.Println("Error in getSinks, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Sink #")

	var apps []ApplicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp ApplicationSinkStruct
		addApp.Index = strings.SplitN(sink, "\n", 2)[0]
		addApp.Name = GetStringInBetween(sink, "Name: ", "\n")
		addApp.Description = "Out: " + GetStringInBetween(sink, "Description: ", "\n")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			log.Println("Error in getSinks, while converting Volume String to int: " + err.Error())
		}
		addApp.devType = DEV_SINK

		apps = append(apps, addApp)
	}

	return apps
}

func GetSources() []ApplicationSinkStruct {
	out, err := ExeCmd("pactl list sources")
	if err != nil {
		log.Println("Error in getSources, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Source #")

	var apps []ApplicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp ApplicationSinkStruct
		addApp.Index = strings.SplitN(sink, "\n", 2)[0]
		addApp.Name = GetStringInBetween(sink, "Name: ", "\n")
		addApp.Description = "In: " + GetStringInBetween(sink, "Description: ", "\n")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			log.Println("Error in getSources, while converting Volume String to int: " + err.Error())
		}
		addApp.devType = DEV_SOURCE

		apps = append(apps, addApp)
	}

	return apps
}

func GetSinkInputs() []ApplicationSinkStruct {
	out, err := ExeCmd("pactl list sink-inputs")
	if err != nil {
		log.Println("Error in getSinkInputs, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Sink Input #")

	var apps []ApplicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp ApplicationSinkStruct
		addApp.Index = strings.SplitN(sink, "\n", 2)[0]
		addApp.Name = GetStringInBetween(sink, "application.name = \"", "\"") // Only for flatpak, not for binary: GetStringInBetween(sink, "pipewire.access.portal.app_id = \"", "\"")
		addApp.Description = "App: " + GetStringInBetween(sink, "application.name = \"", "\"")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			log.Println("Error in getSinkInputs, while converting Volume String to int: " + err.Error())
		}
		addApp.devType = DEV_SINK_INPUT

		apps = append(apps, addApp)
	}

	return apps
}

// from https://stackoverflow.com/a/42331558
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	e := strings.Index(str[s:], end)
	if e == -1 {
		return
	}
	return str[s : s+e]
}
