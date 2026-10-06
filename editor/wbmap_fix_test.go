package editor

import "testing"

// fixTestMap is a 6x4 map: grass land in columns 0..2, ocean in 3..5, with a few broken things
func fixTestMap(t *testing.T) *WbMap {
	t.Helper()
	wb := terrainTestMap(t, 6, 4)
	for _, p := range wb.Plots {
		if p.X <= 2 {
			p.PlotType, p.TerrainType = PlotLand, "TERRAIN_GRASS"
		}
	}
	wb.Map.WrapX = 0
	wb.Players[0].CivType, wb.Players[0].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	wb.Players[1].CivType, wb.Players[1].LeaderType = "CIVILIZATION_GREECE", "LEADER_ALEXANDER"
	wb.Players[0].StartingX, wb.Players[0].StartingY = 4, 1 // in water
	wb.Players[1].StartingX, wb.Players[1].StartingY = 2, 1
	wb.Game.MaxTurns, wb.Game.GameTurn = 5, 10
	wb.Teams[0].AtWar = []uint{1, 99}
	wb.plotAt(1, 1).Cities = []*City{{CityName: "Rome", CityOwner: 0}}                     // no population
	wb.plotAt(0, 0).Cities = []*City{{CityName: "Ghost", CityOwner: 5, CityPopulation: 1}} // empty slot
	wb.plotAt(5, 3).Cities = []*City{{CityName: "Atlantis", CityOwner: 1, CityPopulation: 2}}
	wb.plotAt(0, 2).Units = []*Unit{{UnitType: "UNIT_WARRIOR", UnitOwner: 7}, {UnitType: "UNIT_WARRIOR", UnitOwner: 7}, {UnitType: "UNIT_ARCHER", UnitOwner: 1}}
	wb.Signs = []*Sign{{PlotX: 9, PlotY: 9, PlayerType: -1, Caption: "Far"}, {PlotX: 1, PlotY: 0, PlayerType: 6, Caption: "Hidden"}}
	return wb
}

func TestFixAllProblems(t *testing.T) {
	app := NewApp()
	app.wbMap = fixTestMap(t)
	before := app.wbMap.ToWbFormat()
	problems := app.ValidateMap()
	fixable := 0
	for _, p := range problems {
		if p.Fix != "" && p.Fix != "replace" {
			fixable++
		}
	}
	n, err := app.FixProblems(problems)
	if err != nil {
		t.Fatal(err)
	}
	// Two units of the empty slot give two problems, one fix removes both
	if n != fixable-1 {
		t.Errorf("fixed %d of %d fixable problems", n, fixable)
	}
	for _, p := range app.ValidateMap() {
		if p.Fix != "" && p.Fix != "replace" {
			t.Errorf("not fixed: %s (%s)", p.Code, p.Message)
		}
	}
	m := app.wbMap
	if p := m.Players[0]; m.plotAt(p.StartingX, p.StartingY).PlotType != PlotLand || (p.StartingX == 2 && p.StartingY == 1) {
		t.Errorf("start must move to free land: %d,%d", p.StartingX, p.StartingY)
	}
	if m.Game.MaxTurns != 0 || len(m.Teams[0].AtWar) != 1 || m.Teams[0].AtWar[0] != 1 {
		t.Errorf("game %d, war %v", m.Game.MaxTurns, m.Teams[0].AtWar)
	}
	if m.plotAt(1, 1).Cities[0].CityPopulation != 1 || len(m.plotAt(0, 0).Cities) != 0 {
		t.Error("cities must be fixed")
	}
	if p := m.plotAt(5, 3); p.PlotType != PlotLand || len(p.Cities) != 1 {
		t.Errorf("a city in water keeps its plot as land: %+v", p)
	}
	if u := m.plotAt(0, 2).Units; len(u) != 1 || u[0].UnitOwner != 1 {
		t.Errorf("units = %+v", u)
	}
	if len(m.Signs) != 1 || m.Signs[0].PlayerType != -1 {
		t.Errorf("signs = %+v", m.Signs)
	}
	if s := app.HistoryState(); s.Undo == "" {
		t.Error("fixes must be undoable")
	}
	app.Undo()
	if string(before) != string(app.wbMap.ToWbFormat()) {
		t.Error("undo must restore the map")
	}
	if n, _ := app.FixProblems(nil); n != 0 || app.HistoryState().Undo != "" {
		t.Error("nothing to fix is not a step")
	}
}

