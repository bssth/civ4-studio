package editor

import "testing"

func opsTestMap() *WbMap {
	wb := NewWbMap()
	wb.Map.GridWidth, wb.Map.GridHeight = 2, 1
	wb.Players[0] = &Player{CivType: "CIVILIZATION_ROME", LeaderType: "LEADER_CAESAR", Team: 0,
		AttitudePlayer: []uint{1}, AttitudeExtra: []int{-3}}
	wb.Players[1] = &Player{CivType: "CIVILIZATION_GREECE", LeaderType: "LEADER_ALEXANDER", Team: 1,
		AttitudePlayer: []uint{0}, AttitudeExtra: []int{2}}
	wb.Plots = []*Plot{
		{X: 0, Y: 0, Units: []*Unit{{UnitType: "UNIT_WARRIOR", UnitOwner: 0}, {UnitType: "UNIT_ARCHER", UnitOwner: 1}},
			Cities: []*City{{CityName: "Rome", CityOwner: 0, PlayerCulture: map[uint]uint64{0: 10, 1: 3}}}},
		{X: 1, Y: 0, Cities: []*City{{CityName: "Athens", CityOwner: 1, PlayerCulture: map[uint]uint64{1: 5}}}},
	}
	wb.Signs = []*Sign{{PlayerType: 1, Caption: "Greek sign"}, {PlayerType: -1, Caption: "Public"}}
	return wb
}

func TestSwapPlayers(t *testing.T) {
	wb := opsTestMap()
	if err := wb.SwapPlayers(0, 1); err != nil {
		t.Fatal(err)
	}
	if wb.Players[0].CivType != "CIVILIZATION_GREECE" || wb.Players[1].CivType != "CIVILIZATION_ROME" {
		t.Fatalf("players not swapped")
	}
	rome, athens := wb.Plots[0].Cities[0], wb.Plots[1].Cities[0]
	if rome.CityOwner != 1 || athens.CityOwner != 0 {
		t.Errorf("city owners not remapped: %d %d", rome.CityOwner, athens.CityOwner)
	}
	if rome.PlayerCulture[1] != 10 || rome.PlayerCulture[0] != 3 || athens.PlayerCulture[0] != 5 {
		t.Errorf("culture not remapped: %v %v", rome.PlayerCulture, athens.PlayerCulture)
	}
	if wb.Plots[0].Units[0].UnitOwner != 1 || wb.Plots[0].Units[1].UnitOwner != 0 {
		t.Errorf("unit owners not remapped")
	}
	if wb.Players[1].AttitudePlayer[0] != 0 || wb.Players[0].AttitudePlayer[0] != 1 {
		t.Errorf("attitudes not remapped: %v %v", wb.Players[0].AttitudePlayer, wb.Players[1].AttitudePlayer)
	}
	if wb.Signs[0].PlayerType != 0 || wb.Signs[1].PlayerType != -1 {
		t.Errorf("signs not remapped: %+v", wb.Signs)
	}
	if err := wb.SwapPlayers(0, 99); err == nil {
		t.Error("invalid player index must be rejected")
	}
}

func TestClearPlayer(t *testing.T) {
	wb := opsTestMap()
	units, cities, err := wb.ClearPlayer(1, true)
	if err != nil {
		t.Fatal(err)
	}
	if units != 1 || cities != 1 {
		t.Errorf("expected 1 unit and 1 city removed, got %d and %d", units, cities)
	}
	if p := wb.Players[1]; p.CivType != NonePlayer || p.Team != 1 {
		t.Errorf("slot not cleared or team lost: %+v", p)
	}
	if len(wb.Plots[0].Units) != 1 || len(wb.Plots[1].Cities) != 0 {
		t.Errorf("assets not removed")
	}
	if _, ok := wb.Plots[0].Cities[0].PlayerCulture[1]; ok {
		t.Error("culture of the removed player must be removed")
	}
	if len(wb.Players[0].AttitudePlayer) != 0 || len(wb.Players[0].AttitudeExtra) != 0 {
		t.Errorf("attitude to the removed player must be removed: %+v", wb.Players[0])
	}

	wb = opsTestMap()
	if _, _, err := wb.ClearPlayer(0, false); err != nil {
		t.Fatal(err)
	}
	if len(wb.Plots[0].Units) != 2 || len(wb.Plots[0].Cities) != 1 {
		t.Error("assets must stay when not asked to remove them")
	}
}
