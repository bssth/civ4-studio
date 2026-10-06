package editor

// CityEntry is a city with the plot it stands on; Index is its position in the plot's city list
type CityEntry struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Index int   `json:"index"`
	City  *City `json:"city"`
}

// UnitEntry is a unit with the plot it stands on; Index is its position in the plot's unit list
type UnitEntry struct {
	X     int   `json:"x"`
	Y     int   `json:"y"`
	Index int   `json:"index"`
	Unit  *Unit `json:"unit"`
}

// GetCities returns all cities of the map in the order of plots.
func (a *App) GetCities() []CityEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := []CityEntry{}
	if a.wbMap == nil {
		return result
	}
	for _, p := range a.wbMap.Plots {
		for i, c := range p.Cities {
			result = append(result, CityEntry{X: int(p.X), Y: int(p.Y), Index: i, City: c})
		}
	}
	return result
}

// GetUnits returns all units of the map in the order of plots.
func (a *App) GetUnits() []UnitEntry {
	a.mu.Lock()
	defer a.mu.Unlock()
	result := []UnitEntry{}
	if a.wbMap == nil {
		return result
	}
	for _, p := range a.wbMap.Plots {
		for i, u := range p.Units {
			result = append(result, UnitEntry{X: int(p.X), Y: int(p.Y), Index: i, Unit: u})
		}
	}
	return result
}
