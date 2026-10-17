package resources

import (
	_ "embed"
)

// IconBytes is the application icon, 256×256. Linux shows it for the window.
//
//go:embed icon.png
var IconBytes []byte
