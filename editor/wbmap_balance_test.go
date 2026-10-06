package editor

import "testing"

func TestStartBalance(t *testing.T) {
	wb := terrainTestMap(t, 20, 10)
	wb.Map.WrapX = 1
	for _, p := range wb.Plots {
		if p.X >= 2 && p.X <= 8 {
			p.PlotType, p.TerrainType = PlotLand, "TERRAIN_GRASS"
		}
	}
	wb.plotAt(5, 5).IsWOfRiver = false
	wb.plotAt(4, 5).IsWOfRiver = true // the western edge of 5,5
	wb.plotAt(6, 5).PlotType = PlotHills
	wb.plotAt(6, 6).FeatureType = []string{"FEATURE_FOREST"}
	wb.plotAt(5, 7).BonusType = "BONUS_WHEAT"
	wb.plotAt(4, 4).BonusType = "BONUS_IRON"
	wb.plotAt(3, 3).BonusType = "BONUS_GOLD" // a corner of the fat cross, not counted
	wb.plotAt(7, 5).BonusType = "BONUS_MAGIC"
	wb.Players[0].CivType, wb.Players[0].LeaderType = "CIVILIZATION_ROME", "LEADER_CAESAR"
	wb.Players[0].StartingX, wb.Players[0].StartingY = 5, 5
	wb.Players[1].CivType, wb.Players[1].LeaderType = "CIVILIZATION_GREECE", "LEADER_ALEXANDER"
	wb.Players[1].StartingX, wb.Players[1].StartingY = 18, 5 // in water, 7 plots away across the seam
	wb.Players[2].CivType, wb.Players[2].LeaderType = "CIVILIZATION_EGYPT", "LEADER_RAMESSES"
	wb.Players[2].RandomStartLocation = true

	result := wb.StartBalance()
	if len(result) != 2 {
		t.Fatalf("starts = %+v", result)
	}
	s := result[0]
	if s.Land != 21 || s.Good != 21 || s.Hills != 1 || s.Forests != 1 || s.Water != 0 || s.Coastal || !s.River {
		t.Errorf("land of player 0: %+v", s)
	}
	if s.Food != 1 || s.Strategic != 1 || s.Luxury != 1 || len(s.Resources) != 3 {
		t.Errorf("resources of player 0: %+v", s)
	}
	if s.Nearest != 7 || result[1].Nearest != 7 {
		t.Errorf("nearest = %v, %v", s.Nearest, result[1].Nearest)
	}
	if w := result[1]; w.Water != 21 || w.Land != 0 || w.River {
		t.Errorf("start in the ocean: %+v", w)
	}
}
