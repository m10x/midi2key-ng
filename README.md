# midi2key-ng
[![Release](https://img.shields.io/github/release/m10x/midi2key-ng.svg?color=brightgreen)](https://github.com/m10x/midi2key-ng/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/m10x/midi2key-ng)](https://goreportcard.com/report/github.com/m10x/midi2key-ng)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m10x/midi2key-ng)](https://golang.org/)
[![Fyne.io](https://img.shields.io/badge/Fyne-v2-blue)](https://fyne.io/)
[![Gomidi](https://img.shields.io/badge/Gomidi-v2-blue)](https://gitlab.com/gomidi/midi/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## About

Map Buttons, Knobs and Sliders of your Midi Controller to Different Functions. With GUI. Developed for Linux and a Behringer X Touch Mini, but most features are also working on Windows/MacOS and with other Midi Controllers.

## Features
Give your midicontroller the ability to:
- emulate key presses or combos
- write text
- run console commands
- control your audio (currently only for linux)
  - mute/unmute/toggle your input/output devices or even specific applications
  - increase/decrease/set volume of your input/output devices or even specific applications

## Screenshots
Overview of Assignments
![image](https://user-images.githubusercontent.com/4344935/199974889-86d36ddc-32c7-48cc-b986-65a83aa575a3.png)

New Assignment
![image](https://user-images.githubusercontent.com/4344935/199975309-8205d9cf-65dd-4c01-b717-c5ccb2826150.png)

Edit an Assignment
![image](https://user-images.githubusercontent.com/4344935/199975097-e79b21e4-bd12-433b-9003-53939384a237.png)

## How to Install

### Option 1: Download precompiled binary (Linux, Windows, MacOS)
Download a precompiled binary from the [latest Release](https://github.com/m10x/midi2key-ng/releases).  

### Option 2: Fetch using go (Linux, Windows, MacOS)
The repository can be fetched and installed using Go.  
`go install -v github.com/m10x/midi2key-ng@latest`  
  
## Roadmap
- update fyne (versions higher than 2.4.1 are throwing errors)
- check why systray icon is missing after fyne has been updated
- build midi2key-ng 1.1.0 executables (cross compiling didn't work anymore using the script)
- warn if key is already assigend
- reorder rows
- multiple profiles
- hotkeys to start/stop listening
- implement soundboard functionality using [beep](https://github.com/faiface/beep)
- add mouse emulation functionalities
- add optional textbox with log output
- add code comments
- export / import Key Mapping
- improve design, layout etc.
- test other midi controllers
- add Windows Audio Control
- add MacOS Audio Control

## For Developers
- `sudo apt install libx11-dev xorg-dev libxtst-dev`
- `sudo apt-get install libasound2-dev`

## Credits

### Frontend Framework:  
**fyne**  
https://fyne.io/

### MIDI Library:
**gomidi**  
https://gitlab.com/gomidi/midi/ 
https://pkg.go.dev/gitlab.com/gomidi/midi/v2  

### Simulate Keyboard + Mouse
**robot-go**
https://github.com/go-vgo/robotgo
