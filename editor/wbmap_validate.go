package editor

import (
	"fmt"
	"sort"
)

const (
	SeverityError   = "error"
	SeverityWarning = "warning"
)

// Problem is an issue found in the scenario. X/Y, Player and Team point to the place of the problem, -1 if not set.
type Problem struct {
	Severity string `json:"severity"`
	Section  string `json:"section"`
	Message  string `json:"message"`
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Player   int    `json:"player"`
	Team     int    `json:"team"`
}

type validator struct {
	m        *WbMap
	data     *GameData
	problems []Problem
	// unknown types are reported once per type with the number of occurrences
	unknown      map[string]*unknownType
	unknownOrder []string
}

type unknownType struct {
	problem Problem
	count   int
}

func (v *validator) add(severity, section, message string, x, y, player, team int) {
	v.problems = append(v.problems, Problem{Severity: severity, Section: section, Message: message, X: x, Y: y, Player: player, Team: team})
}

// checkType reports a type missing from game data. Categories without loaded data are not checked,
// so a map is not flooded with errors when game data is not available.
func (v *validator) checkType(section, category, what, value string, x, y, player, team int) {
	if value == "" || value == NonePlayer || v.data == nil {
		return
	}
	table := v.data.Table(category)
	if table.Len() == 0 || table.Get(value) != nil {
		return
	}
	key := section + "\x00" + what + "\x00" + value
	if u, ok := v.unknown[key]; ok {
		u.count++
		return
	}
	v.unknown[key] = &unknownType{
		problem: Problem{Severity: SeverityError, Section: section, X: x, Y: y, Player: player, Team: team,
			Message: fmt.Sprintf("unknown %s %s", what, value)},
		count: 1,
	}
	v.unknownOrder = append(v.unknownOrder, key)
}

func (m *WbMap) inGrid(x, y int) bool {
	return m.Map != nil && x >= 0 && y >= 0 && uint64(x) < m.Map.GridWidth && uint64(y) < m.Map.GridHeight
}

// Validate checks the scenario for problems that make the game fail or behave unexpectedly.
// data is the loaded game data, it may be empty (then types are not checked).
func (m *WbMap) Validate(data *GameData) []Problem {
	v := &validator{m: m, data: data, unknown: make(map[string]*unknownType)}
	v.game()
	v.mapSection()
	v.teams()
	v.players()
	v.plots()

	for _, key := range v.unknownOrder {
		u := v.unknown[key]
		if u.count > 1 {
			u.problem.Message += fmt.Sprintf(" (%d times)", u.count)
		}
		v.problems = append(v.problems, u.problem)
	}

	// Errors first, the rest keeps the order of sections
	sort.SliceStable(v.problems, func(i, j int) bool {
		return v.problems[i].Severity == SeverityError && v.problems[j].Severity != SeverityError
	})
	if v.problems == nil {
		return []Problem{}
	}
	return v.problems
}

func (v *validator) game() {
	g := v.m.Game
	if g == nil {
		v.add(SeverityError, "game", "the map has no BeginGame section", -1, -1, -1, -1)
		return
	}
	v.checkType("game", InfoEras, "era", g.Era, -1, -1, -1, -1)
	v.checkType("game", InfoSpeeds, "game speed", g.Speed, -1, -1, -1, -1)
	v.checkType("game", InfoCalendars, "calendar", g.Calendar, -1, -1, -1, -1)
	for _, victory := range g.Victory {
		v.checkType("game", InfoVictories, "victory", victory, -1, -1, -1, -1)
	}
	for _, option := range g.Option {
		v.checkType("game", InfoGameOptions, "game option", option, -1, -1, -1, -1)
	}
	for _, option := range g.MPOption {
		v.checkType("game", InfoMPOptions, "multiplayer option", option, -1, -1, -1, -1)
	}
	for _, option := range g.ForceControl {
		v.checkType("game", InfoForceControls, "force control", option, -1, -1, -1, -1)
	}
	if g.MaxTurns > 0 && g.MaxTurns <= g.GameTurn {
		v.add(SeverityError, "game", fmt.Sprintf("max turns (%d) must be greater than the starting turn (%d)", g.MaxTurns, g.GameTurn), -1, -1, -1, -1)
	}
}

