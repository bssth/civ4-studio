package editor

import (
	"math"
	"sort"
)

// Resources of Beyond the Sword by what they give a starting city; others (e.g. of mods) count as luxury
var (
	foodResources = map[string]bool{
		"BONUS_WHEAT": true, "BONUS_CORN": true, "BONUS_RICE": true, "BONUS_COW": true, "BONUS_PIG": true,
		"BONUS_SHEEP": true, "BONUS_DEER": true, "BONUS_BANANA": true, "BONUS_FISH": true, "BONUS_CLAM": true,
		"BONUS_CRAB": true,
	}
	strategicResources = map[string]bool{
		"BONUS_IRON": true, "BONUS_COPPER": true, "BONUS_HORSE": true, "BONUS_STONE": true, "BONUS_MARBLE": true,
		"BONUS_IVORY": true, "BONUS_COAL": true, "BONUS_OIL": true, "BONUS_ALUMINUM": true, "BONUS_URANIUM": true,
	}
)

// StartInfo describes the land around a start position: the 21 plots a city there works (the "fat cross")
type StartInfo struct {
	Player int `json:"player"`
	X      int `json:"x"`
	Y      int `json:"y"`
	// Land plots, of them flat or hilly grassland and plains ("good"), hills, peaks and forests or jungle
	Land    int `json:"land"`
	Good    int `json:"good"`
	Hills   int `json:"hills"`
	Peaks   int `json:"peaks"`
	Forests int `json:"forests"`
	Water   int `json:"water"`
	// The start plot is next to water (a harbour is possible) or at a river
	Coastal bool `json:"coastal"`
	River   bool `json:"river"`
	// Resources by kind and their types
	Food      int      `json:"food"`
	Strategic int      `json:"strategic"`
	Luxury    int      `json:"luxury"`
	Resources []string `json:"resources"`
	// Distance to the nearest other start, 0 if there is no other
	Nearest float64 `json:"nearest"`
}

// StartBalance analyses the start positions of all players with a civilization and a fixed start on the map.
func (m *WbMap) StartBalance() []StartInfo {
	result := []StartInfo{}
	if m.Map == nil || m.Map.GridWidth == 0 {
		return result
	}
	g := terrainGrid{w: int(m.Map.GridWidth), h: int(m.Map.GridHeight), wrapX: m.Map.WrapX != 0}
	plots := make([]*Plot, g.w*g.h)
	for _, p := range m.Plots {
		if i, ok := g.index(int(p.X), int(p.Y)); ok {
			plots[i] = p
		}
	}
	at := func(x, y int) *Plot {
		if i, ok := g.index(x, y); ok {
			return plots[i]
		}
		return nil
	}

	for i, pl := range m.Players {
		if isEmptySlot(pl) || pl.RandomStartLocation || at(pl.StartingX, pl.StartingY) == nil {
			continue
		}
		s := StartInfo{Player: i, X: pl.StartingX, Y: pl.StartingY, Resources: []string{}}
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if abs(dx) == 2 && abs(dy) == 2 {
					continue
				}
				p := at(s.X+dx, s.Y+dy)
				if p == nil {
					continue
				}
				if p.PlotType == PlotOcean {
					s.Water++
					if abs(dx) <= 1 && abs(dy) <= 1 && !(dx == 0 && dy == 0) {
						s.Coastal = true
					}
				} else {
					s.Land++
					switch p.PlotType {
					case PlotHills:
						s.Hills++
					case PlotPeak:
						s.Peaks++
					}
					if p.PlotType != PlotPeak && (p.TerrainType == "TERRAIN_GRASS" || p.TerrainType == "TERRAIN_PLAINS") {
						s.Good++
					}
					if len(p.FeatureType) > 0 && (p.FeatureType[0] == "FEATURE_FOREST" || p.FeatureType[0] == "FEATURE_JUNGLE") {
						s.Forests++
					}
				}
				if b := p.BonusType; b != "" && b != NonePlayer {
					s.Resources = append(s.Resources, b)
					switch {
					case foodResources[b]:
						s.Food++
					case strategicResources[b]:
						s.Strategic++
					default:
						s.Luxury++
					}
				}
			}
		}
		sort.Strings(s.Resources)
		// A river along any edge of the start plot: its own southern and eastern edges,
		// the southern edge of the plot above and the eastern edge of the plot on the left
		self := at(s.X, s.Y)
		above, left := at(s.X, s.Y+1), at(s.X-1, s.Y)
		s.River = self.IsNOfRiver || self.IsWOfRiver || (above != nil && above.IsNOfRiver) || (left != nil && left.IsWOfRiver)
		result = append(result, s)
	}

	for i := range result {
		nearest := math.Inf(1)
		for j := range result {
			if i != j {
				nearest = math.Min(nearest, g.distance(result[i].X, result[i].Y, result[j].X, result[j].Y))
			}
		}
		if !math.IsInf(nearest, 1) {
			result[i].Nearest = math.Round(nearest*10) / 10
		}
	}
	return result
}

// StartBalance returns the analysis of the start positions, see WbMap.StartBalance.
func (a *App) StartBalance() []StartInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return []StartInfo{}
	}
	return a.wbMap.StartBalance()
}
