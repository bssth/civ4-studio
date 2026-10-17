package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMapsDir(t *testing.T) {
	game := t.TempDir()
	mod := filepath.Join(game, ModsDir, "TestMod")
	withConfig(t, Config{GameDir: game, Mod: "TestMod"})

	if got := GetMapsDir(); got != "" {
		t.Errorf("no maps folders: got %q", got)
	}

	writeTestFile(t, filepath.Join(game, PublicMapsDir, ".keep"), "")
	if got := GetMapsDir(); got != filepath.Join(game, PublicMapsDir) {
		t.Errorf("the game folder is expected when the mod has none, got %q", got)
	}

	writeTestFile(t, filepath.Join(mod, PrivateMapsDir, ".keep"), "")
	if got := GetMapsDir(); got != filepath.Join(mod, PrivateMapsDir) {
		t.Errorf("PrivateMaps of the mod is expected when it has no PublicMaps, got %q", got)
	}

	writeTestFile(t, filepath.Join(mod, PublicMapsDir, ".keep"), "")
	if got := GetMapsDir(); got != filepath.Join(mod, PublicMapsDir) {
		t.Errorf("PublicMaps of the mod goes first, got %q", got)
	}
}

func TestMapDialogDir(t *testing.T) {
	game := t.TempDir()
	writeTestFile(t, filepath.Join(game, PrivateMapsDir, ".keep"), "")
	withConfig(t, Config{GameDir: game})

	mapDir := filepath.Join(t.TempDir(), "maps")
	writeTestFile(t, filepath.Join(mapDir, "map.CivBeyondSwordWBSave"), "")
	current := filepath.Join(mapDir, "map.CivBeyondSwordWBSave")
	if got := mapDialogDir(current); got != mapDir {
		t.Errorf("folder of the open map expected, got %q", got)
	}

	if err := os.RemoveAll(mapDir); err != nil {
		t.Fatal(err)
	}
	if got := mapDialogDir(current); got != filepath.Join(game, PrivateMapsDir) {
		t.Errorf("a removed folder must not be used, got %q", got)
	}
	if got := mapDialogDir(""); got != filepath.Join(game, PrivateMapsDir) {
		t.Errorf("maps folder of the game expected for a new map, got %q", got)
	}
}
