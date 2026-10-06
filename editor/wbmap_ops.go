package editor

import (
	"errors"
	"fmt"
)

// remapPlayers changes every reference to a player (units, cities, culture, attitudes, signs)
// using mapping old index -> new index. Indexes missing in the mapping are kept as they are.
func (m *WbMap) remapPlayers(mapping map[int]int) {
	remap := func(i int) int {
		if n, ok := mapping[i]; ok {
			return n
		}
		return i
	}

	for _, plot := range m.Plots {
		for _, unit := range plot.Units {
			unit.UnitOwner = remap(unit.UnitOwner)
		}
		for _, city := range plot.Cities {
			city.CityOwner = uint(remap(int(city.CityOwner)))
			if len(city.PlayerCulture) > 0 {
				culture := make(map[uint]uint64, len(city.PlayerCulture))
				for player, value := range city.PlayerCulture {
					culture[uint(remap(int(player)))] += value
				}
				city.PlayerCulture = culture
			}
		}
	}
	for _, player := range m.Players {
		for i, other := range player.AttitudePlayer {
			player.AttitudePlayer[i] = uint(remap(int(other)))
		}
	}
	for _, sign := range m.Signs {
		if sign.PlayerType >= 0 {
			sign.PlayerType = remap(sign.PlayerType)
		}
	}
}

// SwapPlayers exchanges two player slots and updates all references to them,
// so units, cities and culture stay with their civilization.
func (m *WbMap) SwapPlayers(a, b int) error {
	if a < 0 || b < 0 || a >= len(m.Players) || b >= len(m.Players) {
		return fmt.Errorf("players %d and %d must be within 0..%d", a, b, len(m.Players)-1)
	}
	if a == b {
		return nil
	}
	m.Players[a], m.Players[b] = m.Players[b], m.Players[a]
	m.remapPlayers(map[int]int{a: b, b: a})
	return nil
}

// EmptyPlayer returns an empty player slot as the game writes it
func EmptyPlayer(team uint) *Player {
	return &Player{
		LeaderType: NonePlayer, CivType: NonePlayer, Team: team,
		Handicap: "HANDICAP_NOBLE", Color: NonePlayer, ArtStyle: NonePlayer,
	}
}

// ClearPlayer turns a player slot into an empty one, keeping its team.
// With removeAssets its units and cities are removed from the map, and other players
// forget their attitude to it. Returns the number of removed units and cities.
func (m *WbMap) ClearPlayer(index int, removeAssets bool) (units int, cities int, err error) {
	if index < 0 || index >= len(m.Players) {
		return 0, 0, fmt.Errorf("there is no player %d", index)
	}
	handicap := m.Players[index].Handicap
	m.Players[index] = EmptyPlayer(m.Players[index].Team)
	if handicap != "" {
		m.Players[index].Handicap = handicap
	}

	if !removeAssets {
		return 0, 0, nil
	}
	for _, plot := range m.Plots {
		keptUnits := plot.Units[:0]
		for _, unit := range plot.Units {
			if unit.UnitOwner == index {
				units++
			} else {
				keptUnits = append(keptUnits, unit)
			}
		}
		plot.Units = keptUnits

		keptCities := plot.Cities[:0]
		for _, city := range plot.Cities {
			if int(city.CityOwner) == index {
				cities++
			} else {
				delete(city.PlayerCulture, uint(index))
				keptCities = append(keptCities, city)
			}
		}
		plot.Cities = keptCities
	}
	for _, player := range m.Players {
		for i := 0; i < len(player.AttitudePlayer) && i < len(player.AttitudeExtra); {
			if int(player.AttitudePlayer[i]) == index {
				player.AttitudePlayer = append(player.AttitudePlayer[:i], player.AttitudePlayer[i+1:]...)
				player.AttitudeExtra = append(player.AttitudeExtra[:i], player.AttitudeExtra[i+1:]...)
				continue
			}
			i++
		}
	}
	return units, cities, nil
}

// SwapPlayers exchanges two player slots and updates units, cities, culture, attitudes and signs.
func (a *App) SwapPlayers(first, second int) error {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	err := a.wbMap.SwapPlayers(first, second)
	if err == nil && first != second {
		a.history.push(snapshotEntry(fmt.Sprintf("swap:%d,%d", first, second), before, a.wbMap.snapshot()))
	}
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if first != second {
		a.setDirty(true)
		ConsoleWrite("Swapped players %d and %d", first, second)
	}
	return nil
}

// ClearPlayer empties a player slot; with removeAssets its units and cities are removed too.
func (a *App) ClearPlayer(index int, removeAssets bool) error {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	units, cities, err := a.wbMap.ClearPlayer(index, removeAssets)
	if err == nil {
		a.history.push(snapshotEntry(fmt.Sprintf("clear:%d", index), before, a.wbMap.snapshot()))
	}
	a.mu.Unlock()
	if err != nil {
		return err
	}
	a.setDirty(true)
	ConsoleWrite("Cleared player %d (removed %d units and %d cities)", index, units, cities)
	return nil
}
