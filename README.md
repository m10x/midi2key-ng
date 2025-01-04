# midi2key-ng
[![Release](https://img.shields.io/github/release/m10x/midi2key-ng.svg?color=brightgreen)](https://github.com/m10x/midi2key-ng/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/m10x/midi2key-ng)](https://goreportcard.com/report/github.com/m10x/midi2key-ng)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m10x/midi2key-ng)](https://golang.org/)
[![Fyne.io](https://img.shields.io/badge/Fyne-v2-blue)](https://fyne.io/)
[![Gomidi](https://img.shields.io/badge/Gomidi-v2-blue)](https://gitlab.com/gomidi/midi/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## About

Map Buttons, Knobs and Sliders of your Midi Controller to Different Functions. With GUI. Developed for Linux (x11 & wayland) and a Behringer X Touch Mini.

SOOOUUUUND



➜  ~ pactl unload-module module-remap-source
➜  ~ pactl unload-module module-loopback
➜  ~ pactl unload-module module-combine-sink
➜  ~ pactl unload-module module-null-sink

pactl list sources `Name: soundboard_mic`

pactl load-module module-null-sink sink_name=soundboard_mix sink_properties=device.description=SoundboardMix
pactl get-default-sink
pactl load-module module-combine-sink sink_name=soundboard_router slaves=alsa_output.usb-Audient_Audient_iD4-00.analog-surround-40,soundboard_mix sink_properties=device.description="SoundboardRouter"
pactl get-default-source
pactl load-module module-loopback sink=soundboard_mix source=alsa_input.usb-Audient_Audient_iD4-00.multichannel-input
pactl load-module module-remap-source master=soundboard_mix.monitor source_name=soundboard_mic source_properties=device.description="soundboard_mic"

pactl set-default-source soundboard_mic

paplay --device=soundboard_router -p Downloads/Short\ Song\ \(English\ Song\)🎵\ \[W\ Lyrics\]\ 30\ seconds.wav --volume=32000



## Features
Give your midicontroller the ability to:
- emulate key presses, mouse clicks/movements
  - Look [here](https://git.sr.ht/~geb/dotool/tree/master/doc/dotool.1.scd#L62) for possible input emulations
- write text
- run console commands
- soundboard
  - play audio files (e.g. wav, flac, ogg) as microphone input
  - new source soundboard_mic combines the default microphone with a new audio sink soundboard_router by utilizing pactl
  - run `paplay --list-file-formats` to list all available formats
- control your audio
  - input/output devices, applications, focused application (Currently only Gnome)
  - increase/decrease/set volume
  - mute/unmute/toggle

## Screenshots
Overview of Assignments
![image](https://user-images.githubusercontent.com/4344935/199974889-86d36ddc-32c7-48cc-b986-65a83aa575a3.png)

New Assignment
![image](https://user-images.githubusercontent.com/4344935/199975309-8205d9cf-65dd-4c01-b717-c5ccb2826150.png)

Edit an Assignment
![image](https://user-images.githubusercontent.com/4344935/199975097-e79b21e4-bd12-433b-9003-53939384a237.png)

## How to Install

### Option 1: Download precompiled binary
Download a precompiled binary from the [latest Release](https://github.com/m10x/midi2key-ng/releases).  

### Option 2: Install using go
The repository can be fetched and installed using Go.  
`go install -v github.com/m10x/midi2key-ng@latest`

### Requirements
- Install [DoTool](https://sr.ht/~geb/dotool/) for input emulation
    - `git clone https://git.sr.ht/\~geb/dotool` 
    - `sudo apt install scdoc`
    - `cd dotool && ./build.sh && sudo ./build.sh install`
    - `sudo udevadm control --reload && sudo udevadm trigger`
- Install Gnome Extension [Window Calls Extended](https://github.com/hseliger/window-calls-extended) to control audio of focused application
  
## Roadmap
- soundfile picker
- sort table
- spam actions if key keeps getting pressed (hold)
- reorder rows
- multiple profiles
- hotkeys to start/stop listening
- add optional textbox with log output
- add code comments
- create default Key Mapping for Behringer X Touch Mini with an easy option to add more defaults
- export / import Key Mapping
- improve design, layout etc.
- test other midi controllers

## Credits

### Frontend Framework:  
**fyne**  
https://fyne.io/

### MIDI Library:
**gomidi**  
https://gitlab.com/gomidi/midi/ 
https://pkg.go.dev/gitlab.com/gomidi/midi/v2

### Input Emulation:
**dotool**
https://sr.ht/~geb/dotool/