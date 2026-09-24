package editor

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	EventConsole     = "console"
	EventXmlProgress = "xml-progress"
	EventXmlDone     = "xml-done"
	EventXmlReset    = "xml-reset"
	EventMapLoaded   = "map-loaded"
	EventMapState    = "map-state"
)

const (
	wbSaveExt        = ".CivBeyondSwordWBSave"
	autoSaveInterval = 5 * time.Minute
	backupExt        = ".bak"
)

// testFilePath is referenced from test files; kept here so it survives even when
// the dev-only "load test map" code path is removed.
const testFilePath = "resources/test.CivBeyondSwordWBSave"

var wbSaveFilters = []runtime.FileFilter{
	{
		DisplayName: "Civilization IV WorldBuilder Save (*" + wbSaveExt + ")",
		Pattern:     "*" + wbSaveExt,
	},
	{DisplayName: "All files", Pattern: "*.*"},
}

// App is the root struct bound to the Wails frontend.
// It holds the currently open map and all configuration.
type App struct {
	ctx context.Context

	mu       sync.Mutex
	filePath string
	wbMap    *WbMap
	xmlReady bool
	// dirty is set when the map has changes that are not saved yet
	dirty bool
	// backedUp keeps files that were already copied to .bak in this session
	backedUp map[string]bool
}

func NewApp() *App {
	return &App{backedUp: make(map[string]bool)}
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

	// A map passed on the command line (e.g. opened through the file association)
	for _, arg := range flag.Args() {
		if strings.EqualFold(filepath.Ext(arg), wbSaveExt) {
			if err := a.OpenMap(arg); err != nil {
				ConsoleWrite("Cannot open %s: %s", arg, err.Error())
			}
			break
		}
	}

	go a.autoSaveLoop(ctx)
}

// beforeClose asks the user to confirm quitting when there are unsaved changes.
// Returning true prevents the window from closing.
func (a *App) beforeClose(ctx context.Context) bool {
	return !a.confirmDiscard()
}

// confirmDiscard returns true if there are no unsaved changes or the user agreed to lose them
func (a *App) confirmDiscard() bool {
	a.mu.Lock()
	dirty := a.dirty
	a.mu.Unlock()
	if !dirty || a.ctx == nil {
		return true
	}

	answer, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Unsaved changes",
		Message:       "The map has unsaved changes. Discard them?",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		ConsoleWrite(err.Error())
		return false
	}
	return strings.EqualFold(answer, "Yes") || strings.EqualFold(answer, "Ok")
}

func (a *App) autoSaveLoop(ctx context.Context) {
	ticker := time.NewTicker(autoSaveInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.autoSave()
		}
	}
}

func (a *App) autoSave() {
	if GlobalConfig == nil || !GlobalConfig.AutoSave {
		return
	}
	a.mu.Lock()
	path, dirty := a.filePath, a.dirty
	a.mu.Unlock()
	// A new map has no file yet, it is saved only when the user chooses the location
	if !dirty || path == "" {
		return
	}
	if _, err := a.SaveMap(path); err != nil {
		ConsoleWrite("Autosave failed: %s", err.Error())
		return
	}
	ConsoleWrite("Autosaved %s", path)
}

// emit sends an event to the frontend. It is a no-op until the app is started (e.g. in tests)
func (a *App) emit(event string, data ...interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, event, data...)
	}
}

// setDirty updates the unsaved changes flag and notifies the frontend when it changes
func (a *App) setDirty(dirty bool) {
	a.mu.Lock()
	changed := a.dirty != dirty
	a.dirty = dirty
	a.mu.Unlock()
	if changed {
		a.emit(EventMapState, dirty)
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
// When the game directory or the mod changes, loaded game XML is dropped and the
// frontend is asked to load it again.
func (a *App) SetConfig(config *Config) {
	if config == nil {
		return
	}
	old := GlobalConfig
	GlobalConfig = config
	if err := SaveConfig(); err != nil {
		ConsoleWrite(err.Error())
	}

	if old == nil || old.GameDir != config.GameDir || old.Mod != config.Mod {
		a.ResetGameXML()
	}
}

// ChooseGameDir shows a directory dialog and returns the chosen path ("" if cancelled).
func (a *App) ChooseGameDir() (string, error) {
	current := ""
	if GlobalConfig != nil {
		current = GlobalConfig.GameDir
	}
	return runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Choose Beyond the Sword directory",
		DefaultDirectory: current,
	})
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

// LoadGameXML scans the configured game (and mod) directories and loads texts and
// info tables (civilizations, techs, eras etc.). Emits xml-progress events while running.
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

	data, err := LoadAllXML(func(s string) {
		a.emit(EventXmlProgress, s)
	})
	if err != nil {
		return err
	}
	SetGameData(data)

	a.mu.Lock()
	a.xmlReady = true
	a.mu.Unlock()

	a.emit(EventXmlDone)
	return nil
}

