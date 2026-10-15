package editor

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
)

// problemFixes are the automatic fixes of problem codes, see WbMap.Fix
var problemFixes = map[string]string{
	"unknownType":            "replace",
	"game.maxTurns":          "resetMaxTurns",
	"teams.missingRelation":  "removeRelation",
	"players.startOutside":   "moveStart",
	"players.startWater":     "moveStart",
	"players.startPeak":      "moveStart",
	"plots.cityEmptyOwner":   "removeCity",
	"plots.cityWater":        "makeLand",
	"plots.cityNoPopulation": "setPopulation",
	"plots.twoCities":        "keepFirstCity",
	"plots.unitEmptyOwner":   "removeUnits",
	"plots.signOutside":      "removeSign",
	"plots.signPlayer":       "showSignToAll",
}

func argInt(p Problem, name string) (int, bool) {
	v, err := strconv.Atoi(p.Args[name])
	return v, err == nil
}

// plotAt returns the plot of a problem, nil if there is none
func (m *WbMap) plotAt(x, y int) *Plot {
	if i := m.FindPlot(x, y); i >= 0 {
		return m.Plots[i]
	}
	return nil
}

// Fix applies the automatic fix of a problem. It returns false when the problem has no fix
// or it is already fixed (e.g. by an earlier fix of the same batch).
func (m *WbMap) Fix(p Problem) bool {
	switch problemFixes[p.Code] {
	case "resetMaxTurns":
		if m.Game == nil || m.Game.MaxTurns == 0 {
			return false
		}
		// Zero means the length of the game speed
		m.Game.MaxTurns = 0
		return true

	case "removeRelation":
		other, ok := argInt(p, "other")
		if !ok {
			return false
		}
		for _, team := range m.Teams {
			if int(team.TeamID) != p.Team {
				continue
			}
			lists := map[string]*[]uint{
				"contact": &team.ContactWithTeam, "war": &team.AtWar, "openBorders": &team.OpenBordersWithTeam,
				"defensivePact": &team.DefensivePactWithTeam, "permanent": &team.PermanentWarPeace,
			}
			list := lists[p.Args["relation"]]
			if list == nil {
				return false
			}
			before := len(*list)
			*list = slices.DeleteFunc(*list, func(v uint) bool { return int(v) == other })
			return len(*list) != before
		}
		return false

	case "moveStart":
		if p.Player < 0 || p.Player >= len(m.Players) {
			return false
		}
		x, y, ok := m.nearestStart(p.Player)
		if !ok {
			return false
		}
		pl := m.Players[p.Player]
		pl.StartingX, pl.StartingY, pl.RandomStartLocation = x, y, false
		return true

	case "removeCity":
		plot := m.plotAt(p.X, p.Y)
		if plot == nil {
			return false
		}
		before := len(plot.Cities)
		plot.Cities = slices.DeleteFunc(plot.Cities, func(c *City) bool { return int(c.CityOwner) == p.Player })
		return len(plot.Cities) != before

	case "makeLand":
		plot := m.plotAt(p.X, p.Y)
		if plot == nil || plot.PlotType != PlotOcean {
			return false
		}
		plot.PlotType, plot.TerrainType = PlotLand, "TERRAIN_GRASS"
		plot.FeatureType, plot.FeatureVariety, plot.BonusType = nil, nil, ""
		return true

	case "setPopulation":
		plot := m.plotAt(p.X, p.Y)
		changed := false
		if plot != nil {
			for _, c := range plot.Cities {
				if c.CityPopulation == 0 {
					c.CityPopulation, changed = 1, true
				}
			}
		}
		return changed

	case "keepFirstCity":
		plot := m.plotAt(p.X, p.Y)
		if plot == nil || len(plot.Cities) < 2 {
			return false
		}
		plot.Cities = plot.Cities[:1]
		return true

	case "removeUnits":
		plot := m.plotAt(p.X, p.Y)
		if plot == nil {
			return false
		}
		before := len(plot.Units)
		plot.Units = slices.DeleteFunc(plot.Units, func(u *Unit) bool { return u.UnitOwner == p.Player })
		return len(plot.Units) != before

	case "removeSign":
		x, okX := argInt(p, "x")
		y, okY := argInt(p, "y")
		if !okX || !okY {
			return false
		}
		before := len(m.Signs)
		m.Signs = slices.DeleteFunc(m.Signs, func(s *Sign) bool {
			return s.PlotX == x && s.PlotY == y && s.Caption == p.Args["caption"]
		})
		return len(m.Signs) != before

	case "showSignToAll":
		changed := false
		for _, s := range m.Signs {
			if s.PlotX == p.X && s.PlotY == p.Y && s.Caption == p.Args["caption"] && s.PlayerType == p.Player {
				s.PlayerType, changed = -1, true
			}
		}
		return changed
	}
	return false
}

