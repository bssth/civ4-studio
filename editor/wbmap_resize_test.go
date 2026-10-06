package editor

import (
	"testing"
)

func TestResizeKeepsContentInPlace(t *testing.T) {
	app := paintTestApp(t) // 4x4 ocean map
	m := app.wbMap
	app.GetPlot(1, 2).TerrainType, app.GetPlot(1, 2).PlotType = "TERRAIN_GRASS", PlotLand
	m.Players[0].CivType, m.Players[0].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	m.Players[0].StartingX, m.Players[0].StartingY = 1, 2
	app.GetPlot(1, 2).Cities = []*City{{CityName: "Rome", CityOwner: 0}}
	app.GetPlot(3, 0).Units = []*Unit{{UnitType: "UNIT_WARRIOR"}}
	m.Signs = []*Sign{{PlotX: 1, PlotY: 2, PlayerType: -1, Caption: "Rome"}, {PlotX: 3, PlotY: 3, PlayerType: -1, Caption: "Corner"}}

	// Two ocean columns on the west, one row more on the south, one column less on the east
	result, err := app.ResizeMap(2, -1, 0, 1)
	if err != nil {
		t.Fatal(err)
	}
	if m.Map.GridWidth != 5 || m.Map.GridHeight != 5 || len(m.Plots) != 25 {
		t.Fatalf("size %dx%d with %d plots", m.Map.GridWidth, m.Map.GridHeight, len(m.Plots))
	}
	for i, p := range m.Plots {
		if int(p.X) != i/5 || int(p.Y) != i%5 {
			t.Fatalf("plot %d is %d,%d, plots must be written column by column", i, p.X, p.Y)
		}
	}
	if p := app.GetPlot(3, 3); len(p.Cities) != 1 || p.TerrainType != "TERRAIN_GRASS" {
		t.Errorf("the city must move with its plot: %+v", p)
	}
	if p := app.GetPlot(0, 0); p.TerrainType != DefaultOceanTerrain || p.PlotType != PlotOcean {
		t.Errorf("new plots must be ocean: %+v", p)
	}
	if p := m.Players[0]; p.StartingX != 3 || p.StartingY != 3 {
		t.Errorf("start must move: %d,%d", p.StartingX, p.StartingY)
	}
	if result.Units != 1 || result.Cities != 0 || result.Signs != 1 || len(result.Starts) != 0 {
		t.Errorf("result = %+v", result)
	}
	if len(m.Signs) != 1 || m.Signs[0].PlotX != 3 || m.Signs[0].PlotY != 3 {
		t.Errorf("signs = %+v", m.Signs)
	}
	if s := app.HistoryState(); s.Undo != "resize:5,5" {
		t.Errorf("undo label = %q", s.Undo)
	}

	app.Undo()
	if m := app.wbMap; m.Map.GridWidth != 4 || len(m.Plots) != 16 || len(m.Signs) != 2 || len(app.GetPlot(3, 0).Units) != 1 {
		t.Errorf("undo must restore the map: %dx%d, %d plots, %d signs", m.Map.GridWidth, m.Map.GridHeight, len(m.Plots), len(m.Signs))
	}
	if p := app.wbMap.Players[0]; p.StartingX != 1 || p.StartingY != 2 {
		t.Errorf("undo must restore the start: %d,%d", p.StartingX, p.StartingY)
	}
}

func TestResizeReportsStartsOutsideAndLimits(t *testing.T) {
	app := paintTestApp(t)
	m := app.wbMap
	m.Players[1].CivType, m.Players[1].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	m.Players[1].StartingX, m.Players[1].StartingY = 3, 3
	if _, err := app.ResizeMap(0, 0, -1, -0); err == nil {
		t.Error("a map smaller than the minimum size must be refused")
	}
	if m.Map.GridHeight != 4 || len(app.history.undo) != 0 {
		t.Error("a refused resize must change nothing")
	}
	result, err := app.ResizeMap(0, 0, 0, 0)
	if err != nil || len(app.history.undo) != 0 || result == nil {
		t.Errorf("an empty resize is not a step: %v", err)
	}
	if _, err := app.ResizeMap(-1, 0, 0, 0); err == nil {
		t.Error("3 columns are less than the minimum")
	}
	if _, err := app.ResizeMap(0, 4, 0, 4); err != nil {
		t.Fatal(err)
	}
	if p := m.Players[1]; p.StartingX != 3 || p.StartingY != 7 {
		t.Fatalf("start = %d,%d", p.StartingX, p.StartingY)
	}
	// Removing the north half leaves the start outside
	result, err = app.ResizeMap(0, 0, -4, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Starts) != 1 || result.Starts[0] != 1 {
		t.Errorf("player 1 must be reported: %+v", result)
	}
}

func TestSetPlotSigns(t *testing.T) {
	app := paintTestApp(t)
	app.wbMap.Signs = []*Sign{
		{PlotX: 0, PlotY: 0, PlayerType: -1, Caption: "A"},
		{PlotX: 1, PlotY: 1, PlayerType: -1, Caption: "B"},
		{PlotX: 2, PlotY: 2, PlayerType: -1, Caption: "C"},
	}
	if err := app.SetPlotSigns(1, 1, []*Sign{{PlayerType: 0, Caption: "B1"}, {PlayerType: -1, Caption: "B2"}}); err != nil {
		t.Fatal(err)
	}
	var captions []string
	for _, s := range app.GetSigns() {
		captions = append(captions, s.Caption)
	}
	if got := captions; len(got) != 4 || got[0] != "A" || got[1] != "B1" || got[2] != "B2" || got[3] != "C" {
		t.Errorf("signs = %v", got)
	}
	if s := app.GetSigns()[1]; s.PlotX != 1 || s.PlotY != 1 {
		t.Errorf("new signs must get the plot coordinates: %+v", s)
	}
	if s := app.HistoryState(); s.Undo != "signs:1,1" {
		t.Errorf("undo label = %q", s.Undo)
	}
	if err := app.SetPlotSigns(1, 1, []*Sign{{Caption: "two\nlines"}}); err == nil {
		t.Error("a caption with a line break must be refused")
	}
	if err := app.SetPlotSigns(9, 9, nil); err == nil {
		t.Error("a plot outside of the map must be refused")
	}
	// Removing all signs of a plot and adding to a plot without signs
	if err := app.SetPlotSigns(0, 0, nil); err != nil {
		t.Fatal(err)
	}
	if err := app.SetPlotSigns(3, 3, []*Sign{{PlayerType: -1, Caption: "D"}}); err != nil {
		t.Fatal(err)
	}
	if n := len(app.GetSigns()); n != 4 {
		t.Errorf("expected 4 signs, got %d", n)
	}
	app.Undo()
	app.Undo()
	app.Undo()
	if n := len(app.GetSigns()); n != 3 || app.GetSigns()[1].Caption != "B" {
		t.Errorf("undo must restore the signs: %+v", app.GetSigns())
	}
}
