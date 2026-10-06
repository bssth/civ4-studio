package editor

import (
	"os"
	"strings"
	"testing"
)

func hasProblem(problems []Problem, severity, text string) bool {
	for _, p := range problems {
		if p.Severity == severity && strings.Contains(p.Message, text) {
			return true
		}
	}
	return false
}

func TestValidate(t *testing.T) {
	wb := opsTestMap()
	plots, _ := GenerateOceanPlots(2, 1)
	wb.Plots = append(plots[:0:0], plots...)
	wb.Plots[0].Cities = []*City{{CityName: "Atlantis", CityOwner: 5, CityPopulation: 1}}
	wb.Plots[1].PlotType, wb.Plots[1].TerrainType = PlotLand, "TERRAIN_MUD"
	wb.Plots[1].Units = []*Unit{{UnitType: "UNIT_WARRIOR", UnitOwner: 0}}
	wb.Players[0].StartingX, wb.Players[0].StartingY = 0, 0
	wb.Players[1].StartingX, wb.Players[1].StartingY = 7, 7
	wb.Players[1].Team = 77
	wb.Game.GameTurn, wb.Game.MaxTurns = 10, 5

	data := NewGameData()
	data.Table(InfoTerrains).Set(&TypeInfo{Type: "TERRAIN_OCEAN"})
	data.Table(InfoCivilizations).Set(&TypeInfo{Type: "CIVILIZATION_ROME", Leaders: []string{"LEADER_AUGUSTUS"}})

	problems := wb.Validate(data)
	for _, expected := range []struct{ severity, text string }{
		{SeverityError, "max turns"},
		{SeverityError, "player 0 starts in water"},
		{SeverityError, "start position of player 1 (7, 7) is outside"},
		{SeverityError, "team 77, which does not exist"},
		{SeverityError, "city Atlantis belongs to player 5, which is an empty slot"},
		{SeverityError, "city Atlantis is in water"},
		{SeverityError, "unknown terrain TERRAIN_MUD"},
		{SeverityError, "unknown civilization CIVILIZATION_GREECE"},
		{SeverityWarning, "leader LEADER_CAESAR is not a leader of CIVILIZATION_ROME"},
		{SeverityWarning, "no player is playable"},
	} {
		if !hasProblem(problems, expected.severity, expected.text) {
			t.Errorf("expected %s %q in %+v", expected.severity, expected.text, problems)
		}
	}
	// Categories without loaded data must not be checked
	if hasProblem(problems, SeverityError, "unknown unit") {
		t.Error("units must not be checked without unit data")
	}
	if problems[len(problems)-1].Severity != SeverityWarning {
		t.Error("errors must be listed before warnings")
	}
}

func TestValidateTestMapWithoutGameData(t *testing.T) {
	f, err := os.Open("../" + testFilePath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	wb, err := ParseWbMap(f)
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range wb.Validate(NewGameData()) {
		if p.Severity == SeverityError {
			t.Errorf("unexpected error in the test map: %+v", p)
		}
	}
}

func TestProblemsHaveCodesAndArgs(t *testing.T) {
	wb := opsTestMap()
	plots, _ := GenerateOceanPlots(2, 1)
	wb.Plots = plots
	wb.Plots[1].TerrainType = "TERRAIN_MUD"
	wb.Players[0].StartingX, wb.Players[0].StartingY = 0, 0

	data := NewGameData()
	data.Table(InfoTerrains).Set(&TypeInfo{Type: "TERRAIN_OCEAN"})

	byCode := make(map[string]Problem)
	for _, p := range wb.Validate(data) {
		if p.Code == "" || p.Args == nil {
			t.Errorf("problem without code or args: %+v", p)
		}
		byCode[p.Code] = p
	}
	if p := byCode["players.startWater"]; p.Args["player"] == "" {
		t.Errorf("start in water must have the player argument: %+v", p)
	}
	if p := byCode["unknownType"]; p.Args["what"] != "terrain" || p.Args["value"] != "TERRAIN_MUD" || p.Args["count"] != "1" {
		t.Errorf("unknown type must have what, value and count: %+v", p)
	}
}

func TestValidateSigns(t *testing.T) {
	wb := opsTestMap()
	wb.Signs = []*Sign{
		{PlotX: 1, PlotY: 0, PlayerType: 1, Caption: "Greek sign"},
		{PlotX: 0, PlotY: 0, PlayerType: 5, Caption: "Nobody"},
		{PlotX: 7, PlotY: 0, PlayerType: -1, Caption: "Lost"},
	}
	byCode := make(map[string][]Problem)
	for _, p := range wb.Validate(nil) {
		byCode[p.Code] = append(byCode[p.Code], p)
	}
	if p := byCode["plots.signOutside"]; len(p) != 1 || p[0].Args["caption"] != "Lost" || p[0].Severity != SeverityError {
		t.Errorf("signOutside = %+v", p)
	}
	if p := byCode["plots.signPlayer"]; len(p) != 1 || p[0].Args["player"] != "5" || p[0].X != 0 || p[0].Y != 0 {
		t.Errorf("signPlayer = %+v", p)
	}
}

func TestValidateCityProduction(t *testing.T) {
	wb := opsTestMap()
	wb.Plots[0].Cities[0].ProductionProcess = "PROCESS_MAGIC"
	wb.Plots[1].Cities[0].ProductionUnit = "UNIT_WARRIOR"
	data := NewGameData()
	data.Table(InfoProcesses).Set(&TypeInfo{Type: "PROCESS_WEALTH"})
	data.Table(InfoUnits).Set(&TypeInfo{Type: "UNIT_WARRIOR"})
	data.Table(InfoUnits).Set(&TypeInfo{Type: "UNIT_ARCHER"})
	var found []string
	for _, p := range wb.Validate(data) {
		if p.Code == "unknownType" {
			found = append(found, p.Args["what"]+"="+p.Args["value"])
		}
	}
	if len(found) != 1 || found[0] != "process=PROCESS_MAGIC" {
		t.Errorf("unknown production types = %v", found)
	}
}
