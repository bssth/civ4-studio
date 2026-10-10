package editor

import (
	"math"
	"math/rand"
	"testing"
)

func TestTerrainGridWrapping(t *testing.T) {
	torus := terrainGrid{w: 10, h: 6, wrapX: true, wrapY: true}
	if i, ok := torus.index(-1, 6); !ok || i != 9 {
		t.Errorf("torus index(-1, 6) = %d, %v", i, ok)
	}
	if d := torus.distance(0, 0, 9, 5); d != math.Sqrt2 {
		t.Errorf("torus distance = %v", d)
	}
	if cx, cy, ok := torus.corner(10, -1); !ok || cx != 0 || cy != 5 {
		t.Errorf("torus corner = %d, %d, %v", cx, cy, ok)
	}

	ns := terrainGrid{w: 10, h: 6, wrapY: true}
	if _, ok := ns.index(-1, 0); ok {
		t.Error("a map wrapping north-south only has a western edge")
	}
	if i, ok := ns.index(2, -1); !ok || i != 5*10+2 {
		t.Errorf("index(2, -1) = %d, %v", i, ok)
	}
	if d := ns.distance(0, 0, 9, 5); d != math.Hypot(9, 1) {
		t.Errorf("north-south distance = %v", d)
	}
	if !ns.latitudeAlongX() || torus.latitudeAlongX() {
		t.Error("latitudes go along x only on maps wrapping north-south only")
	}
	if ns.polar(0, 3) != 0 || ns.polar(5, 0) != 1 || torus.polar(0, 0) != 1 {
		t.Errorf("polar = %v %v %v", ns.polar(0, 3), ns.polar(5, 0), torus.polar(0, 0))
	}

	cylinder := terrainGrid{w: 10, h: 6, wrapX: true}
	if _, ok := cylinder.index(0, 6); ok {
		t.Error("a cylinder has a northern edge")
	}
	if cx, cy, ok := cylinder.corner(10, 6); !ok || cx != 0 || cy != 6 {
		t.Errorf("cylinder corner = %d, %d, %v", cx, cy, ok)
	}
	if _, _, ok := cylinder.corner(3, 7); ok {
		t.Error("a corner beyond the northern edge must not exist")
	}
}

func TestNoiseWrapsNorthSouth(t *testing.T) {
	n := newPeriodicNoise(rand.New(rand.NewSource(1)), 40, 30, 6, true)
	for _, x := range []float64{0, 0.3, 0.71} {
		if a, b := n.at(x, 0), n.at(x, 1); math.Abs(a-b) > 1e-12 {
			t.Errorf("noise at %v: %v at the south, %v at the north", x, a, b)
		}
	}
}

func TestGenerateTerrainOnTorus(t *testing.T) {
	wb := terrainTestMap(t, 64, 40)
	wb.Map.WrapX, wb.Map.WrapY = 1, 1
	result, err := wb.GenerateTerrain(defaultTerrain, nil)
	if err != nil {
		t.Fatal(err)
	}
	g := wb.grid()
	plots, err := wb.plotGrid(g)
	if err != nil {
		t.Fatal(err)
	}
	isLand := func(x, y int) bool {
		i, ok := g.index(x, y)
		return ok && plots[i].PlotType != PlotOcean
	}
	seamLand := 0
	for _, p := range wb.Plots {
		x, y := int(p.X), int(p.Y)
		if len(p.FeatureType) > 0 && p.FeatureType[0] == "FEATURE_ICE" {
			t.Errorf("ice on a torus at %d,%d", x, y)
		}
		if p.PlotType != PlotOcean && (y == 0 || y == 39) {
			seamLand++
		}
		// Rivers and coasts continue across the northern seam
		if p.IsNOfRiver && (p.PlotType == PlotOcean || !isLand(x, y-1)) {
			t.Errorf("river on the southern edge of %d,%d is not between land plots", x, y)
		}
		if p.IsWOfRiver && (p.PlotType == PlotOcean || !isLand(x+1, y)) {
			t.Errorf("river on the eastern edge of %d,%d is not between land plots", x, y)
		}
		if p.TerrainType == "TERRAIN_COAST" && !(isLand(x, y-1) || isLand(x, y+1) || isLand(x-1, y) || isLand(x+1, y) ||
			isLand(x-1, y-1) || isLand(x+1, y+1) || isLand(x-1, y+1) || isLand(x+1, y-1)) {
			t.Errorf("coast without land at %d,%d", x, y)
		}
		if p.TerrainType == "TERRAIN_OCEAN" && y == 0 && (isLand(x, -1) || isLand(x-1, -1) || isLand(x+1, -1)) {
			t.Errorf("ocean next to land across the seam at %d,%d", x, y)
		}
	}
	if seamLand == 0 {
		t.Error("a torus has no poles, land may reach the northern and southern rows")
	}
	if result.Land == 0 || result.Rivers == 0 {
		t.Errorf("result = %+v", result)
	}
}