func (v *validator) mapSection() {
	for _, problem := range v.m.Stats().Problems {
		v.add(SeverityError, "map", problem, -1, -1, -1, -1)
	}
	if v.m.Map != nil {
		v.checkType("map", InfoWorldSizes, "world size", v.m.Map.WorldSize, -1, -1, -1, -1)
		v.checkType("map", InfoClimates, "climate", v.m.Map.Climate, -1, -1, -1, -1)
		v.checkType("map", InfoSeaLevels, "sea level", v.m.Map.SeaLevel, -1, -1, -1, -1)
	}
}

func (v *validator) teams() {
	ids := make(map[uint]bool)
	for _, team := range v.m.Teams {
		if ids[team.TeamID] {
			v.add(SeverityError, "teams", fmt.Sprintf("team %d is defined more than once", team.TeamID), -1, -1, -1, int(team.TeamID))
		}
		ids[team.TeamID] = true
	}
	for _, team := range v.m.Teams {
		id := int(team.TeamID)
		for _, tech := range team.Tech {
			v.checkType("teams", InfoTechs, "tech", tech, -1, -1, -1, id)
		}
		for _, project := range team.ProjectType {
			v.checkType("teams", InfoProjects, "project", project, -1, -1, -1, id)
		}
		relations := map[string][]uint{
			"contact": team.ContactWithTeam, "war": team.AtWar, "open borders": team.OpenBordersWithTeam,
			"defensive pact": team.DefensivePactWithTeam, "permanent war/peace": team.PermanentWarPeace,
		}
		for _, name := range SortKeys(relations) {
			for _, other := range relations[name] {
				if !ids[other] {
					v.add(SeverityError, "teams", fmt.Sprintf("team %d has %s with team %d, which does not exist", id, name, other), -1, -1, -1, id)
				}
			}
		}
	}
}

func isEmptySlot(p *Player) bool {
	return (p.CivType == "" || p.CivType == NonePlayer) && (p.LeaderType == "" || p.LeaderType == NonePlayer)
}

func (v *validator) players() {
	teamIDs := make(map[uint]bool)
	for _, team := range v.m.Teams {
		teamIDs[team.TeamID] = true
	}

	active, playable := 0, 0
	for i, p := range v.m.Players {
		if isEmptySlot(p) {
			continue
		}
		active++
		if p.PlayableCiv {
			playable++
		}

		if (p.CivType == "" || p.CivType == NonePlayer) != (p.LeaderType == "" || p.LeaderType == NonePlayer) {
			v.add(SeverityError, "players", fmt.Sprintf("player %d must have both a civilization and a leader", i), -1, -1, i, -1)
		}
		if !teamIDs[p.Team] {
			v.add(SeverityError, "players", fmt.Sprintf("player %d belongs to team %d, which does not exist", i, p.Team), -1, -1, i, -1)
		}

		v.checkType("players", InfoCivilizations, "civilization", p.CivType, -1, -1, i, -1)
		v.checkType("players", InfoLeaders, "leader", p.LeaderType, -1, -1, i, -1)
		v.checkType("players", InfoHandicaps, "handicap", p.Handicap, -1, -1, i, -1)
		v.checkType("players", InfoPlayerColors, "player color", p.Color, -1, -1, i, -1)
		v.checkType("players", InfoArtStyles, "art style", p.ArtStyle, -1, -1, i, -1)
		v.checkType("players", InfoReligions, "religion", p.StateReligion, -1, -1, i, -1)
		v.checkType("players", InfoEras, "era", p.StartingEra, -1, -1, i, -1)
		for j, option := range p.CivicOption {
			v.checkType("players", InfoCivicOptions, "civic option", option, -1, -1, i, -1)
			if j < len(p.Civic) {
				v.checkType("players", InfoCivics, "civic", p.Civic[j], -1, -1, i, -1)
			}
		}

		if v.data != nil {
			if civ := v.data.Table(InfoCivilizations).Get(p.CivType); civ != nil && len(civ.Leaders) > 0 && !IsInSlice(civ.Leaders, p.LeaderType) {
				v.add(SeverityWarning, "players", fmt.Sprintf("leader %s is not a leader of %s in the game files", p.LeaderType, p.CivType), -1, -1, i, -1)
			}
		}

		if !p.RandomStartLocation {
			x, y := p.StartingX, p.StartingY
			switch {
			case !v.m.inGrid(x, y):
				v.add(SeverityError, "players", fmt.Sprintf("start position of player %d (%d, %d) is outside of the map", i, x, y), -1, -1, i, -1)
			default:
				if idx := v.m.FindPlot(x, y); idx >= 0 {
					switch v.m.Plots[idx].PlotType {
					case PlotOcean:
						v.add(SeverityError, "players", fmt.Sprintf("player %d starts in water", i), x, y, i, -1)
					case PlotPeak:
						v.add(SeverityWarning, "players", fmt.Sprintf("player %d starts on a peak", i), x, y, i, -1)
					}
				}
			}
		}
	}

	if active == 0 {
		v.add(SeverityError, "players", "there are no players, choose civilizations for some slots", -1, -1, -1, -1)
	} else if playable == 0 {
		v.add(SeverityWarning, "players", "no player is playable by a human", -1, -1, -1, -1)
	}
}

