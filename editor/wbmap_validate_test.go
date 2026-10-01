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
