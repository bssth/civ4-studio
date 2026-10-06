package editor

import (
	"errors"
	"fmt"
	"sort"
)

const (
	// PlotType values
	PlotPeak  = 0
	PlotHills = 1
	PlotLand  = 2
	PlotOcean = 3

	DefaultOceanTerrain = "TERRAIN_OCEAN"

	// MaxGridSize limits the size of a generated map, the game itself is unstable on much bigger maps
	MaxGridSize = 512
	// PlotsPerGridUnit is how many plots one grid unit of CIV4WorldInfo.xml (iGridWidth/iGridHeight) means
	PlotsPerGridUnit = 4
)

// DefaultPlayerSlots is the number of player (and team) slots in an unmodified Beyond the Sword
const DefaultPlayerSlots = 18

// NewWbMap returns an empty map with default settings and empty player slots, ready for plots to be created
func NewWbMap() *WbMap {
	wb := &WbMap{
		Version: defaultVersion,
		Game:    &Game{StartYear: -4000},
		Map: &MapProps{
			TopLatitude: 90, BottomLatitude: -90, WrapX: 1,
			WorldSize: "WORLDSIZE_STANDARD", Climate: "CLIMATE_TEMPERATE", SeaLevel: "SEALEVEL_MEDIUM",
		},
	}
	for i := uint(0); i < DefaultPlayerSlots; i++ {
		// Empty slots look the same as the game writes them
		wb.Teams = append(wb.Teams, &Team{TeamID: i, ContactWithTeam: []uint{i}})
		wb.Players = append(wb.Players, EmptyPlayer(i))
	}
	return wb
}

// GenerateOceanPlots creates width x height ocean plots in the order the game writes them
// (column by column: x=0,y=0; x=0,y=1; ...)
func GenerateOceanPlots(width, height uint) ([]*Plot, error) {
	if width == 0 || height == 0 {
		return nil, errors.New("map size must be positive")
	}
	if width > MaxGridSize || height > MaxGridSize {
		return nil, fmt.Errorf("map size must not exceed %dx%d", MaxGridSize, MaxGridSize)
	}

	plots := make([]*Plot, 0, width*height)
	for x := uint(0); x < width; x++ {
		for y := uint(0); y < height; y++ {
			plots = append(plots, &Plot{X: x, Y: y, TerrainType: DefaultOceanTerrain, PlotType: PlotOcean})
		}
	}
	return plots, nil
}

// CountStat is a value with the number of times it occurs
type CountStat struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// MapStats is a summary of map contents
type MapStats struct {
	Plots         int         `json:"plots"`
	ExpectedPlots int         `json:"expected_plots"`
	Peaks         int         `json:"peaks"`
	Hills         int         `json:"hills"`
	Flat          int         `json:"flat"`
	Water         int         `json:"water"`
	Cities        int         `json:"cities"`
	Units         int         `json:"units"`
	Signs         int         `json:"signs"`
	StartingPlots int         `json:"starting_plots"`
	Bonuses       int         `json:"bonuses"`
	Terrains      []CountStat `json:"terrains"`
	// Problems are inconsistencies found in the map
	Problems []Problem `json:"problems"`
}

// Stats calculates map statistics and checks plots for basic consistency
func (m *WbMap) Stats() *MapStats {
	stats := &MapStats{Signs: len(m.Signs), Plots: len(m.Plots), Problems: []Problem{}}
	mapProblem := func(code string, args map[string]string, message string) {
		stats.Problems = append(stats.Problems, newProblem(SeverityError, "map", code, args, message, -1, -1, -1, -1))
	}
	if m.Map != nil {
		stats.ExpectedPlots = int(m.Map.GridWidth * m.Map.GridHeight)
	}

	terrains := make(map[string]int)
	outside := 0
	for _, p := range m.Plots {
		switch p.PlotType {
		case PlotPeak:
			stats.Peaks++
		case PlotHills:
			stats.Hills++
		case PlotLand:
			stats.Flat++
		case PlotOcean:
			stats.Water++
		}
		if p.TerrainType != "" {
			terrains[p.TerrainType]++
		}
		if p.BonusType != "" {
			stats.Bonuses++
		}
		if p.StartingPlot {
			stats.StartingPlots++
		}
		stats.Cities += len(p.Cities)
		stats.Units += len(p.Units)
		if m.Map != nil && (uint64(p.X) >= m.Map.GridWidth || uint64(p.Y) >= m.Map.GridHeight) {
			outside++
		}
	}

	for t, n := range terrains {
		stats.Terrains = append(stats.Terrains, CountStat{Type: t, Count: n})
	}
	sort.Slice(stats.Terrains, func(i, j int) bool {
		if stats.Terrains[i].Count != stats.Terrains[j].Count {
			return stats.Terrains[i].Count > stats.Terrains[j].Count
		}
		return stats.Terrains[i].Type < stats.Terrains[j].Type
	})

	if m.Map == nil {
		mapProblem("map.noSection", nil, "the map has no BeginMap section")
	} else if stats.Plots != stats.ExpectedPlots {
		mapProblem("map.plotCount", args("plots", stats.Plots, "width", m.Map.GridWidth, "height", m.Map.GridHeight, "expected", stats.ExpectedPlots),
			fmt.Sprintf("the map has %d plots, but %dx%d = %d are expected", stats.Plots, m.Map.GridWidth, m.Map.GridHeight, stats.ExpectedPlots))
	}
	if outside > 0 {
		mapProblem("map.outside", args("n", outside), fmt.Sprintf("%d plots are outside of the map grid", outside))
	}
	if m.Map != nil && m.Map.TopLatitude <= m.Map.BottomLatitude {
		mapProblem("map.latitudes", nil, "top latitude must be greater than bottom latitude")
	}

	return stats
}
