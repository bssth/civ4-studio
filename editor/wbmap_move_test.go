package editor

import "testing"

func TestMoveAssets(t *testing.T) {
	app := regionTestApp(t, false)
	// 1,1 has a city and a unit; 2,1 is ocean
	if _, err := app.MoveAssets(1, 1, 2, 1, true, true); err == nil {
		t.Fatal("a city can not be moved into water")
	}
	if p := app.GetPlot(1, 1); len(p.Cities) != 1 || len(p.Units) != 1 {
		t.Fatal("a refused move must not change anything")
	}

	target := app.GetPlot(2, 1)
	target.TerrainType, target.PlotType = "TERRAIN_PLAINS", PlotLand
	r, err := app.MoveAssets(1, 1, 2, 1, true, true)
	if err != nil || r.Cities != 1 || r.Units != 1 {
		t.Fatal(r, err)
	}
	if from, to := app.GetPlot(1, 1), app.GetPlot(2, 1); len(from.Cities)+len(from.Units) != 0 || len(to.Cities) != 1 || len(to.Units) != 1 {
		t.Errorf("after the move: %+v / %+v", from, to)
	}
	if h := app.HistoryState(); h.Undo != "move:1,1" {
		t.Errorf("history = %+v", h)
	}
	app.Undo()
	if from, to := app.GetPlot(1, 1), app.GetPlot(2, 1); len(from.Cities) != 1 || len(to.Cities) != 0 || len(to.Units) != 0 {
		t.Error("undo must bring the city and the unit back")
	}
	app.Redo()
	if to := app.GetPlot(2, 1); len(to.Cities) != 1 {
		t.Error("redo must move them again")
	}

	// Only the units, the city stays
	if r, err := app.MoveAssets(2, 1, 3, 1, false, true); err != nil || r.Units != 1 || r.Cities != 0 {
		t.Fatal(r, err)
	}
	if p := app.GetPlot(2, 1); len(p.Cities) != 1 || len(p.Units) != 0 {
		t.Errorf("the city must stay: %+v", p)
	}

	other := app.GetPlot(4, 1)
	other.TerrainType, other.PlotType = "TERRAIN_GRASS", PlotLand
	other.Cities = []*City{{CityName: "Antium"}}
	if _, err := app.MoveAssets(2, 1, 4, 1, true, false); err == nil {
		t.Error("a city can not be moved onto another city")
	}
	peak := app.GetPlot(5, 1)
	peak.PlotType = PlotPeak
	if _, err := app.MoveAssets(2, 1, 5, 1, true, false); err == nil {
		t.Error("a city can not be moved onto a peak")
	}
	for _, c := range [][6]int{{1, 1, 2, 2, 1, 1}, {2, 1, 2, 1, 1, 1}, {2, 1, 9, 9, 1, 1}} {
		if _, err := app.MoveAssets(c[0], c[1], c[2], c[3], c[4] == 1, c[5] == 1); err == nil {
			t.Errorf("move %v must fail", c)
		}
	}
	if _, err := NewApp().MoveAssets(0, 0, 1, 1, true, true); err == nil {
		t.Error("a move without a map must fail")
	}
}