// nearestStart finds the land plot nearest to the current start of a player that is flat or hilly,
// not taken by another start, preferring grassland and plains within a few plots
func (m *WbMap) nearestStart(player int) (int, int, bool) {
	if m.Map == nil {
		return 0, 0, false
	}
	g := m.grid()
	p := m.Players[player]
	sx, sy := min(max(p.StartingX, 0), g.w-1), min(max(p.StartingY, 0), g.h-1)
	taken := make(map[[2]int]bool)
	for i, q := range m.Players {
		if i != player && !isEmptySlot(q) && !q.RandomStartLocation {
			taken[[2]int{q.StartingX, q.StartingY}] = true
		}
	}
	best, bestScore := -1, 0.0
	for i, plot := range m.Plots {
		if plot.PlotType != PlotLand && plot.PlotType != PlotHills {
			continue
		}
		x, y := int(plot.X), int(plot.Y)
		if taken[[2]int{x, y}] || x >= g.w || y >= g.h {
			continue
		}
		score := g.distance(sx, sy, x, y)
		if plot.TerrainType != "TERRAIN_GRASS" && plot.TerrainType != "TERRAIN_PLAINS" {
			score += 3
		}
		if best < 0 || score < bestScore {
			best, bestScore = i, score
		}
	}
	if best < 0 {
		return 0, 0, false
	}
	return int(m.Plots[best].X), int(m.Plots[best].Y), true
}

// FixProblems applies the automatic fixes of the given problems as one undoable step.
// Returns the number of applied fixes.
func (a *App) FixProblems(problems []Problem) (int, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return 0, errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	fixed := 0
	for _, p := range problems {
		if a.wbMap.Fix(p) {
			fixed++
		}
	}
	if fixed > 0 {
		a.history.push(snapshotEntry(fmt.Sprintf("fix:%d", fixed), before, a.wbMap.snapshot()))
	}
	a.mu.Unlock()

	if fixed > 0 {
		a.setDirty(true)
		ConsoleWrite("Fixed %d problems", fixed)
	}
	return fixed, nil
}

// replaceIn replaces a value in a list; an empty replacement removes it. Returns the number of changes.
func replaceIn(list *[]string, from, to string) int {
	n := 0
	for i, v := range *list {
		if v == from {
			(*list)[i] = to
			n++
		}
	}
	if to == "" && n > 0 {
		*list = slices.DeleteFunc(*list, func(v string) bool { return v == "" })
	}
	return n
}

func replaceValue(field *string, from, to string) int {
	if *field == from {
		*field = to
		return 1
	}
	return 0
}

// requiredKinds are kinds of types that can not be removed, only replaced with another type
var requiredKinds = map[string]bool{
	"era": true, "speed": true, "calendar": true, "worldSize": true, "climate": true, "seaLevel": true,
	"civilization": true, "leader": true, "handicap": true, "playerColor": true, "artStyle": true,
	"civicOption": true, "terrain": true,
}

