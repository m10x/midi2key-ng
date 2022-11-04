# midi2key-ng
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m10x/midi2key-ng)](https://golang.org/)
[![Fyne.io](https://img.shields.io/badge/Fyne-v2-blue)](https://fyne.io/)
[![Gomidi](https://img.shields.io/badge/Gomidi-v2-blue)](https://gitlab.com/gomidi/midi/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## About

Map Buttons, Knobs and Sliders of your Midi Controller to Different Functions. With GUI. Crossplatform for Linux, Windows and MacOS.

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
The repository can be fetched and installed using Go.  
`go install -v github.com/m10x/midi2key-ng@latest`  
  
TODO: Release Precompiled Binaries

## Roadmap
- improve handling of knobs and sliders (turning left/sliding down = decrease, turning right/sliding up = increase)
- code cleanup
- release precompiled binaries
- export / import Key Mapping
- add info button to systray
- add optional textbox with log output
- improve design, layout etc.
- add Windows Audio Control
- add MacOS Audio Control

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
