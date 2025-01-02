package resources

import (
	_ "embed"
	"fyne.io/fyne/v2"
)

//go:embed icon.png
var IconBytes []byte
var IconResource = fyne.NewStaticResource("icon.png", IconBytes)
