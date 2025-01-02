package pkgCmd

import (
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/m10x/midi2key-ng/pkg/pkgUtils"
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

	Mute    bool // Mute
	Volume  int  // Volume (in %)
	devType int  // Sink, Source or Sinkinput? Ist das nötig? Hole es aktuell aus dem String
}

// https://stackoverflow.com/a/20438245
func ExeCmd(cmd string) (string, error) {
	log.Printf("command is %s\n", cmd)
	// splitting head => g++ parts => rest of the command
	parts := strings.Fields(cmd)
	head := parts[0]
	parts = parts[1:]

	// Create the command
	command := exec.Command(head, parts...)

	// Set LC_ALL to enforce English output
	command.Env = append(os.Environ(), "LC_ALL=C")

	// Execute the command and capture output
	out, err := command.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
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
		addApp.Name = pkgUtils.GetStringInBetween(sink, "Name: ", "\n")
		addApp.Description = "Out: " + pkgUtils.GetStringInBetween(sink, "Description: ", "\n")
		addApp.Mute = pkgUtils.GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := pkgUtils.GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(pkgUtils.GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.Volume = sinkVolume
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
		addApp.Name = pkgUtils.GetStringInBetween(sink, "Name: ", "\n")
		addApp.Description = "In: " + pkgUtils.GetStringInBetween(sink, "Description: ", "\n")
		addApp.Mute = pkgUtils.GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := pkgUtils.GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(pkgUtils.GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.Volume = sinkVolume
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
		addApp.Name = pkgUtils.GetStringInBetween(sink, "application.name = \"", "\"") // Only for flatpak, not for binary: GetStringInBetween(sink, "pipewire.access.portal.app_id = \"", "\"")
		addApp.Description = "App: " + pkgUtils.GetStringInBetween(sink, "application.name = \"", "\"")
		addApp.Mute = pkgUtils.GetStringInBetween(sink, "Mute: ", "\n") == "yes"
		sinkVolumeStr := pkgUtils.GetStringInBetween(sink, "Volume: ", "\n")
		if sinkVolume, err := strconv.Atoi(strings.TrimSpace(pkgUtils.GetStringInBetween(sinkVolumeStr, " / ", "%"))); err == nil {
			addApp.Volume = sinkVolume
		} else {
			log.Println("Error in getSinkInputs, while converting Volume String to int: " + err.Error())
		}
		addApp.devType = DEV_SINK_INPUT

		apps = append(apps, addApp)
	}

	return apps
}

func GetAppVolume(device string) int {
	for _, x := range GetSinkInputs() {
		if x.Description == device {
			return x.Volume
		}
	}
	return 0
}