// ResetGameXML drops loaded game data, so the next LoadGameXML call reads files again.
func (a *App) ResetGameXML() {
	a.mu.Lock()
	a.xmlReady = false
	a.mu.Unlock()
	SetGameData(NewGameData())
	a.emit(EventXmlReset)
}

// XmlReady indicates whether game data has been loaded.
func (a *App) XmlReady() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.xmlReady
}

// OpenMapDialog shows a file open dialog and loads the selected map.
// Returns the chosen file path on success, "" if the user cancelled.
func (a *App) OpenMapDialog() (string, error) {
	if !a.confirmDiscard() {
		return "", nil
	}

	defaultDir := ""
	if GlobalConfig != nil && GlobalConfig.GameDir != "" {
		defaultDir = GetPublicMapsDir()
	}
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:            "Open WorldBuilder save",
		DefaultDirectory: defaultDir,
		Filters:          wbSaveFilters,
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
	a.setDirty(false)

	a.emit(EventMapLoaded, path)
	return nil
}

// NewMap discards the current map and creates an empty one.
// Returns false if the user decided to keep unsaved changes.
func (a *App) NewMap() bool {
	if !a.confirmDiscard() {
		return false
	}

	a.mu.Lock()
	a.wbMap = &WbMap{
		Version: defaultVersion,
		Game:    &Game{StartYear: -4000},
		Map:     &MapProps{TopLatitude: 90, BottomLatitude: -90, WrapX: 1},
	}
	a.filePath = ""
	a.mu.Unlock()
	// A new map is not saved anywhere yet
	a.setDirty(true)

	a.emit(EventMapLoaded, "")
	return true
}

// SaveMap writes the current map to disk. If path is empty, the previously
// opened path is used; if that is also empty, a save dialog is shown.
// Returns the path the map was saved to, "" if the user cancelled.
func (a *App) SaveMap(path string) (string, error) {
	if path == "" {
		path = a.CurrentMapPath()
	}
	if path == "" {
		return a.SaveMapAs()
	}
	return a.saveTo(path)
}

// SaveMapAs always asks for a new file name.
func (a *App) SaveMapAs() (string, error) {
	a.mu.Lock()
	current := a.filePath
	a.mu.Unlock()

	defaultDir, defaultName := "", "map"+wbSaveExt
	if current != "" {
		defaultDir, defaultName = filepath.Dir(current), filepath.Base(current)
	} else if GlobalConfig != nil && GlobalConfig.GameDir != "" {
		defaultDir = GetPublicMapsDir()
	}

	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Save WorldBuilder save",
		DefaultDirectory: defaultDir,
		DefaultFilename:  defaultName,
		Filters:          wbSaveFilters[:1],
	})
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += wbSaveExt
	}
	return a.saveTo(path)
}

func (a *App) saveTo(path string) (string, error) {
	a.mu.Lock()
	wb := a.wbMap
	var data []byte
	if wb != nil && wb.Game != nil {
		data = wb.ToWbFormat()
	}
	a.mu.Unlock()

	if data == nil {
		return "", errors.New("no map loaded")
	}

	if err := a.backupOnce(path); err != nil {
		ConsoleWrite("Cannot create backup: %s", err.Error())
		return "", err
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		ConsoleWrite(err.Error())
		return "", err
	}

	a.mu.Lock()
	a.filePath = path
	a.mu.Unlock()
	a.setDirty(false)
	ConsoleWrite("Saved %s", path)
	return path, nil
}

// backupOnce copies an existing file to <file>.bak before it is overwritten for the first time in this session
func (a *App) backupOnce(path string) error {
	a.mu.Lock()
	done := a.backedUp[path]
	a.mu.Unlock()
	if done {
		return nil
	}

	src, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(path + backupExt)
	if err != nil {
		return err
	}
	if _, err := io.Copy(dst, src); err != nil {
		dst.Close()
		return err
	}
	if err := dst.Close(); err != nil {
		return err
	}

	a.mu.Lock()
	a.backedUp[path] = true
	a.mu.Unlock()
	ConsoleWrite("Original file is backed up to %s", path+backupExt)
	return nil
}

