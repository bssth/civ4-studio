package editor

import (
	"strings"
	"testing"
)

func TestMapView(t *testing.T) {
	wb, err := ParseWbMap(strings.NewReader(cityUnitSignMap))
	if err != nil {
		t.Fatal(err)
	}
	wb.Map = &MapProps{GridWidth: 3, GridHeight: 3, TopLatitude: 90, BottomLatitude: -90}
	wb.Plots[0].IsNOfRiver = true
	wb.Plots[0].FeatureType, wb.Plots[0].FeatureVariety = []string{"FEATURE_FOREST"}, []string{"1"}
	// A plot outside of the grid must be ignored
	wb.Plots = append(wb.Plots, &Plot{X: 10, Y: 10, TerrainType: "TERRAIN_DESERT"})

	view := wb.View()
	if len(view.Terrain) != 9 {
		t.Fatalf("expected 9 cells, got %d", len(view.Terrain))
	}
	i := 2*3 + 1 // x=1, y=2
	if view.Terrain[i] < 0 || view.Terrains[view.Terrain[i]] != "TERRAIN_GRASS" {
		t.Errorf("terrain is not set: %+v", view)
	}
	if view.Features[view.Feature[i]] != "FEATURE_FOREST" || view.PlotType[i] != PlotLand {
		t.Errorf("feature or plot type is wrong: %+v", view)
	}
	if view.Flags[i]&PlotFlagNOfRiver == 0 || view.CityOwner[i] != 0 || view.UnitCount[i] != 1 || view.UnitOwner[i] != 0 {
		t.Errorf("flags, city or units are wrong: %+v", view)
	}
	if view.Terrain[0] != -1 || view.PlotType[0] != PlotOcean || view.CityOwner[0] != -1 {
		t.Errorf("cells without plots must be empty: %+v", view)
	}
	if len(view.Terrains) != 1 {
		t.Errorf("plots outside of the grid must be skipped: %v", view.Terrains)
	}
}

func TestAppSetPlotAndStart(t *testing.T) {
	wb := NewWbMap()
	plots, _ := GenerateOceanPlots(4, 4)
	wb.Map.GridWidth, wb.Map.GridHeight, wb.Plots = 4, 4, plots
	app := NewApp()
	app.wbMap = wb

	plot := *app.GetPlot(2, 3)
	if err := app.SetPlot(&plot); err != nil || app.dirty {
		t.Fatalf("an unchanged plot must not mark the map dirty: %v", err)
	}

	// The frontend always sends a new object
	edited := *app.GetPlot(2, 3)
	edited.TerrainType, edited.PlotType = "TERRAIN_GRASS", PlotLand
	edited.Cities = []*City{{CityName: "Rome", CityOwner: 0, CityPopulation: 3}}
	if err := app.SetPlot(&edited); err != nil || !app.dirty {
		t.Fatalf("plot must be changed: %v", err)
	}
	if got := app.GetPlot(2, 3); got.TerrainType != "TERRAIN_GRASS" || len(got.Cities) != 1 {
		t.Errorf("plot not stored: %+v", got)
	}

	bad := edited
	bad.X = 99
	if err := app.SetPlot(&bad); err == nil {
		t.Error("a plot outside of the map must be rejected")
	}
	bad = edited
	bad.FeatureType = []string{"FEATURE_FOREST"}
	if err := app.SetPlot(&bad); err == nil {
		t.Error("a feature without variety must be rejected")
	}

	wb.Players[1].RandomStartLocation = true
	if err := app.SetPlayerStart(1, 3, 0); err != nil {
		t.Fatal(err)
	}
	if p := wb.Players[1]; p.StartingX != 3 || p.StartingY != 0 || p.RandomStartLocation {
		t.Errorf("start position not moved: %+v", p)
	}
	if err := app.SetPlayerStart(1, 4, 0); err == nil {
		t.Error("a start outside of the map must be rejected")
	}
}
