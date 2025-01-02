package editor

import (
	"embed"
	"flag"
	"fyne.io/fyne/v2"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"log"
)

var assets embed.FS

var Application fyne.App
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
		Title:  "Civ4 Studio",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Frameless:        true,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Println("Error: ", err)
	}
}