func (v *validator) plots() {
	players := v.m.Players
	// Owners equal to the number of slots are barbarians
	validOwner := func(owner int) bool {
		if owner == len(players) {
			return true
		}
		return owner >= 0 && owner < len(players) && !isEmptySlot(players[owner])
	}

	for _, p := range v.m.Plots {
		x, y := int(p.X), int(p.Y)
		v.checkType("plots", InfoTerrains, "terrain", p.TerrainType, x, y, -1, -1)
		for _, feature := range p.FeatureType {
			v.checkType("plots", InfoFeatures, "feature", feature, x, y, -1, -1)
		}
		v.checkType("plots", InfoBonuses, "resource", p.BonusType, x, y, -1, -1)
		v.checkType("plots", InfoImprovements, "improvement", p.ImprovementType, x, y, -1, -1)
		v.checkType("plots", InfoRoutes, "route", p.RouteType, x, y, -1, -1)

		for _, city := range p.Cities {
			owner := int(city.CityOwner)
			if !validOwner(owner) {
				v.add(SeverityError, "plots", fmt.Sprintf("city %s belongs to player %d, which is an empty slot", city.CityName, owner), x, y, owner, -1)
			}
			if p.PlotType == PlotOcean {
				v.add(SeverityError, "plots", fmt.Sprintf("city %s is in water", city.CityName), x, y, owner, -1)
			}
			if city.CityPopulation == 0 {
				v.add(SeverityWarning, "plots", fmt.Sprintf("city %s has no population", city.CityName), x, y, owner, -1)
			}
			for _, building := range city.BuildingType {
				v.checkType("plots", InfoBuildings, "building", building, x, y, owner, -1)
			}
			for _, religion := range append(append([]string{}, city.ReligionType...), city.HolyCityReligionType...) {
				v.checkType("plots", InfoReligions, "religion", religion, x, y, owner, -1)
			}
		}
		if len(p.Cities) > 1 {
			v.add(SeverityError, "plots", "more than one city on a plot", x, y, -1, -1)
		}

		for _, unit := range p.Units {
			if !validOwner(unit.UnitOwner) {
				v.add(SeverityError, "plots", fmt.Sprintf("unit %s belongs to player %d, which is an empty slot", unit.UnitType, unit.UnitOwner), x, y, unit.UnitOwner, -1)
			}
			v.checkType("plots", InfoUnits, "unit", unit.UnitType, x, y, unit.UnitOwner, -1)
			for _, promotion := range unit.PromotionType {
				v.checkType("plots", InfoPromotions, "promotion", promotion, x, y, unit.UnitOwner, -1)
			}
			v.checkType("plots", InfoUnitAIs, "unit AI", unit.UnitAIType, x, y, unit.UnitOwner, -1)
		}
	}
}

// ValidateMap checks the current map against loaded game data.
func (a *App) ValidateMap() []Problem {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return []Problem{}
	}
	return a.wbMap.Validate(CurrentGameData())
}
