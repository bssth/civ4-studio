package editor

import (
	"bytes"
	"testing"
)

func terrainTestMap(t *testing.T, w, h uint) *WbMap {
	t.Helper()
	wb := NewWbMap()
	plots, err := GenerateOceanPlots(w, h)
	if err != nil {
		t.Fatal(err)
	}
	wb.Map.GridWidth, wb.Map.GridHeight, wb.Plots = uint64(w), uint64(h), plots
	return wb
}

var defaultTerrain = TerrainOptions{Seed: 42, Land: 35, Continents: 3, Hills: 15, Peaks: 5, Forests: 40, Rivers: 12, Resources: true}

func TestGenerateTerrainIsReproducible(t *testing.T) {
	a, b := terrainTestMap(t, 64, 40), terrainTestMap(t, 64, 40)
	if _, err := a.GenerateTerrain(defaultTerrain, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := b.GenerateTerrain(defaultTerrain, nil); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.ToWbFormat(), b.ToWbFormat()) {
		t.Error("the same seed must give the same map")
	}
	other := defaultTerrain
	other.Seed = 7
	c := terrainTestMap(t, 64, 40)
	if _, err := c.GenerateTerrain(other, nil); err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(a.ToWbFormat(), c.ToWbFormat()) {
		t.Error("another seed must give another map")
	}
}

