package editor

import (
	"strings"
	"testing"
)

func paintTestApp(t *testing.T) *App {
	t.Helper()
	wb := NewWbMap()
	plots, err := GenerateOceanPlots(4, 4)
	if err != nil {
		t.Fatal(err)
	}
	wb.Map.GridWidth, wb.Map.GridHeight, wb.Plots = 4, 4, plots
	app := NewApp()
	app.wbMap = wb
	return app
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func TestPaintPlotsAndUndo(t *testing.T) {
	app := paintTestApp(t)

	n, err := app.PaintPlots(&PaintOp{
		Cells:   []PlotXY{{1, 1}, {1, 2}, {1, 1}, {9, 9}},
		Terrain: strPtr("TERRAIN_GRASS"),
		Feature: strPtr("FEATURE_FOREST"), FeatureVariety: 2,
	})
	if err != nil || n != 2 {
		t.Fatalf("expected 2 painted plots, got %d (%v)", n, err)
	}
	p := app.GetPlot(1, 1)
	if p.TerrainType != "TERRAIN_GRASS" || p.PlotType != PlotLand || p.FeatureType[0] != "FEATURE_FOREST" || p.FeatureVariety[0] != "2" {
		t.Errorf("land terrain must make the plot flat and get the feature: %+v", p)
	}
	if !app.dirty || app.HistoryState().Undo != "paint:2" {
		t.Errorf("paint must be recorded: %+v", app.HistoryState())
	}

	// Painting the same again changes nothing and is not recorded
	if n, _ := app.PaintPlots(&PaintOp{Cells: []PlotXY{{1, 1}}, Terrain: strPtr("TERRAIN_GRASS")}); n != 0 {
		t.Errorf("repainting the same must change nothing, got %d", n)
	}

	if _, err := app.PaintPlots(&PaintOp{Cells: []PlotXY{{1, 1}}, PlotType: intPtr(PlotHills), Feature: strPtr("")}); err != nil {
		t.Fatal(err)
	}
	if p := app.GetPlot(1, 1); p.PlotType != PlotHills || len(p.FeatureType) != 0 {
		t.Errorf("height and feature removal not applied: %+v", p)
	}

	// Undo both steps, then redo one
	for i := 0; i < 2; i++ {
		if _, err := app.Undo(); err != nil {
			t.Fatal(err)
		}
	}
	if p := app.GetPlot(1, 1); p.TerrainType != DefaultOceanTerrain || p.PlotType != PlotOcean || len(p.FeatureType) != 0 {
		t.Errorf("undo must restore the ocean: %+v", p)
	}
	if _, err := app.Undo(); err == nil {
		t.Error("undo with empty history must fail")
	}
	state, err := app.Redo()
	if err != nil || state.Redo == "" || state.Undo != "paint:2" {
		t.Errorf("unexpected state after redo: %+v (%v)", state, err)
	}
	if p := app.GetPlot(1, 2); p.TerrainType != "TERRAIN_GRASS" {
		t.Errorf("redo must paint again: %+v", p)
	}

	// A new edit drops the redo stack
	if _, err := app.PaintPlots(&PaintOp{Cells: []PlotXY{{0, 0}}, Bonus: strPtr("BONUS_FISH")}); err != nil {
		t.Fatal(err)
	}
	if app.HistoryState().Redo != "" {
		t.Error("a new edit must clear redo")
	}
}

func TestPaintWaterTerrainUsesGameData(t *testing.T) {
	app := paintTestApp(t)
	data := NewGameData()
	data.Table(InfoTerrains).Set(&TypeInfo{Type: "TERRAIN_SWAMP_WATER", Water: true})
	SetGameData(data)
	t.Cleanup(func() { SetGameData(NewGameData()) })

	app.PaintPlots(&PaintOp{Cells: []PlotXY{{2, 2}}, Terrain: strPtr("TERRAIN_GRASS")})
	app.PaintPlots(&PaintOp{Cells: []PlotXY{{2, 2}}, Terrain: strPtr("TERRAIN_SWAMP_WATER")})
	if p := app.GetPlot(2, 2); p.PlotType != PlotOcean {
		t.Errorf("water terrain from game data must make the plot water: %+v", p)
	}
}

func TestUndoPlotEditAndStart(t *testing.T) {
	app := paintTestApp(t)
	edited := *app.GetPlot(3, 3)
	edited.Cities = []*City{{CityName: "Rome", CityPopulation: 2}}
	if err := app.SetPlot(&edited); err != nil {
		t.Fatal(err)
	}
	if err := app.SetPlayerStart(0, 3, 3); err != nil {
		t.Fatal(err)
	}

	app.Undo()
	if p := app.wbMap.Players[0]; p.StartingX != 0 || p.StartingY != 0 {
		t.Errorf("start must be moved back: %+v", p)
	}
	app.Undo()
	if len(app.GetPlot(3, 3).Cities) != 0 {
		t.Error("the city must be removed by undo")
	}
	app.Redo()
	if c := app.GetPlot(3, 3).Cities; len(c) != 1 || c[0].CityName != "Rome" {
		t.Errorf("redo must restore the city: %+v", c)
	}

	// History states must not change when the current plot is edited in place
	app.GetPlot(3, 3).Cities[0].CityName = "Changed"
	app.Undo()
	app.Redo()
	if c := app.GetPlot(3, 3).Cities[0]; c.CityName != "Rome" {
		t.Errorf("history must keep its own copies, got %q", c.CityName)
	}
}

func TestHistoryIsLimitedAndResetOnNewMap(t *testing.T) {
	app := paintTestApp(t)
	for i := 0; i < maxHistory+20; i++ {
		terrain := "TERRAIN_GRASS"
		if i%2 == 1 {
			terrain = "TERRAIN_PLAINS"
		}
		app.PaintPlots(&PaintOp{Cells: []PlotXY{{0, 0}}, Terrain: &terrain})
	}
	if len(app.history.undo) != maxHistory {
		t.Errorf("history must be limited to %d, got %d", maxHistory, len(app.history.undo))
	}
	app.mu.Lock()
	app.history.reset()
	app.mu.Unlock()
	if s := app.HistoryState(); s.Undo != "" || s.Redo != "" {
		t.Errorf("history must be empty: %+v", s)
	}
}

func TestPaintReveal(t *testing.T) {
	app := paintTestApp(t)
	team := 2
	op := &PaintOp{Cells: []PlotXY{{0, 0}, {1, 0}}, RevealTeam: &team, Reveal: true}
	if n, err := app.PaintPlots(op); err != nil || n != 2 {
		t.Fatalf("revealed %d: %v", n, err)
	}
	app.GetPlot(1, 0).TeamReveal = []uint{2, 5}
	if got := app.GetRevealed(2); got[:4] != "1100" || len(got) != 16 {
		t.Errorf("revealed = %q", got)
	}
	// Revealing again changes nothing, hiding removes only that team
	if n, _ := app.PaintPlots(op); n != 0 {
		t.Errorf("already revealed, changed %d", n)
	}
	op.Reveal = false
	if n, _ := app.PaintPlots(op); n != 2 {
		t.Errorf("hidden %d", n)
	}
	if r := app.GetPlot(1, 0).TeamReveal; len(r) != 1 || r[0] != 5 {
		t.Errorf("team reveal = %v", r)
	}
	app.wbMap.Teams[2].RevealMap = true
	if got := app.GetRevealed(2); got != strings.Repeat("1", 16) {
		t.Errorf("a team with the revealed map sees everything: %q", got)
	}
	bad := -1
	if _, err := app.PaintPlots(&PaintOp{Cells: []PlotXY{{0, 0}}, RevealTeam: &bad, Reveal: true}); err == nil {
		t.Error("a negative team must be refused")
	}
}
