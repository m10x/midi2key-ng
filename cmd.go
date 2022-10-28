package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

const SINK = 0
const SOURCE = 1
const SINK_INPUT = 2

type applicationSinkStruct struct {
	index       string // Sink Input #
	name        string // sinkinput: sink: Name
	description string // sinkinput: application.name, sink: Description
	mute        bool   // Mute
	volume      int    // Volume (in %)
	typ         int    // Sink, Source or Sinkinput?
}

// https://stackoverflow.com/a/20438245
func exeCmd(cmd string) ([]byte, error) {
	fmt.Printf("command is %s\n", cmd)
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

func getSinks() []applicationSinkStruct {
	out, err := exeCmd("pactl list sinks")
	if err != nil {
		fmt.Println("Error in getSinks, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Sink #")

	var apps []applicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp applicationSinkStruct
		addApp.index = strings.SplitN(sink, "\n", 2)[0]
		addApp.name = GetStringInBetween(sink, "Name: ", "\n")
		addApp.description = "Out: " + GetStringInBetween(sink, "Description: ", "\n")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			fmt.Println("Error in getSinks, while converting Volume String to int: " + err.Error())
		}
		addApp.typ = SINK

		apps = append(apps, addApp)
	}

	return apps
}

func getSources() []applicationSinkStruct {
	out, err := exeCmd("pactl list sources")
	if err != nil {
		fmt.Println("Error in getSources, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Source #")

	var apps []applicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp applicationSinkStruct
		addApp.index = strings.SplitN(sink, "\n", 2)[0]
		addApp.name = GetStringInBetween(sink, "Name: ", "\n")
		addApp.description = "In: " + GetStringInBetween(sink, "Description: ", "\n")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			fmt.Println("Error in getSources, while converting Volume String to int: " + err.Error())
		}
		addApp.typ = SOURCE

		apps = append(apps, addApp)
	}

	return apps
}

func getSinkInputs() []applicationSinkStruct {
	out, err := exeCmd("pactl list sink-inputs")
	if err != nil {
		fmt.Println("Error in getSinkInputs, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Sink Input #")

	var apps []applicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp applicationSinkStruct
		addApp.index = strings.SplitN(sink, "\n", 2)[0]
		addApp.name = GetStringInBetween(sink, "application.name = \"", "\"") // Only for flatpak, not for binary: GetStringInBetween(sink, "pipewire.access.portal.app_id = \"", "\"")
		addApp.description = "App: " + GetStringInBetween(sink, "application.name = \"", "\"")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.volume = sinkVolume
		} else {
			fmt.Println("Error in getSinkInputs, while converting Volume String to int: " + err.Error())
		}
		addApp.typ = SINK_INPUT

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