func TestGenerateTerrainShape(t *testing.T) {
	wb := terrainTestMap(t, 80, 50)
	wb.Plots[0].Cities = []*City{{CityName: "Kept"}}
	result, err := wb.GenerateTerrain(defaultTerrain, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := terrainGrid{w: 80, h: 50, wrapX: true}
	plots := make([]*Plot, 80*50)
	for _, p := range wb.Plots {
		i, _ := g.index(int(p.X), int(p.Y))
		plots[i] = p
	}
	land, hills, peaks := 0, 0, 0
	for _, p := range wb.Plots {
		if p.PlotType != PlotOcean {
			land++
		}
		switch p.PlotType {
		case PlotHills:
			hills++
		case PlotPeak:
			peaks++
		}
	}
	if share := float64(land) / float64(len(wb.Plots)); share < 0.32 || share > 0.38 || land != result.Land {
		t.Errorf("land share %.2f (%d, reported %d)", share, land, result.Land)
	}
	if h := float64(hills+peaks) / float64(land); h < 0.15 || h > 0.25 {
		t.Errorf("hills and peaks share of land %.2f", h)
	}
	isLand := func(x, y int) bool {
		i, ok := g.index(x, y)
		return ok && plots[i].PlotType != PlotOcean
	}
	for _, p := range wb.Plots {
		x, y := int(p.X), int(p.Y)
		water := p.PlotType == PlotOcean
		if water != (p.TerrainType == "TERRAIN_OCEAN" || p.TerrainType == "TERRAIN_COAST") {
			t.Fatalf("plot %d,%d: height %d with terrain %s", x, y, p.PlotType, p.TerrainType)
		}
		if p.TerrainType == "TERRAIN_COAST" && !(isLand(x-1, y) || isLand(x+1, y) || isLand(x, y-1) || isLand(x, y+1) ||
			isLand(x-1, y-1) || isLand(x+1, y+1) || isLand(x-1, y+1) || isLand(x+1, y-1)) {
			t.Errorf("coast without land at %d,%d", x, y)
		}
		// Rivers run along edges between two land plots
		if p.IsNOfRiver && (water || !isLand(x, y-1)) {
			t.Errorf("river on the southern edge of %d,%d is not between land plots", x, y)
		}
		if p.IsWOfRiver && (water || !isLand(x+1, y)) {
			t.Errorf("river on the eastern edge of %d,%d is not between land plots", x, y)
		}
		if p.PlotType == PlotPeak && p.BonusType != "" {
			t.Errorf("a resource on a peak at %d,%d", x, y)
		}
	}
	// Poles are cold
	for _, p := range wb.Plots {
		if p.PlotType != PlotOcean && (p.Y == 0 || p.Y == 49) && p.TerrainType != "TERRAIN_SNOW" {
			t.Errorf("land at the pole %d,%d is %s", p.X, p.Y, p.TerrainType)
		}
	}
	if result.Rivers == 0 || result.Resources == 0 {
		t.Errorf("result = %+v", result)
	}
	if len(wb.Plots[0].Cities) != 1 {
		t.Error("cities must stay")
	}
}

func TestGenerateTerrainUsesBonusesOfGameData(t *testing.T) {
	wb := terrainTestMap(t, 40, 30)
	data := NewGameData()
	data.Table(InfoBonuses).Set(&TypeInfo{Type: "BONUS_FISH"})
	if _, err := wb.GenerateTerrain(defaultTerrain, data); err != nil {
		t.Fatal(err)
	}
	for _, p := range wb.Plots {
		if p.BonusType != "" && p.BonusType != "BONUS_FISH" {
			t.Fatalf("resource %s is not in the game data", p.BonusType)
		}
	}
}

func TestGenerateTerrainOptions(t *testing.T) {
	wb := terrainTestMap(t, 20, 20)
	for _, o := range []TerrainOptions{
		{Land: 2, Continents: 1}, {Land: 30, Continents: 0}, {Land: 30, Continents: 1, Hills: 70, Peaks: 20},
		{Land: 30, Continents: 1, Rivers: -1},
	} {
		if _, err := wb.GenerateTerrain(o, nil); err == nil {
			t.Errorf("options %+v must be refused", o)
		}
	}
	wb.Plots = wb.Plots[:10]
	if _, err := wb.GenerateTerrain(defaultTerrain, nil); err == nil {
		t.Error("a map without all plots must be refused")
	}
}

func TestPlaceStarts(t *testing.T) {
	wb := terrainTestMap(t, 80, 50)
	if _, err := wb.GenerateTerrain(defaultTerrain, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := wb.PlaceStarts(1); err == nil {
		t.Error("a map without players must be refused")
	}
	for i := 0; i < 6; i++ {
		wb.Players[i].CivType, wb.Players[i].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
		wb.Players[i].RandomStartLocation = true
	}
	players, err := wb.PlaceStarts(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 6 {
		t.Fatalf("players = %v", players)
	}
	g := terrainGrid{w: 80, h: 50, wrapX: true}
	for _, i := range players {
		p := wb.Players[i]
		plot := wb.Plots[wb.FindPlot(p.StartingX, p.StartingY)]
		if p.RandomStartLocation || plot.PlotType == PlotOcean || plot.PlotType == PlotPeak ||
			(plot.TerrainType != "TERRAIN_GRASS" && plot.TerrainType != "TERRAIN_PLAINS") {
			t.Errorf("player %d starts at %d,%d on %s (%d)", i, p.StartingX, p.StartingY, plot.TerrainType, plot.PlotType)
		}
		for _, j := range players {
			q := wb.Players[j]
			if i < j && g.distance(p.StartingX, p.StartingY, q.StartingX, q.StartingY) < 6 {
				t.Errorf("players %d and %d start too close", i, j)
			}
		}
	}
}

func TestAppGenerateTerrainUndo(t *testing.T) {
	app := NewApp()
	app.wbMap = terrainTestMap(t, 30, 20)
	before := app.wbMap.ToWbFormat()
	if _, err := app.GenerateTerrain(defaultTerrain); err != nil {
		t.Fatal(err)
	}
	if s := app.HistoryState(); s.Undo != "generate:42" {
		t.Errorf("undo label = %q", s.Undo)
	}
	app.Undo()
	if !bytes.Equal(before, app.wbMap.ToWbFormat()) {
		t.Error("undo must restore the map")
	}
	if _, err := app.GenerateTerrain(TerrainOptions{Land: 1}); err == nil || !bytes.Equal(before, app.wbMap.ToWbFormat()) {
		t.Error("refused options must change nothing")
	}
}
