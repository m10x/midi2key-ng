package main

import (
	"context"
	"fmt"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	fmt.Println("app started")
}

func (a *App) shutdown(ctx context.Context) {
	closeDriver()
	fmt.Println("app shutdowned")
}

func (a *App) Initialize() string {
	fmt.Println("initialize")
	return initialize()
}

func (a *App) LoadDevices() string {
	fmt.Println("load devices")
	return getInputPorts()
}

func (a *App) Listen(listen bool, device string) string {
	if listen {
		fmt.Println("start listening " + device)
		return startListen(device)
	} else {
		fmt.Println("stop listening " + device)
		return stopListen()
	}
}