// ReplaceType replaces every use of a type of the given kind ("what" of unknownType problems) with another
// type; an empty replacement removes it where it is optional (a resource, a building, a promotion, a unit...).
// Returns the number of replaced uses.
func (m *WbMap) ReplaceType(what, from, to string) (int, error) {
	if from == "" || from == to {
		return 0, errors.New("choose another type to replace with")
	}
	if to == "" && requiredKinds[what] {
		return 0, fmt.Errorf("a %s can not be removed, choose a replacement", typeNames[what])
	}
	n := 0
	if g := m.Game; g != nil {
		switch what {
		case "era":
			n += replaceValue(&g.Era, from, to)
		case "speed":
			n += replaceValue(&g.Speed, from, to)
		case "calendar":
			n += replaceValue(&g.Calendar, from, to)
		case "victory":
			n += replaceIn(&g.Victory, from, to)
		case "gameOption":
			n += replaceIn(&g.Option, from, to)
		case "mpOption":
			n += replaceIn(&g.MPOption, from, to)
		case "forceControl":
			n += replaceIn(&g.ForceControl, from, to)
		}
	}
	if p := m.Map; p != nil {
		switch what {
		case "worldSize":
			n += replaceValue(&p.WorldSize, from, to)
		case "climate":
			n += replaceValue(&p.Climate, from, to)
		case "seaLevel":
			n += replaceValue(&p.SeaLevel, from, to)
		}
	}
	for _, t := range m.Teams {
		switch what {
		case "tech":
			n += replaceIn(&t.Tech, from, to)
		case "project":
			n += replaceIn(&t.ProjectType, from, to)
		}
	}
	for _, p := range m.Players {
		switch what {
		case "civilization":
			n += replaceValue(&p.CivType, from, to)
		case "leader":
			n += replaceValue(&p.LeaderType, from, to)
		case "handicap":
			n += replaceValue(&p.Handicap, from, to)
		case "playerColor":
			n += replaceValue(&p.Color, from, to)
		case "artStyle":
			n += replaceValue(&p.ArtStyle, from, to)
		case "religion":
			n += replaceValue(&p.StateReligion, from, to)
		case "era":
			n += replaceValue(&p.StartingEra, from, to)
		case "civicOption":
			n += replaceIn(&p.CivicOption, from, to)
		case "civic":
			// Civics are paired with civic options, keep the pairs aligned
			for i := range p.Civic {
				if p.Civic[i] == from {
					p.Civic[i] = to
					if to == "" {
						p.Civic[i] = NonePlayer
					}
					n++
				}
			}
		}
	}
	for _, plot := range m.Plots {
		switch what {
		case "terrain":
			n += replaceValue(&plot.TerrainType, from, to)
		case "feature":
			for i := len(plot.FeatureType) - 1; i >= 0; i-- {
				if plot.FeatureType[i] != from {
					continue
				}
				n++
				if to != "" {
					plot.FeatureType[i] = to
					continue
				}
				plot.FeatureType = slices.Delete(plot.FeatureType, i, i+1)
				if i < len(plot.FeatureVariety) {
					plot.FeatureVariety = slices.Delete(plot.FeatureVariety, i, i+1)
				}
			}
		case "resource":
			n += replaceValue(&plot.BonusType, from, to)
		case "improvement":
			n += replaceValue(&plot.ImprovementType, from, to)
		case "route":
			n += replaceValue(&plot.RouteType, from, to)
		}
		for _, c := range plot.Cities {
			switch what {
			case "building":
				n += replaceIn(&c.BuildingType, from, to) + replaceValue(&c.ProductionBuilding, from, to)
			case "religion":
				n += replaceIn(&c.ReligionType, from, to) + replaceIn(&c.HolyCityReligionType, from, to)
			case "unit":
				n += replaceValue(&c.ProductionUnit, from, to)
			case "project":
				n += replaceValue(&c.ProductionProject, from, to)
			case "process":
				n += replaceValue(&c.ProductionProcess, from, to)
			}
		}
		units := plot.Units[:0]
		for _, u := range plot.Units {
			switch what {
			case "unit":
				if u.UnitType == from {
					n++
					if to == "" {
						continue
					}
					u.UnitType = to
				}
			case "promotion":
				n += replaceIn(&u.PromotionType, from, to)
			case "unitAI":
				n += replaceValue(&u.UnitAIType, from, to)
			}
			units = append(units, u)
		}
		plot.Units = units
	}
	return n, nil
}

// ReplaceType replaces a type everywhere in the map as one undoable step, see WbMap.ReplaceType.
func (a *App) ReplaceType(what, from, to string) (int, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return 0, errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	n, err := a.wbMap.ReplaceType(what, from, to)
	if err != nil || n == 0 {
		before.restore(a.wbMap)
		a.mu.Unlock()
		return 0, err
	}
	a.history.push(snapshotEntry("replace:"+what, before, a.wbMap.snapshot()))
	a.mu.Unlock()

	a.setDirty(true)
	ConsoleWrite("Replaced %s %s with %q (%d times)", typeNames[what], from, to, n)
	return n, nil
}
