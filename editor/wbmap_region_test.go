package editor

import (
	"bytes"
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
)

// regionTestApp is an 8x4 ocean map with land, a city, a unit and a river at 1,1
func regionTestApp(t *testing.T, wrapX bool) *App {
	t.Helper()
	wb := NewWbMap()
	plots, err := GenerateOceanPlots(8, 4)
	if err != nil {
		t.Fatal(err)
	}
	wb.Map.GridWidth, wb.Map.GridHeight, wb.Plots = 8, 4, plots
	wb.Map.WrapX = 0
	if wrapX {
		wb.Map.WrapX = 1
	}
	app := NewApp()
	app.wbMap = wb
	p := app.GetPlot(1, 1)
	p.TerrainType, p.PlotType = "TERRAIN_GRASS", PlotHills
	p.FeatureType, p.FeatureVariety = []string{"FEATURE_FOREST"}, []string{"1"}
	p.BonusType, p.IsNOfRiver, p.RiverWEDirection = "BONUS_WHEAT", true, 1
	p.Landmark = "Here"
	p.Cities = []*City{{CityName: "Rome", CityOwner: 0}}
	p.Units = []*Unit{{UnitType: "UNIT_WARRIOR"}}
	return app
}

func TestCopyPasteLand(t *testing.T) {
	app := regionTestApp(t, false)
	info, err := app.CopyRegion(Region{X: 0, Y: 0, Width: 2, Height: 2})
	if err != nil {
		t.Fatal(err)
	}
	if info.Width != 2 || info.Height != 2 || info.Cities != 1 || info.Units != 1 {
		t.Errorf("clipboard = %+v", info)
	}
	// The copy must not change when the source changes
	app.GetPlot(1, 1).TerrainType = "TERRAIN_DESERT"

	n, err := app.PasteRegion(4, 2, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	p := app.GetPlot(5, 3)
	if n != 1 || p.TerrainType != "TERRAIN_GRASS" || p.PlotType != PlotHills || p.BonusType != "BONUS_WHEAT" ||
		!p.IsNOfRiver || p.RiverWEDirection != 1 || len(p.FeatureType) != 1 || p.FeatureVariety[0] != "1" {
		t.Errorf("pasted plot (%d changed) = %+v", n, p)
	}
	if len(p.Cities) != 0 || len(p.Units) != 0 || p.Landmark != "" {
		t.Errorf("land only must not paste cities, units and landmarks: %+v", p)
	}
	if s := app.HistoryState(); s.Undo != "paste:1" {
		t.Errorf("undo label = %q", s.Undo)
	}

	// With assets, partly outside of the map: the top row and the right column are skipped
	if n, err = app.PasteRegion(6, 2, true, false, false); err != nil {
		t.Fatal(err)
	}
	if p := app.GetPlot(7, 3); n != 1 || len(p.Cities) != 1 || len(p.Units) != 1 {
		t.Errorf("pasted with assets (%d changed): %+v", n, p)
	}
	// Changing the pasted city must not change the clipboard
	app.GetPlot(7, 3).Cities[0].CityName = "Changed"
	if info := app.GetClipboard(); info.Cities != 1 {
		t.Error("clipboard lost its city")
	}
	if _, err = app.PasteRegion(0, 0, true, false, false); err != nil {
		t.Fatal(err)
	}
	if c := app.GetPlot(1, 1).Cities; len(c) != 1 || c[0].CityName != "Rome" {
		t.Errorf("clipboard must keep its own copy: %+v", c)
	}

	app.Undo()
	app.Undo()
	app.Undo()
	if p := app.GetPlot(5, 3); p.TerrainType != DefaultOceanTerrain || p.IsNOfRiver {
		t.Errorf("undo must restore the plot: %+v", p)
	}
}

func TestPasteAcrossTheSeam(t *testing.T) {
	app := regionTestApp(t, true)
	// Columns 7, 0 and 1 of a wrapping map
	info, err := app.CopyRegion(Region{X: 7, Y: 1, Width: 3, Height: 1})
	if err != nil || info.Width != 3 {
		t.Fatal(info, err)
	}
	if _, err := app.PasteRegion(6, 3, false, false, false); err != nil {
		t.Fatal(err)
	}
	// 1,1 is the third column of the copy, it goes to column 8 = 0
	if p := app.GetPlot(0, 3); p.TerrainType != "TERRAIN_GRASS" {
		t.Errorf("paste must wrap: %+v", p)
	}

	flat := regionTestApp(t, false)
	if _, err := flat.CopyRegion(Region{X: 7, Y: 1, Width: 3, Height: 1}); err != nil {
		t.Fatal(err)
	}
	if info := flat.GetClipboard(); info.Width != 3 || info.Cities != 0 {
		t.Errorf("columns outside of a flat map are empty: %+v", info)
	}
	if _, err := flat.CopyRegion(Region{X: 0, Y: 0, Width: 9, Height: 1}); err == nil {
		t.Error("a region bigger than the map must be refused")
	}
	if _, err := NewApp().PasteRegion(0, 0, false, false, false); err == nil {
		t.Error("paste without a map must fail")
	}
}

func TestClearRegion(t *testing.T) {
	app := regionTestApp(t, false)
	removed, err := app.ClearRegion(Region{X: 0, Y: 0, Width: 4, Height: 4}, true, false)
	if err != nil {
		t.Fatal(err)
	}
	p := app.GetPlot(1, 1)
	if removed.Units != 1 || removed.Cities != 0 || len(p.Units) != 0 || len(p.Cities) != 1 {
		t.Errorf("removed %+v, plot %+v", removed, p)
	}
	if removed, _ = app.ClearRegion(Region{X: 4, Y: 0, Width: 4, Height: 4}, true, true); removed.Cities != 0 {
		t.Error("nothing to remove there")
	}
	if n := len(app.history.undo); n != 1 {
		t.Errorf("an empty clear is not a step, got %d steps", n)
	}
	app.Undo()
	if len(app.GetPlot(1, 1).Units) != 1 {
		t.Error("undo must bring the unit back")
	}
}

func TestWriteDataURL(t *testing.T) {
	png := []byte{0x89, 'P', 'N', 'G', 1, 2, 3}
	path := filepath.Join(t.TempDir(), "map.png")
	if err := writeDataURL(path, "data:image/png;base64,"+base64.StdEncoding.EncodeToString(png)); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(path); !bytes.Equal(data, png) {
		t.Errorf("written %v", data)
	}
	if err := writeDataURL(path, "data:text/plain;base64,AAAA"); err == nil {
		t.Error("only PNG images can be written")
	}
}

// flipTestClip is a 3x2 clip: 0,0 is grass with a river on its southern edge flowing east,
// 1,1 has a river on its eastern edge flowing south and a city
func flipTestClip() *regionClip {
	c := &regionClip{width: 3, height: 2, plots: make([]*Plot, 6)}
	for x := 0; x < 3; x++ {
		for y := 0; y < 2; y++ {
			c.plots[x*2+y] = &Plot{X: uint(x), Y: uint(y), TerrainType: DefaultOceanTerrain, PlotType: PlotOcean}
		}
	}
	c.plots[0].TerrainType, c.plots[0].IsNOfRiver, c.plots[0].RiverWEDirection = "TERRAIN_GRASS", true, 1
	c.plots[3].IsWOfRiver, c.plots[3].RiverNSDirection = true, 2
	c.plots[3].Cities = []*City{{CityName: "Rome"}}
	return c
}

func TestFlipClip(t *testing.T) {
	c := flipTestClip()
	at := func(clip *regionClip, x, y int) *Plot { return clip.plots[x*clip.height+y] }

	h := c.flipped(true, false)
	if p := at(h, 2, 0); p.TerrainType != "TERRAIN_GRASS" || !p.IsNOfRiver || p.RiverWEDirection != 3 {
		t.Errorf("west-east flip: the southern river must move with its plot and flow west: %+v", p)
	}
	// The eastern edge of 1,1 is the western edge of the mirrored 1,1, i.e. the eastern edge of 0,1
	if p := at(h, 0, 1); !p.IsWOfRiver || p.RiverNSDirection != 2 {
		t.Errorf("west-east flip: the eastern river must move to the plot on the left: %+v", p)
	}
	if p := at(h, 1, 1); p.IsWOfRiver || len(p.Cities) != 1 {
		t.Errorf("west-east flip: the city stays in the middle column, the river leaves it: %+v", p)
	}

	v := c.flipped(false, true)
	// The southern edge of 0,0 becomes the northern edge of 0,1, which is outside of the clip
	if p := at(v, 0, 1); p.TerrainType != "TERRAIN_GRASS" || p.IsNOfRiver {
		t.Errorf("north-south flip: a river on the outer edge is dropped: %+v", p)
	}
	if p := at(v, 1, 0); !p.IsWOfRiver || p.RiverNSDirection != 0 || len(p.Cities) != 1 {
		t.Errorf("north-south flip: the eastern river stays with its plot and flows north: %+v", p)
	}

	// A river on the southern edge of the top row moves to the southern edge of the row below... and back
	c.plots[1].IsNOfRiver, c.plots[1].RiverWEDirection = true, 1
	v = c.flipped(false, true)
	if p := at(v, 0, 1); !p.IsNOfRiver {
		t.Errorf("north-south flip: the river between the rows must stay between them: %+v", p)
	}
	if c.flipped(false, false) != c {
		t.Error("no flip must return the clip itself")
	}
	// Flipping does not change the clipboard
	if !at(c, 0, 0).IsNOfRiver || at(c, 0, 0).RiverWEDirection != 1 {
		t.Error("the original clip changed")
	}
}

func TestPasteFlipped(t *testing.T) {
	app := regionTestApp(t, false)
	if _, err := app.CopyRegion(Region{X: 0, Y: 0, Width: 3, Height: 3}); err != nil {
		t.Fatal(err)
	}
	if _, err := app.PasteRegion(4, 0, true, true, true); err != nil {
		t.Fatal(err)
	}
	// 1,1 is the middle of a 3x3 area, it stays in the middle: 4+1, 0+1
	if p := app.GetPlot(5, 1); p.TerrainType != "TERRAIN_GRASS" || len(p.Cities) != 1 {
		t.Errorf("flipped paste: %+v", p)
	}
	// Its southern river becomes the northern edge, i.e. the southern edge of 5,2, flowing west
	if p := app.GetPlot(5, 2); !p.IsNOfRiver || p.RiverWEDirection != 3 {
		t.Errorf("flipped river: %+v", p)
	}
}

func TestSearch(t *testing.T) {
	app := regionTestApp(t, false)
	m := app.wbMap
	m.Players[0].CivType, m.Players[0].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	m.Players[0].CivDesc, m.Players[0].LeaderName = "Roman Empire", "Caesar"
	m.Players[0].StartingX, m.Players[0].StartingY = 3, 2
	m.Signs = []*Sign{{PlotX: 6, PlotY: 0, PlayerType: -1, Caption: "Road to Rome"}}

	kinds := map[string]SearchResult{}
	for _, r := range app.SearchMap("rom") {
		kinds[r.Kind] = r
	}
	if r := kinds["start"]; r.Label != "Roman Empire (Caesar)" || r.X != 3 || r.Y != 2 || r.Player != 0 {
		t.Errorf("start = %+v", r)
	}
	if r := kinds["city"]; r.Label != "Rome" || r.X != 1 || r.Y != 1 {
		t.Errorf("city = %+v", r)
	}
	if r := kinds["sign"]; r.X != 6 {
		t.Errorf("sign = %+v", r)
	}
	if r := app.SearchMap("HERE"); len(r) != 1 || r[0].Kind != "landmark" {
		t.Errorf("landmark = %+v", r)
	}
	if r := app.SearchMap("  "); len(r) != 0 {
		t.Errorf("empty query = %+v", r)
	}
	for i := 0; i < 80; i++ {
		m.Signs = append(m.Signs, &Sign{PlotX: 0, PlotY: 0, Caption: "x"})
	}
	if n := len(app.SearchMap("x")); n != maxSearchResults {
		t.Errorf("results must be limited, got %d", n)
	}
}
