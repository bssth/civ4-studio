package main

import (
	"embed"
	"github.com/bssth/civ4-studio/editor"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	editor.SetAssetFS(assets)
	editor.RunApplication()
}
