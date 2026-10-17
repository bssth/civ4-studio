package editor

import (
	"embed"
	"flag"
	"log"

	"github.com/bssth/civ4-studio/resources"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
)

var assets embed.FS

var GlobalConfig *Config

const steamDefaultGameDir = "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Sid Meier's Civilization IV Beyond the Sword\\Beyond the Sword"

var (
	RunGame = flag.Bool("run-game", false, "Run game with current mod instead of editor")
	DevMode = flag.Bool("dev", false, "Run in dev mode (load test map, show debug logs etc.)")
)

func SetAssetFS(fs embed.FS) {
	assets = fs
}

func RunApplication() {
	flag.Parse()

	if *RunGame {
		_ = LaunchGame("")
		return
	}

	app := NewApp()

	err := wails.Run(&options.App{
		Title:  AppTitle(),
		Width:  1024,
		Height: 768,
		// A map editor needs the whole screen; the size above is used when the window is restored
		WindowStartState: options.Maximised,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		OnBeforeClose:    app.beforeClose,
		Frameless:        true,
		Bind: []interface{}{
			app,
		},
		// Windows and macOS take the icon from the build folder, Linux needs it for the window
		Linux: &linux.Options{Icon: resources.IconBytes},
	})

	if err != nil {
		log.Println("Error: ", err)
	}
}
