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
// Code and Args let the frontend show the problem in the interface language; Message is the English text.
type Problem struct {
	Severity string            `json:"severity"`
	Section  string            `json:"section"`
	Code     string            `json:"code"`
	Args     map[string]string `json:"args"`
	Message  string            `json:"message"`
	X        int               `json:"x"`
	Y        int               `json:"y"`
	Player   int               `json:"player"`
	Team     int               `json:"team"`
}

func newProblem(severity, section, code string, args map[string]string, message string, x, y, player, team int) Problem {
	if args == nil {
		args = map[string]string{}
	}
	return Problem{Severity: severity, Section: section, Code: code, Args: args, Message: message,
		X: x, Y: y, Player: player, Team: team}
}

// args builds problem arguments from name/value pairs
func args(pairs ...any) map[string]string {
	result := make(map[string]string, len(pairs)/2)
	for i := 0; i+1 < len(pairs); i += 2 {
		result[fmt.Sprint(pairs[i])] = fmt.Sprint(pairs[i+1])
	}
	return result
}

// typeNames are English names of the kinds of types checked against game data, by code
var typeNames = map[string]string{
	"era": "era", "speed": "game speed", "calendar": "calendar", "victory": "victory", "gameOption": "game option",
	"mpOption": "multiplayer option", "forceControl": "force control", "worldSize": "world size", "climate": "climate",
	"seaLevel": "sea level", "tech": "tech", "project": "project", "civilization": "civilization", "leader": "leader",
	"handicap": "handicap", "playerColor": "player color", "artStyle": "art style", "religion": "religion",
	"civicOption": "civic option", "civic": "civic", "terrain": "terrain", "feature": "feature", "resource": "resource",
	"improvement": "improvement", "route": "route", "building": "building", "unit": "unit", "promotion": "promotion",
	"unitAI": "unit AI",
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

func (v *validator) add(severity, section, code string, args map[string]string, message string, x, y, player, team int) {
	v.problems = append(v.problems, newProblem(severity, section, code, args, message, x, y, player, team))
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
		problem: newProblem(SeverityError, section, "unknownType", args("what", what, "value", value),
			fmt.Sprintf("unknown %s %s", typeNames[what], value), x, y, player, team),
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
	v.signs()

	for _, key := range v.unknownOrder {
		u := v.unknown[key]
		u.problem.Args["count"] = fmt.Sprint(u.count)
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
		v.add(SeverityError, "game", "game.noSection", nil, "the map has no BeginGame section", -1, -1, -1, -1)
		return
	}
	v.checkType("game", InfoEras, "era", g.Era, -1, -1, -1, -1)
	v.checkType("game", InfoSpeeds, "speed", g.Speed, -1, -1, -1, -1)
	v.checkType("game", InfoCalendars, "calendar", g.Calendar, -1, -1, -1, -1)
	for _, victory := range g.Victory {
		v.checkType("game", InfoVictories, "victory", victory, -1, -1, -1, -1)
	}
	for _, option := range g.Option {
		v.checkType("game", InfoGameOptions, "gameOption", option, -1, -1, -1, -1)
	}
	for _, option := range g.MPOption {
		v.checkType("game", InfoMPOptions, "mpOption", option, -1, -1, -1, -1)
	}
	for _, option := range g.ForceControl {
		v.checkType("game", InfoForceControls, "forceControl", option, -1, -1, -1, -1)
	}
	if g.MaxTurns > 0 && g.MaxTurns <= g.GameTurn {
		v.add(SeverityError, "game", "game.maxTurns", args("max", g.MaxTurns, "start", g.GameTurn),
			fmt.Sprintf("max turns (%d) must be greater than the starting turn (%d)", g.MaxTurns, g.GameTurn), -1, -1, -1, -1)
	}
}

func (v *validator) mapSection() {
	v.problems = append(v.problems, v.m.Stats().Problems...)
	if v.m.Map != nil {
		v.checkType("map", InfoWorldSizes, "worldSize", v.m.Map.WorldSize, -1, -1, -1, -1)
		v.checkType("map", InfoClimates, "climate", v.m.Map.Climate, -1, -1, -1, -1)
		v.checkType("map", InfoSeaLevels, "seaLevel", v.m.Map.SeaLevel, -1, -1, -1, -1)
	}
}

// relationNames are English names of team relations, by code
var relationNames = map[string]string{
	"contact": "contact", "war": "war", "openBorders": "open borders",
	"defensivePact": "defensive pact", "permanent": "permanent war/peace",
}

func (v *validator) teams() {
	ids := make(map[uint]bool)
	for _, team := range v.m.Teams {
		if ids[team.TeamID] {
			v.add(SeverityError, "teams", "teams.duplicate", args("team", team.TeamID),
				fmt.Sprintf("team %d is defined more than once", team.TeamID), -1, -1, -1, int(team.TeamID))
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
			"contact": team.ContactWithTeam, "war": team.AtWar, "openBorders": team.OpenBordersWithTeam,
			"defensivePact": team.DefensivePactWithTeam, "permanent": team.PermanentWarPeace,
		}
		for _, relation := range SortKeys(relations) {
			for _, other := range relations[relation] {
				if !ids[other] {
					v.add(SeverityError, "teams", "teams.missingRelation", args("team", id, "relation", relation, "other", other),
						fmt.Sprintf("team %d has %s with team %d, which does not exist", id, relationNames[relation], other), -1, -1, -1, id)
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
			v.add(SeverityError, "players", "players.civLeader", args("player", i),
				fmt.Sprintf("player %d must have both a civilization and a leader", i), -1, -1, i, -1)
		}
		if !teamIDs[p.Team] {
			v.add(SeverityError, "players", "players.noTeam", args("player", i, "team", p.Team),
				fmt.Sprintf("player %d belongs to team %d, which does not exist", i, p.Team), -1, -1, i, -1)
		}

		v.checkType("players", InfoCivilizations, "civilization", p.CivType, -1, -1, i, -1)
		v.checkType("players", InfoLeaders, "leader", p.LeaderType, -1, -1, i, -1)
		v.checkType("players", InfoHandicaps, "handicap", p.Handicap, -1, -1, i, -1)
		v.checkType("players", InfoPlayerColors, "playerColor", p.Color, -1, -1, i, -1)
		v.checkType("players", InfoArtStyles, "artStyle", p.ArtStyle, -1, -1, i, -1)
		v.checkType("players", InfoReligions, "religion", p.StateReligion, -1, -1, i, -1)
		v.checkType("players", InfoEras, "era", p.StartingEra, -1, -1, i, -1)
		for j, option := range p.CivicOption {
			v.checkType("players", InfoCivicOptions, "civicOption", option, -1, -1, i, -1)
			if j < len(p.Civic) {
				v.checkType("players", InfoCivics, "civic", p.Civic[j], -1, -1, i, -1)
			}
		}

		if v.data != nil {
			if civ := v.data.Table(InfoCivilizations).Get(p.CivType); civ != nil && len(civ.Leaders) > 0 && !IsInSlice(civ.Leaders, p.LeaderType) {
				v.add(SeverityWarning, "players", "players.foreignLeader", args("leader", p.LeaderType, "civ", p.CivType),
					fmt.Sprintf("leader %s is not a leader of %s in the game files", p.LeaderType, p.CivType), -1, -1, i, -1)
			}
		}

		if !p.RandomStartLocation {
			x, y := p.StartingX, p.StartingY
			switch {
			case !v.m.inGrid(x, y):
				v.add(SeverityError, "players", "players.startOutside", args("player", i, "x", x, "y", y),
					fmt.Sprintf("start position of player %d (%d, %d) is outside of the map", i, x, y), -1, -1, i, -1)
			default:
				if idx := v.m.FindPlot(x, y); idx >= 0 {
					switch v.m.Plots[idx].PlotType {
					case PlotOcean:
						v.add(SeverityError, "players", "players.startWater", args("player", i),
							fmt.Sprintf("player %d starts in water", i), x, y, i, -1)
					case PlotPeak:
						v.add(SeverityWarning, "players", "players.startPeak", args("player", i),
							fmt.Sprintf("player %d starts on a peak", i), x, y, i, -1)
					}
				}
			}
		}
	}

	if active == 0 {
		v.add(SeverityError, "players", "players.none", nil, "there are no players, choose civilizations for some slots", -1, -1, -1, -1)
	} else if playable == 0 {
		v.add(SeverityWarning, "players", "players.noPlayable", nil, "no player is playable by a human", -1, -1, -1, -1)
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
				v.add(SeverityError, "plots", "plots.cityEmptyOwner", args("city", city.CityName, "player", owner),
					fmt.Sprintf("city %s belongs to player %d, which is an empty slot", city.CityName, owner), x, y, owner, -1)
			}
			if p.PlotType == PlotOcean {
				v.add(SeverityError, "plots", "plots.cityWater", args("city", city.CityName),
					fmt.Sprintf("city %s is in water", city.CityName), x, y, owner, -1)
			}
			if city.CityPopulation == 0 {
				v.add(SeverityWarning, "plots", "plots.cityNoPopulation", args("city", city.CityName),
					fmt.Sprintf("city %s has no population", city.CityName), x, y, owner, -1)
			}
			for _, building := range city.BuildingType {
				v.checkType("plots", InfoBuildings, "building", building, x, y, owner, -1)
			}
			for _, religion := range append(append([]string{}, city.ReligionType...), city.HolyCityReligionType...) {
				v.checkType("plots", InfoReligions, "religion", religion, x, y, owner, -1)
			}
		}
		if len(p.Cities) > 1 {
			v.add(SeverityError, "plots", "plots.twoCities", nil, "more than one city on a plot", x, y, -1, -1)
		}

		for _, unit := range p.Units {
			if !validOwner(unit.UnitOwner) {
				v.add(SeverityError, "plots", "plots.unitEmptyOwner", args("unit", unit.UnitType, "player", unit.UnitOwner),
					fmt.Sprintf("unit %s belongs to player %d, which is an empty slot", unit.UnitType, unit.UnitOwner), x, y, unit.UnitOwner, -1)
			}
			v.checkType("plots", InfoUnits, "unit", unit.UnitType, x, y, unit.UnitOwner, -1)
			for _, promotion := range unit.PromotionType {
				v.checkType("plots", InfoPromotions, "promotion", promotion, x, y, unit.UnitOwner, -1)
			}
			v.checkType("plots", InfoUnitAIs, "unitAI", unit.UnitAIType, x, y, unit.UnitOwner, -1)
		}
	}
}

func (v *validator) signs() {
	players := v.m.Players
	for _, sign := range v.m.Signs {
		x, y := sign.PlotX, sign.PlotY
		if !v.m.inGrid(x, y) {
			v.add(SeverityError, "plots", "plots.signOutside", args("caption", sign.Caption, "x", x, "y", y),
				fmt.Sprintf("sign %q at %d,%d is outside of the map", sign.Caption, x, y), -1, -1, -1, -1)
			continue
		}
		if p := sign.PlayerType; p >= 0 && (p >= len(players) || isEmptySlot(players[p])) {
			v.add(SeverityWarning, "plots", "plots.signPlayer", args("caption", sign.Caption, "player", p),
				fmt.Sprintf("sign %q is shown only to player %d, which is an empty slot", sign.Caption, p), x, y, p, -1)
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