func TestTwoCitiesFix(t *testing.T) {
	wb := fixTestMap(t)
	wb.plotAt(2, 2).Cities = []*City{{CityName: "A", CityPopulation: 1}, {CityName: "B", CityPopulation: 1}}
	for _, p := range wb.Validate(nil) {
		if p.Code == "plots.twoCities" && !wb.Fix(p) {
			t.Error("fix failed")
		}
	}
	if c := wb.plotAt(2, 2).Cities; len(c) != 1 || c[0].CityName != "A" {
		t.Errorf("cities = %+v", c)
	}
}

func TestReplaceType(t *testing.T) {
	app := NewApp()
	app.wbMap = fixTestMap(t)
	m := app.wbMap
	m.plotAt(0, 0).BonusType = "BONUS_MITHRIL"
	m.plotAt(1, 0).BonusType = "BONUS_MITHRIL"
	m.plotAt(1, 1).FeatureType, m.plotAt(1, 1).FeatureVariety = []string{"FEATURE_MAGIC"}, []string{"2"}
	m.plotAt(1, 1).Cities[0].BuildingType = []string{"BUILDING_TOWER", "BUILDING_WALLS"}
	m.plotAt(1, 1).Cities[0].ProductionBuilding = "BUILDING_TOWER"
	m.Teams[0].Tech = []string{"TECH_MAGIC", "TECH_MINING"}

	if n, err := app.ReplaceType("resource", "BONUS_MITHRIL", "BONUS_IRON"); err != nil || n != 2 {
		t.Fatalf("replaced %d: %v", n, err)
	}
	if m.plotAt(1, 0).BonusType != "BONUS_IRON" {
		t.Error("resource not replaced")
	}
	if n, _ := app.ReplaceType("feature", "FEATURE_MAGIC", ""); n != 1 || len(m.plotAt(1, 1).FeatureType) != 0 || len(m.plotAt(1, 1).FeatureVariety) != 0 {
		t.Error("feature must be removed with its variety")
	}
	if n, _ := app.ReplaceType("building", "BUILDING_TOWER", ""); n != 2 {
		t.Errorf("building and production: %d", n)
	}
	if c := m.plotAt(1, 1).Cities[0]; len(c.BuildingType) != 1 || c.ProductionBuilding != "" {
		t.Errorf("city = %+v", c)
	}
	if n, _ := app.ReplaceType("tech", "TECH_MAGIC", ""); n != 1 || len(m.Teams[0].Tech) != 1 {
		t.Errorf("techs = %v", m.Teams[0].Tech)
	}
	if n, _ := app.ReplaceType("unit", "UNIT_WARRIOR", ""); n != 2 || len(m.plotAt(0, 2).Units) != 1 {
		t.Errorf("units = %+v", m.plotAt(0, 2).Units)
	}
	if _, err := app.ReplaceType("civilization", "CIVILIZATION_ROME", ""); err == nil {
		t.Error("a civilization can not be removed")
	}
	if n, _ := app.ReplaceType("civilization", "CIVILIZATION_ROME", "CIVILIZATION_CARTHAGE"); n != 1 || m.Players[0].CivType != "CIVILIZATION_CARTHAGE" {
		t.Error("civilization not replaced")
	}
	if n, err := app.ReplaceType("resource", "BONUS_NOTHING", "BONUS_IRON"); n != 0 || err != nil {
		t.Error("nothing to replace")
	}
	steps := len(app.history.undo)
	if steps != 6 {
		t.Errorf("each replacement is a step, got %d", steps)
	}
	for i := 0; i < steps; i++ {
		app.Undo()
	}
	if m := app.wbMap; m.plotAt(0, 0).BonusType != "BONUS_MITHRIL" || len(m.plotAt(0, 2).Units) != 3 || m.Players[0].CivType != "CIVILIZATION_ROME" {
		t.Error("undo must restore everything")
	}
}