// LaunchGame starts the game with the currently configured mod and map.
func (a *App) LaunchGame() error {
	return LaunchGame(a.CurrentMapPath())
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
	Dirty       bool   `json:"dirty"`
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
		Dirty:       a.dirty,
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
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	changed := a.wbMap.Game == nil || !bytes.Equal(a.wbMap.Game.ToWbFormat(), g.ToWbFormat())
	a.wbMap.Game = g
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
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
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	changed := !sameWbFormat(a.wbMap.Teams, teams)
	a.wbMap.Teams = teams
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
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
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	changed := !sameWbFormat(a.wbMap.Players, players)
	a.wbMap.Players = players
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
	return nil
}

// sameWbFormat compares sections by their saved form, so e.g. a nil and an empty list are equal
func sameWbFormat[T WbStructPackable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !bytes.Equal(a[i].ToWbFormat(), b[i].ToWbFormat()) {
			return false
		}
	}
	return true
}

// EnumOption pairs a raw type identifier with its localized description.
type EnumOption struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	// Group is a parent type (era of a tech, civic option of a civic), may be empty
	Group string `json:"group,omitempty"`
}

// CivilizationOption describes a civilization with the values used to fill in a player
type CivilizationOption struct {
	Type             string   `json:"type"`
	Description      string   `json:"description"`
	ShortDescription string   `json:"short_description"`
	Adjective        string   `json:"adjective"`
	Color            string   `json:"color"`
	ArtStyle         string   `json:"art_style"`
	Leaders          []string `json:"leaders"`
	Playable         bool     `json:"playable"`
}

// optionKeys maps frontend names of option lists to info categories
var optionKeys = map[string]string{
	"eras":          InfoEras,
	"speeds":        InfoSpeeds,
	"calendars":     InfoCalendars,
	"victories":     InfoVictories,
	"gameOptions":   InfoGameOptions,
	"mpOptions":     InfoMPOptions,
	"forceControls": InfoForceControls,
	"leaders":       InfoLeaders,
	"handicaps":     InfoHandicaps,
	"colors":        InfoPlayerColors,
	"artStyles":     InfoArtStyles,
	"techs":         InfoTechs,
	"religions":     InfoReligions,
	"civics":        InfoCivics,
	"civicOptions":  InfoCivicOptions,
	"projects":      InfoProjects,
	"worldSizes":    InfoWorldSizes,
	"climates":      InfoClimates,
	"seaLevels":     InfoSeaLevels,
}

func describe(data *GameData, info *TypeInfo) string {
	if info.Description == "" {
		return HumanizeType(info.Type)
	}
	return data.Text(info.Description)
}

// GetOptions returns all option lists loaded from game XML, keyed by list name (eras, techs, leaders...).
// Entries are in the order they are defined in the game files.
func (a *App) GetOptions() map[string][]EnumOption {
	data := CurrentGameData()
	result := make(map[string][]EnumOption, len(optionKeys))
	for key, category := range optionKeys {
		infos := data.Table(category).All()
		options := make([]EnumOption, 0, len(infos))
		for _, info := range infos {
			options = append(options, EnumOption{
				Type:        info.Type,
				Description: describe(data, info),
				Group:       info.Group,
			})
		}
		result[key] = options
	}
	return result
}

// GetCivilizations returns civilizations with their default names, colors and leaders.
func (a *App) GetCivilizations() []CivilizationOption {
	data := CurrentGameData()
	infos := data.Table(InfoCivilizations).All()
	result := make([]CivilizationOption, 0, len(infos))
	for _, info := range infos {
		leaders := info.Leaders
		if leaders == nil {
			leaders = []string{}
		}
		result = append(result, CivilizationOption{
			Type:             info.Type,
			Description:      describe(data, info),
			ShortDescription: data.Text(info.ShortDescription),
			Adjective:        data.Text(info.Adjective),
			Color:            info.DefaultPlayerColor,
			ArtStyle:         info.ArtStyleType,
			Leaders:          leaders,
			Playable:         info.Playable,
		})
	}
	return result
}

// TranslateKey looks up a localized string for a XML text key.
func (a *App) TranslateKey(key string) string {
	return GetLangString(key)
}
