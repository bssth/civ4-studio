package editor

import "strings"

// Version is set at build time from the release tag:
// wails build -ldflags "-X github.com/bssth/civ4-studio/editor.Version=v1.2.3"
var Version = "dev"

// AppTitle is the window title with the version, e.g. "Civ4 Studio v1.2.3"
func AppTitle() string {
	if Version == "" || Version == "dev" {
		return "Civ4 Studio (development build)"
	}
	return "Civ4 Studio " + Version
}

// GetVersion returns the application version ("dev" for local builds).
func (a *App) GetVersion() string {
	return strings.TrimSpace(Version)
}