func TestGenerateTerrainWrappingNorthSouth(t *testing.T) {
	wb := terrainTestMap(t, 40, 64)
	wb.Map.WrapX, wb.Map.WrapY = 0, 1
	if _, err := wb.GenerateTerrain(defaultTerrain, nil); err != nil {
		t.Fatal(err)
	}
	// The poles are the western and eastern edges
	for _, p := range wb.Plots {
		if p.PlotType != PlotOcean && (p.X == 0 || p.X == 39) && p.TerrainType != "TERRAIN_SNOW" {
			t.Errorf("land at the pole %d,%d is %s", p.X, p.Y, p.TerrainType)
		}
	}
}

func TestRegionAcrossNorthernSeam(t *testing.T) {
	app := regionTestApp(t, false)
	app.wbMap.Map.WrapY = 1
	// Rows 3 and 0, 1 of a map wrapping north-south: 1,1 is the third row of the copy
	info, err := app.CopyRegion(Region{X: 1, Y: 3, Width: 1, Height: 3})
	if err != nil || info.Height != 3 || info.Cities != 1 {
		t.Fatal(info, err)
	}
	if _, err := app.PasteRegion(4, 2, false, false, false); err != nil {
		t.Fatal(err)
	}
	// The third row goes to row 4 = 0
	if p := app.GetPlot(4, 0); p.TerrainType != "TERRAIN_GRASS" || !p.IsNOfRiver {
		t.Errorf("paste must wrap north-south: %+v", p)
	}
	removed, err := app.ClearRegion(Region{X: 1, Y: -1, Width: 1, Height: 3}, true, true)
	if err != nil || removed.Units != 1 || removed.Cities != 1 {
		t.Errorf("clearing rows 3, 0 and 1: %+v, %v", removed, err)
	}
}

func TestStartBalanceAcrossNorthernSeam(t *testing.T) {
	wb := terrainTestMap(t, 20, 10)
	wb.Map.WrapX, wb.Map.WrapY = 1, 1
	for _, p := range wb.Plots {
		if p.Y <= 2 || p.Y >= 7 {
			p.PlotType, p.TerrainType = PlotLand, "TERRAIN_GRASS"
		}
	}
	wb.plotAt(5, 0).IsNOfRiver = true // the southern edge of 5,0 is the northern edge of 5,9
	wb.Players[0].CivType, wb.Players[0].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	wb.Players[0].StartingX, wb.Players[0].StartingY = 5, 9
	wb.Players[1].CivType, wb.Players[1].LeaderType = "CIVILIZATION_GREECE", "LEADER_ALEXANDER"
	wb.Players[1].StartingX, wb.Players[1].StartingY = 5, 2

	result := wb.StartBalance()
	if len(result) != 2 {
		t.Fatalf("starts = %+v", result)
	}
	if s := result[0]; s.Land != 21 || s.Water != 0 || !s.River || s.Nearest != 3 {
		t.Errorf("start at the seam: %+v", s)
	}
}
