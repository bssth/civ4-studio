package editor

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventConsole     = "console"
	EventXmlProgress = "xml-progress"
	EventXmlDone     = "xml-done"
	EventMapLoaded   = "map-loaded"
)

// testFilePath is referenced from test files; kept here so it survives even when
// the dev-only "load test map" code path is removed.
const testFilePath = "resources/test.CivBeyondSwordWBSave"

// App is the root struct bound to the Wails frontend.
// It holds the currently open map and all configuration.
type App struct {
	ctx context.Context

	mu       sync.Mutex
	filePath string
	wbMap    *WbMap
	xmlReady bool
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	console := GetConsoleChannel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line := <-console:
				runtime.EventsEmit(ctx, EventConsole, line)
			}
		}
	}()

	config, fileExists := GetConfig()
	GlobalConfig = &config
	if !fileExists {
		ConsoleWrite("Config file not found, opening settings")
	} else {
		ConsoleWrite("Config file found, opening editor")
	}
}

// WriteConsole appends a line to the editor console.
func (a *App) WriteConsole(line string) {
	ConsoleWrite(line)
}

// GetConfig returns current config.
func (a *App) GetConfig() *Config {
	return GlobalConfig
}

// SetConfig persists the supplied config and replaces the current one.
func (a *App) SetConfig(config *Config) {
	GlobalConfig = config
	if err := SaveConfig(); err != nil {
		ConsoleWrite(err.Error())
	}
}

// GetModsList returns mod folders found under the configured game directory.
func (a *App) GetModsList() []string {
	if GlobalConfig == nil {
		return nil
	}
	return GetModsList(GlobalConfig.GameDir)
}

// CheckGameDir verifies the configured game directory contains the expected files.
// Empty string is returned if the directory looks valid.
func (a *App) CheckGameDir() string {
	if GlobalConfig == nil {
		return "no config"
	}
	if GlobalConfig.GameDir == "" {
		return "game directory is not configured"
	}
	if err := CheckGameDirectory(GlobalConfig.GameDir); err != nil {
		return err.Error()
	}
	return ""
}

// LoadGameXML scans the configured game (and mod) directories and populates
// in-memory enum tables (eras, speeds, victories, etc.). Emits xml-progress
// events while running.
func (a *App) LoadGameXML() error {
	a.mu.Lock()
	already := a.xmlReady
	a.mu.Unlock()
	if already {
		return nil
	}

	if msg := a.CheckGameDir(); msg != "" {
		return errors.New(msg)
	}

	err := LoadAllXML(func(s string) {
		runtime.EventsEmit(a.ctx, EventXmlProgress, s)
	})
	if err != nil {
		return err
	}

	a.mu.Lock()
	a.xmlReady = true
	a.mu.Unlock()

	runtime.EventsEmit(a.ctx, EventXmlDone)
	return nil
}

// XmlReady indicates whether enum tables have been populated.
func (a *App) XmlReady() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.xmlReady
}

// OpenMapDialog shows a file open dialog and loads the selected map.
// Returns the chosen file path on success, "" if the user cancelled.
func (a *App) OpenMapDialog() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Open WorldBuilder save",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "Civilization IV WorldBuilder Save (*.CivBeyondSwordWBSave)",
				Pattern:     "*.CivBeyondSwordWBSave",
			},
			{DisplayName: "All files", Pattern: "*.*"},
		},
	})
	if err != nil || path == "" {
		return "", err
	}
	if err := a.OpenMap(path); err != nil {
		return "", err
	}
	return path, nil
}

// OpenMap parses a WorldBuilder save from disk and replaces the current map.
func (a *App) OpenMap(path string) error {
	f, err := os.Open(path)
	if err != nil {
		ConsoleWrite(err.Error())
		return err
	}
	defer f.Close()

	wb, err := ParseWbMap(f)
	if err != nil {
		ConsoleWrite(err.Error())
		return err
	}

	a.mu.Lock()
	a.wbMap = wb
	a.filePath = path
	a.mu.Unlock()

	runtime.EventsEmit(a.ctx, EventMapLoaded, path)
	return nil
}

// NewMap discards the current map and creates an empty one.
func (a *App) NewMap() {
	a.mu.Lock()
	a.wbMap = &WbMap{
		Version: defaultVersion,
		Game:    &Game{StartYear: -4000},
		Map:     &MapProps{TopLatitude: 90, BottomLatitude: -90, WrapX: 1},
	}
	a.filePath = ""
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, EventMapLoaded, "")
}

