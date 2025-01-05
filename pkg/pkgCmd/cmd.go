package pkgCmd

import (
	"fmt"
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

func Prepare() {
	log.Println("Initiating preparations")
	// dotool
	dotoold, err := IsProgramRunning("dotoold")
	if err != nil {
		log.Println("Error checking if dotoold is running: " + err.Error())
	}
	if dotoold {
		log.Println("dotoold is already running")
	} else {
		log.Println("dotoold is not running. Starting it in the background.")
		output, err := ExeCmd("cat /etc/default/keyboard")
		keyboardLayout := "us"
		if err != nil {
			log.Println("Error reading keyboard layout, defaulting to 'us': " + err.Error())
		} else {
			keyboardLayout = pkgUtils.GetStringInBetween(output, `XKBLAYOUT="`, `"`)
			log.Println("Detected keyboard layout " + keyboardLayout)
		}
		err = ExeCmdBackground("dotoold", []string{}, []string{"DOTOOL_XKB_LAYOUT=" + keyboardLayout})
		if err != nil {
			log.Println("Error starting dotoold: " + err.Error())
		} else {
			log.Println("Started dotoold")
		}
	}

	// pactl
	output, err := ExeCmd("pactl list sources")
	if err != nil {
		log.Println("Error listing audio sources: " + err.Error())
	}
	if strings.Contains(output, "Name: soundboard_mic") {
		log.Println("soundboard_mic is already available")
	} else {
		log.Println("soundboard_mic is not available yet. Creating it with the default audio source and sink.")
		defaultSource, err := ExeCmd("pactl get-default-source")
		if err != nil {
			log.Println("Error getting default audio source: " + err.Error())
		}
		log.Println("The default source is " + defaultSource)
		defaultSink, err := ExeCmd("pactl get-default-sink")
		if err != nil {
			log.Println("Error getting default audio sink: " + err.Error())
		}
		log.Println("The default sink is " + defaultSink)
		// Create new sink
		_, err = ExeCmd("pactl load-module module-null-sink sink_name=soundboard_mix sink_properties=device.description=SoundboardMix")
		if err != nil {
			log.Println("Error creating new sink: " + err.Error())
		}
		// Combine new sink with default sink
		_, err = ExeCmd(`pactl load-module module-combine-sink sink_name=soundboard_router slaves=` + defaultSink + `,soundboard_mix sink_properties=device.description="SoundboardRouter"`)
		if err != nil {
			log.Println("Error combining sinks: " + err.Error())
		}
		// Loopback to default source
		_, err = ExeCmd("pactl load-module module-loopback sink=soundboard_mix source=" + defaultSource)
		if err != nil {
			log.Println("Error loopbacking: " + err.Error())
		}
		// Create new source
		_, err = ExeCmd(`pactl load-module module-remap-source master=soundboard_mix.monitor source_name=soundboard_mic source_properties=device.description="soundboard_mic"`)
		if err != nil {
			log.Println("Error creating new source: " + err.Error())
		}
		log.Println("Created new source soundboard_mic")
	}

	log.Println("Finished preparations")
}

// https://stackoverflow.com/a/20438245
func ExeCmd(cmd string) (string, error) {
	log.Printf("command is %s\n", cmd)

	// Create the command
	command := exec.Command("sh", "-c", cmd)

	// Set LC_ALL to enforce English output
	command.Env = append(os.Environ(), "LC_ALL=C")

	// Execute the command and capture output
	out, err := command.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}

func ExeCmdRoutine(cmd string) {

	go func() {
		// ExeCmd aufrufen
		out, err := ExeCmd(cmd)
		if err != nil {
			log.Println("Error running command " + cmd + " in go routine: " + err.Error())
			return
		}
		if out != "" {
			log.Println("Command " + cmd + " returned " + out)
		}
	}()

}

// ExeCmdBackground starts a program in the background using nohup with optional environment variables.
func ExeCmdBackground(program string, args []string, envVars []string) error {
	// Combine the program and arguments into a single command
	cmd := exec.Command("nohup", append([]string{program}, args...)...)

	// Set environment variables
	cmd.Env = append(os.Environ(), "LC_ALL=C") // Set LC_ALL for consistent behavior
	cmd.Env = append(cmd.Env, envVars...)

	// Redirect stdout and stderr to avoid terminal dependency
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Detach the process (let it run in the background)
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start %s: %w", program, err)
	}

	// Program started successfully
	log.Printf("Started %s in the background with PID %d\n", program, cmd.Process.Pid)
	return nil
}

func IsProgramRunning(programName string) (bool, error) {
	// Use pgrep to check if the program is running
	cmd := fmt.Sprintf("pgrep -x %s", programName)
	output, err := ExeCmd(cmd)

	// If pgrep finds no match, it returns an error with no output
	if err != nil {
		// Check if the error is because no processes were found
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			// No process with the given name was found
			return false, nil
		}
		// If there's a different error, return it
		return false, err
	}

	// If there's output, the program is running
	if output != "" {
		return true, nil
	}

	return false, nil
}

// GetFocusedWindowPID retrieves the PID of the currently focused window using gdbus.
func GetFocusedWindowPID() (string, error) {
	cmd := `gdbus call --session --dest org.gnome.Shell --object-path /org/gnome/Shell/Extensions/WindowsExt --method org.gnome.Shell.Extensions.WindowsExt.FocusPID`
	output, err := ExeCmd(cmd)
	if err != nil {
		return "", err
	}

	// Extract the window PID from the gdbus output
	pid := pkgUtils.GetStringInBetween(output, `('`, `',)`)
	return pid, nil
}

func GetPulseAudioSinkDescriptionByPID(pid string) (string, error) {
	cmd := `pactl list sink-inputs`
	output, err := ExeCmd(cmd)
	if err != nil {
		return "", err
	}

	lines := strings.Split(output, "\n")
	var description string
	var foundPID bool
	var startOfProperties int

	// First, find the line with the application.process.id (PID)
	for i, line := range lines {
		// When we find a line containing the matching PID (note that the PID is in quotes)
		if strings.Contains(line, "application.process.id") && strings.Contains(line, fmt.Sprintf("\"%s\"", pid)) {
			// Flag that we've found the PID
			foundPID = true
			// Now, we need to go upwards to the "Properties:" section
			for j := i; j >= 0; j-- {
				if strings.Contains(lines[j], "Properties:") {
					startOfProperties = j
					break
				}
			}
			break
		}
	}

	// If no PID was found, return an error
	if !foundPID {
		return "", fmt.Errorf("no PulseAudio sink description found for PID '%s'", pid)
	}

	// Now, go down from "Properties:" to find the application.name
	for i := startOfProperties; i < len(lines); i++ {
		if strings.Contains(lines[i], "application.name") {
			// Ensure the line contains '=' before attempting to split
			if strings.Contains(lines[i], "=") {
				// Extract the application name (description)
				parts := strings.SplitN(lines[i], "=", 2)
				if len(parts) > 1 {
					// Strip quotes and whitespace from the description
					description = strings.Trim(strings.TrimSpace(parts[1]), "\"")
					break
				}
			}
		}
	}

	// If no description was found, return an error
	if description == "" {
		return "", fmt.Errorf("no application.name found for PID '%s'", pid)
	}

	return description, nil
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
