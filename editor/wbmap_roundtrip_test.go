package editor

import (
	"os"
	"strings"
	"testing"
)

// normalizeWb strips indentation and empty lines, so files can be compared regardless of formatting
func normalizeWb(data string) []string {
	var result []string
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}
	return result
}

func assertRoundTrip(t *testing.T, source string) *WbMap {
	t.Helper()

	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	expected := normalizeWb(source)
	actual := normalizeWb(string(wb.ToWbFormat()))
	for i := 0; i < len(expected) && i < len(actual); i++ {
		if expected[i] != actual[i] {
			t.Fatalf("line %d differs after round trip:\nexpected: %s\nactual:   %s", i+1, expected[i], actual[i])
		}
	}
	if len(expected) != len(actual) {
		t.Fatalf("line count differs after round trip: expected %d, got %d", len(expected), len(actual))
	}

	return wb
}

func TestRoundTripTestMap(t *testing.T) {
	data, err := os.ReadFile("../" + testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	assertRoundTrip(t, string(data))
}

const cityUnitSignMap = `Version=11
BeginGame
	Era=ERA_ANCIENT
	Victory=VICTORY_TIME
	GameTurn=0
	MaxCityElimination=0
	NumAdvancedStartPoints=0
	TargetScore=0
	StartYear=-4000
	Description=A map, with commas, in description
	Tutorial=0
	MaxTurns=0
EndGame
BeginPlot
	x=1,y=2
	StartingPlot=0
	TerrainType=TERRAIN_GRASS
	PlotType=2
	SomeModKey=42
	BeginUnit
		UnitType=UNIT_WARRIOR,UnitOwner=0
		Level=2,Experience=5
		PromotionType=PROMOTION_COMBAT1
		PromotionType=PROMOTION_COMBAT2
		Damage=0
		FacingDirection=4
	EndUnit
	BeginCity
		CityOwner=0
		CityName=Rome, the Eternal
		CityPopulation=3
		BuildingType=BUILDING_PALACE
		BuildingType=BUILDING_BARRACKS
		ReligionType=RELIGION_JUDAISM
		HolyCityReligionType=RELIGION_JUDAISM
		Player0Culture=10
		Player3Culture=7
	EndCity
	TeamReveal=0,1,2,
EndPlot
BeginSign
	plotX=1
	plotY=2
	playerType=-1
	caption=Hello, world
EndSign
`

func TestRoundTripCitiesUnitsSigns(t *testing.T) {
	wb := assertRoundTrip(t, cityUnitSignMap)

	if wb.Game.Description != "A map, with commas, in description" {
		t.Errorf("description with commas was cut: %q", wb.Game.Description)
	}

	plot := wb.Plots[0]
	if len(plot.TeamReveal) != 3 {
		t.Errorf("expected 3 revealed teams, got %v", plot.TeamReveal)
	}

	city := plot.Cities[0]
	if len(city.BuildingType) != 2 || city.PlayerCulture[3] != 7 {
		t.Errorf("city parsed incorrectly: %+v", city)
	}

	if len(plot.Units[0].PromotionType) != 2 {
		t.Errorf("unit promotions parsed incorrectly: %v", plot.Units[0].PromotionType)
	}

	if len(wb.Signs) != 1 || wb.Signs[0].Caption != "Hello, world" {
		t.Errorf("sign parsed incorrectly: %+v", wb.Signs)
	}
}

func TestParseUnclosedSection(t *testing.T) {
	_, err := ParseWbMap(strings.NewReader("Version=11\nBeginGame\nEra=ERA_ANCIENT\n"))
	if err == nil {
		t.Fatal("expected error for unclosed section")
	}
}

func TestNewMapCanBeSaved(t *testing.T) {
	wb := &WbMap{Version: defaultVersion, Game: &Game{}}
	if len(wb.ToWbFormat()) == 0 {
		t.Fatal("empty output for a map without map properties")
	}
}