// SaveMap writes the current map to disk. If path is empty, the previously
// opened path is used; if that is also empty, a save dialog is shown.
func (a *App) SaveMap(path string) (string, error) {
	a.mu.Lock()
	wb := a.wbMap
	current := a.filePath
	a.mu.Unlock()

	if wb == nil || wb.Game == nil {
		return "", errors.New("no map loaded")
	}

	if path == "" {
		path = current
	}
	if path == "" {
		dialogPath, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
			Title:           "Save WorldBuilder save",
			DefaultFilename: "map.CivBeyondSwordWBSave",
			Filters: []runtime.FileFilter{
				{
					DisplayName: "Civilization IV WorldBuilder Save (*.CivBeyondSwordWBSave)",
					Pattern:     "*.CivBeyondSwordWBSave",
				},
			},
		})
		if err != nil || dialogPath == "" {
			return "", err
		}
		path = dialogPath
	}

	if err := os.WriteFile(path, wb.ToWbFormat(), 0644); err != nil {
		ConsoleWrite(err.Error())
		return "", err
	}

	a.mu.Lock()
	a.filePath = path
	a.mu.Unlock()
	return path, nil
}

// LaunchGame starts the game with the currently configured mod and map.
func (a *App) LaunchGame() error {
	a.mu.Lock()
	path := a.filePath
	a.mu.Unlock()
	return LaunchGame(path)
}

// CurrentMapPath returns the path of the currently open file, or "".
func (a *App) CurrentMapPath() string {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.filePath
}

// HasMap reports whether a map is currently loaded.
func (a *App) HasMap() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.wbMap != nil
}

// MapInfo is a lightweight summary of the currently loaded map.
type MapInfo struct {
	Path        string `json:"path"`
	Version     int    `json:"version"`
	TeamsCount  int    `json:"teams_count"`
	PlayerCount int    `json:"player_count"`
	PlotCount   int    `json:"plot_count"`
}

// GetMapInfo returns a small summary of the loaded map (or nil).
func (a *App) GetMapInfo() *MapInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	return &MapInfo{
		Path:        a.filePath,
		Version:     a.wbMap.Version,
		TeamsCount:  len(a.wbMap.Teams),
		PlayerCount: len(a.wbMap.Players),
		PlotCount:   len(a.wbMap.Plots),
	}
}

// GetGame returns the current Game settings (nil if no map loaded).
func (a *App) GetGame() *Game {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	return a.wbMap.Game
}

// SetGame replaces the Game settings of the current map.
func (a *App) SetGame(g *Game) error {
	if g == nil {
		return errors.New("game is nil")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return errors.New("no map loaded")
	}
	a.wbMap.Game = g
	return nil
}

// GetTeams returns the current list of teams.
func (a *App) GetTeams() []*Team {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	return a.wbMap.Teams
}

// SetTeams replaces the teams list.
func (a *App) SetTeams(teams []*Team) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return errors.New("no map loaded")
	}
	a.wbMap.Teams = teams
	return nil
}

// GetPlayers returns the current list of players.
func (a *App) GetPlayers() []*Player {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	return a.wbMap.Players
}

// SetPlayers replaces the players list.
func (a *App) SetPlayers(players []*Player) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return errors.New("no map loaded")
	}
	a.wbMap.Players = players
	return nil
}

// EnumOption pairs a raw type identifier with its localized description.
type EnumOption struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

func enumOptions(infos map[string]*TypeInfo) []EnumOption {
	keys := SortKeys(infos)
	result := make([]EnumOption, 0, len(keys))
	for _, k := range keys {
		result = append(result, EnumOption{
			Type:        infos[k].Type,
			Description: GetLangString(infos[k].Description),
		})
	}
	return result
}

func (a *App) GetEraOptions() []EnumOption          { return enumOptions(EraInfos) }
func (a *App) GetSpeedOptions() []EnumOption        { return enumOptions(SpeedInfos) }
func (a *App) GetCalendarOptions() []EnumOption     { return enumOptions(CalendarInfos) }
func (a *App) GetVictoryOptions() []EnumOption      { return enumOptions(VictoryInfos) }
func (a *App) GetGameOptionOptions() []EnumOption   { return enumOptions(GameOptionInfos) }
func (a *App) GetMPOptionOptions() []EnumOption     { return enumOptions(GameMPInfos) }
func (a *App) GetForceControlOptions() []EnumOption { return enumOptions(ForceControlInfos) }

// TranslateKey looks up a localized string for a XML text key.
func (a *App) TranslateKey(key string) string {
	return GetLangString(key)
}
