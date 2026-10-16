package editor

import (
	"errors"
	"fmt"
)

// MoveResult tells how many cities and units were moved
type MoveResult struct {
	Cities int `json:"cities"`
	Units  int `json:"units"`
}

// MoveAssets moves the city (with city) and/or the units (with units) of plot fromX, fromY to plot toX, toY.
// A city can not be moved into water, onto a peak or onto a plot with another city; then nothing is moved.
func (m *WbMap) MoveAssets(fromX, fromY, toX, toY int, city, units bool) (MoveResult, error) {
	var r MoveResult
	if fromX == toX && fromY == toY {
		return r, errors.New("the plots are the same")
	}
	fi, ti := m.FindPlot(fromX, fromY), m.FindPlot(toX, toY)
	if fi < 0 {
		return r, fmt.Errorf("there is no plot %d, %d", fromX, fromY)
	}
	if ti < 0 {
		return r, fmt.Errorf("there is no plot %d, %d", toX, toY)
	}
	from, to := m.Plots[fi], m.Plots[ti]
	city = city && len(from.Cities) > 0
	units = units && len(from.Units) > 0
	if !city && !units {
		return r, errors.New("there is nothing to move")
	}
	if city {
		switch {
		case len(to.Cities) > 0:
			return r, fmt.Errorf("there is already a city on plot %d, %d", toX, toY)
		case to.PlotType == PlotOcean:
			return r, errors.New("a city can not stand in water")
		case to.PlotType == PlotPeak:
			return r, errors.New("a city can not stand on a peak")
		}
		r.Cities = len(from.Cities)
		to.Cities, from.Cities = append(to.Cities, from.Cities...), nil
	}
	if units {
		r.Units = len(from.Units)
		to.Units, from.Units = append(to.Units, from.Units...), nil
	}
	return r, nil
}

// MoveAssets moves the city and/or the units of a plot to another plot as one undoable step, see WbMap.MoveAssets.
func (a *App) MoveAssets(fromX, fromY, toX, toY int, city, units bool) (MoveResult, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return MoveResult{}, errors.New("no map loaded")
	}
	indexes := []int{a.wbMap.FindPlot(fromX, fromY), a.wbMap.FindPlot(toX, toY)}
	var before []*Plot
	for _, i := range indexes {
		if i >= 0 {
			before = append(before, clonePlot(a.wbMap.Plots[i]))
		}
	}
	result, err := a.wbMap.MoveAssets(fromX, fromY, toX, toY, city, units)
	if err != nil {
		a.mu.Unlock()
		return result, err
	}
	after := []*Plot{clonePlot(a.wbMap.Plots[indexes[0]]), clonePlot(a.wbMap.Plots[indexes[1]])}
	a.history.push(plotsEntry(fmt.Sprintf("move:%d,%d", fromX, fromY), indexes, before, after))
	a.mu.Unlock()

	a.setDirty(true)
	ConsoleWrite("Moved %d cities and %d units from %d, %d to %d, %d", result.Cities, result.Units, fromX, fromY, toX, toY)
	return result, nil
}
