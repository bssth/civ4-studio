package editor

import (
	"strings"
	"sync"
)

// Info categories. The value is the name of the XML element that holds the list of entries,
// e.g. <Civ4TechInfos><TechInfos><TechInfo>...</TechInfo></TechInfos></Civ4TechInfos>
const (
	InfoEras          = "EraInfos"
	InfoSpeeds        = "GameSpeedInfos"
	InfoCalendars     = "CalendarInfos"
	InfoVictories     = "VictoryInfos"
	InfoGameOptions   = "GameOptionInfos"
	InfoMPOptions     = "MPOptionInfos"
	InfoForceControls = "ForceControlInfos"
	InfoCivilizations = "CivilizationInfos"
	InfoLeaders       = "LeaderHeadInfos"
	InfoHandicaps     = "HandicapInfos"
	InfoPlayerColors  = "PlayerColorInfos"
	InfoArtStyles     = "ArtStyleTypes"
	InfoTechs         = "TechInfos"
	InfoReligions     = "ReligionInfos"
	InfoCivics        = "CivicInfos"
	InfoCivicOptions  = "CivicOptionInfos"
	InfoProjects      = "ProjectInfos"
	InfoWorldSizes    = "WorldInfos"
	InfoClimates      = "ClimateInfos"
	InfoSeaLevels     = "SeaLevelInfos"
)

// KnownInfoCategories lists all info categories loaded from game XML files
var KnownInfoCategories = []string{
	InfoEras, InfoSpeeds, InfoCalendars, InfoVictories, InfoGameOptions, InfoMPOptions, InfoForceControls,
	InfoCivilizations, InfoLeaders, InfoHandicaps, InfoPlayerColors, InfoArtStyles, InfoTechs, InfoReligions,
	InfoCivics, InfoCivicOptions, InfoProjects, InfoWorldSizes, InfoClimates, InfoSeaLevels,
}

// TypeInfo is a single entry of a game info XML file (a civilization, a tech, a leader etc.)
type TypeInfo struct {
	Type string
	// Description is a text key, use GetLangString to get a readable text
	Description string
	// Group is an optional parent type: era of a tech, civic option of a civic
	Group string
	// Civilization-only fields
	ShortDescription   string
	Adjective          string
	DefaultPlayerColor string
	ArtStyleType       string
	Leaders            []string
	Playable           bool
}

// InfoTable keeps entries of one category in the order they are defined in XML files
type InfoTable struct {
	order []string
	items map[string]*TypeInfo
}

func NewInfoTable() *InfoTable {
	return &InfoTable{items: make(map[string]*TypeInfo)}
}

// Set adds an entry or replaces an existing one keeping its position
func (t *InfoTable) Set(info *TypeInfo) {
	if _, ok := t.items[info.Type]; !ok {
		t.order = append(t.order, info.Type)
	}
	t.items[info.Type] = info
}

func (t *InfoTable) Get(infoType string) *TypeInfo {
	return t.items[infoType]
}

func (t *InfoTable) Len() int {
	return len(t.order)
}

// All returns entries in definition order
func (t *InfoTable) All() []*TypeInfo {
	result := make([]*TypeInfo, 0, len(t.order))
	for _, k := range t.order {
		result = append(result, t.items[k])
	}
	return result
}

// GameData holds everything loaded from game (and mod) XML files
type GameData struct {
	Lang          map[string]string
	Infos         map[string]*InfoTable
	IntDefines    map[string]int
	FloatDefines  map[string]float64
	StringDefines map[string]string
}

func NewGameData() *GameData {
	data := &GameData{
		Lang:          make(map[string]string),
		Infos:         make(map[string]*InfoTable),
		IntDefines:    make(map[string]int),
		FloatDefines:  make(map[string]float64),
		StringDefines: make(map[string]string),
	}
	for _, category := range KnownInfoCategories {
		data.Infos[category] = NewInfoTable()
	}
	return data
}

// Table returns entries of a category (never nil)
func (d *GameData) Table(category string) *InfoTable {
	if t, ok := d.Infos[category]; ok {
		return t
	}
	return NewInfoTable()
}

// Text returns language string by key. If it's not found, returns key itself
func (d *GameData) Text(key string) string {
	if key == "" {
		return ""
	}

	// Support keys with and without TXT_KEY_ prefix as automatic fallback
	for _, variant := range []string{key, "TXT_KEY_" + key} {
		if value, ok := d.Lang[variant]; ok {
			return value
		}
	}

	return key
}

var (
	gameDataMu sync.RWMutex
	gameData   = NewGameData()
)

// CurrentGameData returns the data loaded by the last LoadAllXML call
func CurrentGameData() *GameData {
	gameDataMu.RLock()
	defer gameDataMu.RUnlock()
	return gameData
}

// SetGameData replaces current game data (pass NewGameData() to reset it)
func SetGameData(data *GameData) {
	gameDataMu.Lock()
	defer gameDataMu.Unlock()
	gameData = data
}

// GetLangString returns language string by key. If it's not found, returns key itself
func GetLangString(key string) string {
	return CurrentGameData().Text(key)
}

// HumanizeType turns a type identifier into a readable name, e.g. PLAYERCOLOR_DARK_RED -> Dark Red.
// Used for entries that have no description text.
func HumanizeType(infoType string) string {
	name := infoType
	if i := strings.Index(name, "_"); i >= 0 && i < len(name)-1 {
		name = name[i+1:]
	}
	words := strings.Split(strings.ToLower(name), "_")
	for i, w := range words {
		if w != "" {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
