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
BeginPlayer
	CivDesc=Greek Tribe
	CivShortDesc=Greece
	LeaderName=Alexander
	CivAdjective=Greek
	FlagDecal=Art/Interface/TeamColor/FlagDECAL_Helmet.dds
	WhiteFlag=0
	LeaderType=LEADER_ALEXANDER
	CivType=CIVILIZATION_GREECE
	Team=0
	Handicap=HANDICAP_NOBLE
	Color=PLAYERCOLOR_LIGHT_BLUE
	ArtStyle=ARTSTYLE_GRECO_ROMAN
	PlayableCiv=1
	MinorNationStatus=0
	StartingGold=0
	RandomStartLocation=0
	StartingX=180,StartingY=68
	StartingEra=ERA_ANCIENT
	CivicOption=CIVICOPTION_GOVERNMENT,Civic=CIVIC_DESPOTISM
	CivicOption=CIVICOPTION_LEGAL,Civic=CIVIC_BARBARISM
	AttitudePlayer=1,AttitudeExtra=-2
EndPlayer
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

	if player := wb.Players[0]; len(player.Civic) != 2 || player.Civic[1] != "CIVIC_BARBARISM" {
		t.Errorf("civics parsed incorrectly: %+v", player)
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

func TestSameWbFormatIgnoresNilLists(t *testing.T) {
	a := []*Team{{TeamID: 1}}
	b := []*Team{{TeamID: 1, Tech: []string{}, AtWar: []uint{}}}
	if !sameWbFormat(a, b) {
		t.Error("nil and empty lists must be treated as equal")
	}
	b[0].Tech = append(b[0].Tech, "TECH_MINING")
	if sameWbFormat(a, b) {
		t.Error("different techs must be detected")
	}
}

func TestGenerateOceanPlots(t *testing.T) {
	plots, err := GenerateOceanPlots(3, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(plots) != 6 || plots[1].X != 0 || plots[1].Y != 1 || plots[2].X != 1 {
		t.Fatalf("plots must be generated column by column: %+v", plots)
	}

	wb := &WbMap{Version: defaultVersion, Game: &Game{}, Map: &MapProps{GridWidth: 3, GridHeight: 2, TopLatitude: 90, BottomLatitude: -90}, Plots: plots}
	stats := wb.Stats()
	if stats.Water != 6 || len(stats.Problems) != 0 {
		t.Errorf("unexpected stats: %+v", stats)
	}
	if !strings.Contains(string(wb.ToWbFormat()), "num plots written=6") {
		t.Error("number of plots must be written from actual plots")
	}

	if _, err := GenerateOceanPlots(0, 10); err == nil {
		t.Error("empty map size must be rejected")
	}
}

func TestStatsReportsProblems(t *testing.T) {
	wb := &WbMap{Game: &Game{}, Map: &MapProps{GridWidth: 2, GridHeight: 2, TopLatitude: 10, BottomLatitude: 20},
		Plots: []*Plot{{X: 5, Y: 0}}}
	if problems := wb.Stats().Problems; len(problems) != 3 {
		t.Errorf("expected plot count, outside plot and latitude problems, got %v", problems)
	}
}

func TestAppMapProps(t *testing.T) {
	app := NewApp()
	app.wbMap = &WbMap{Game: &Game{}}
	props := app.GetMapProps()
	if props == nil || props.TopLatitude != 90 {
		t.Fatalf("default map properties expected: %+v", props)
	}

	if err := app.CreatePlots(4, 3); err != nil {
		t.Fatal(err)
	}
	if !app.dirty || len(app.wbMap.Plots) != 12 || app.wbMap.Map.GridWidth != 4 {
		t.Error("plots must be created and the map marked dirty")
	}
	if err := app.CreatePlots(4, 3); err == nil {
		t.Error("plots must not be created twice")
	}

	changed := *app.wbMap.Map
	changed.GridWidth = 10
	if err := app.SetMapProps(&changed); err == nil {
		t.Error("grid size of a map with plots must not change")
	}
	changed = *app.wbMap.Map
	changed.Climate = "CLIMATE_ARID"
	if err := app.SetMapProps(&changed); err != nil || app.wbMap.Map.Climate != "CLIMATE_ARID" {
		t.Errorf("climate must be changed: %v", err)
	}
	changed.TopLatitude = -95
	if err := app.SetMapProps(&changed); err == nil {
		t.Error("invalid latitudes must be rejected")
	}
}

func TestNewWbMapRoundTrip(t *testing.T) {
	wb := NewWbMap()
	plots, err := GenerateOceanPlots(8, 4)
	if err != nil {
		t.Fatal(err)
	}
	wb.Map.GridWidth, wb.Map.GridHeight, wb.Plots = 8, 4, plots

	parsed := assertRoundTrip(t, string(wb.ToWbFormat()))
	if len(parsed.Players) != DefaultPlayerSlots || len(parsed.Teams) != DefaultPlayerSlots || len(parsed.Plots) != 32 {
		t.Errorf("unexpected new map contents: %d players, %d teams, %d plots", len(parsed.Players), len(parsed.Teams), len(parsed.Plots))
	}
	if problems := parsed.Stats().Problems; len(problems) != 0 {
		t.Errorf("a new map must have no problems: %v", problems)
	}
}
