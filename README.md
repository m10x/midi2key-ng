# midi2key-ng
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/m10x/midi2key-ng)](https://golang.org/)
[![Fyne.io](https://img.shields.io/badge/Fyne-v2-blue)](https://fyne.io/)
[![Gomidi](https://img.shields.io/badge/Gomidi-v2-blue)](https://gitlab.com/gomidi/midi/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://www.apache.org/licenses/LICENSE-2.0)

## About

Map Keys, Knobs and Sliders of your Midi Controller to Different Functions. With GUI. Crossplatform for Linux, Windows and MacOS.

## Features
Give your midicontroller the ability to:
- emulate key presses or combos
- write text
- run console commands
- control your audio (currently only for linux)
  - mute/unmute/toggle your input/output devices or even specific applications
  - increase/decrease/set volume of your input/output devices or even specific applications

## Roadmap
- implement support for knobs and sliders
- fix fyne preferences
- export / import Key Mapping
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
