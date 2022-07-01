package main

import (
	"embed"
	"os"
	"runtime"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
	"github.com/wailsapp/wails/v2/pkg/options"
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := NewApp()

	// Create menu
	AppMenu := menu.NewMenu()
	FileMenu := AppMenu.AddSubmenu("File")
	//FileMenu.AddText("&Open", keys.CmdOrCtrl("o"), func(_ *menu.CallbackData))
	FileMenu.AddSeparator()
	FileMenu.AddText("Quit", keys.CmdOrCtrl("q"), func(_ *menu.CallbackData) {
		os.Exit(0)
	})
	AboutMenu := AppMenu.AddSubmenu("About")
	AboutMenu.AddText("About", keys.CmdOrCtrl("a"), func(_ *menu.CallbackData) {

	})

	if runtime.GOOS == "darwin" {
		AppMenu.Append(menu.EditMenu()) // on macos platform, we should append EditMenu to enable Cmd+C,Cmd+V,Cmd+Z... shortcut
	}

	// Create application with options
	err := wails.Run(&options.App{
		Title:            "midi2key-ng",
		Width:            1024,
		Height:           768,
		Assets:           assets,
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Menu:             AppMenu,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err)
	}
}
