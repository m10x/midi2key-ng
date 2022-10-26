package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type applicationSinkStruct struct {
	index  string // Sink Input #
	appid  string // pipewire.access.portal.app_id
	name   string // application.name
	mute   bool   // Mute
	volume int    // Volume (in %)
}

// https://stackoverflow.com/a/20438245
func exeCmd(cmd string) ([]byte, error) {
	fmt.Println("command is ", cmd)
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

func getInputSinks() []applicationSinkStruct {
	out, err := exeCmd("pactl list sink-inputs")
	if err != nil {
		fmt.Println("Error in getInputSinks, while executing command: " + err.Error())
	}
	outStr := strings.Split(string(out), "Sink Input #")

	var apps []applicationSinkStruct

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		var addApp applicationSinkStruct
		addApp.index = strings.SplitN(sink, "\n", 2)[0]
		addApp.appid = GetStringInBetween(sink, "pipewire.access.portal.app_id = \"", "\"")
		addApp.name = GetStringInBetween(sink, "application.name = \"", "\"")
		addApp.mute = GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(GetStringInBetween(sinkVolumeStr, " / ", "%")); err == nil {
			addApp.volume = sinkVolume
		} else {
			fmt.Println("Error in getInputSinks, while converting Volume String to int: " + err.Error())
		}

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

func getOutputSinks() {
	out, err := exeCmd("pactl list sinks")
	if err != nil {
		fmt.Println("Error getOutputSinks" + err.Error())
	}
	outStr := strings.Split(string(out), "Sink #")

	var sinks = []string{}

	for i, sink := range outStr {
		if i == 0 {
			continue
		}
		number := strings.SplitN(sink, "\n", 2)[0]
		// Get Volume, Get Mute, Get application.Name, Struct erstellen welches die Werte enthält. Icon name?

		fmt.Println(number)
		sinks = append(sinks, number)
	}
	fmt.Println(sinks)
}
