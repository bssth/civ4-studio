package editor

import (
	"errors"
	"fmt"
	"sort"
)

// MinGridSize is the smallest map size that can be made by resizing
const MinGridSize = 4

// ResizeResult tells what was lost when a map was made smaller
type ResizeResult struct {
	Units  int `json:"units"`
	Cities int `json:"cities"`
	Signs  int `json:"signs"`
	// Starts are players whose start position is outside of the new map
	Starts []int `json:"starts"`
}

// Resize adds (positive) or removes (negative) columns on the west and east sides and rows on the
// north and south sides. Plots, start positions and signs keep their places on the map; new plots
// are ocean. Units and cities of removed plots are lost.
func (m *WbMap) Resize(west, east, north, south int) (*ResizeResult, error) {
	if m.Map == nil {
		return nil, errors.New("the map has no BeginMap section")
	}
	width := int(m.Map.GridWidth) + west + east
	height := int(m.Map.GridHeight) + north + south
	if width < MinGridSize || height < MinGridSize || width > MaxGridSize || height > MaxGridSize {
		return nil, fmt.Errorf("the new map size %dx%d must be within %d..%d", width, height, MinGridSize, MaxGridSize)
	}
	result := &ResizeResult{Starts: []int{}}
	inside := func(x, y int) bool { return x >= 0 && y >= 0 && x < width && y < height }

	byPosition := make(map[[2]int]*Plot, width*height)
	for _, p := range m.Plots {
		x, y := int(p.X)+west, int(p.Y)+south
		if !inside(x, y) {
			result.Units += len(p.Units)
			result.Cities += len(p.Cities)
			continue
		}
		p.X, p.Y = uint(x), uint(y)
		byPosition[[2]int{x, y}] = p
	}
	plots := make([]*Plot, 0, width*height)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			p := byPosition[[2]int{x, y}]
			if p == nil {
				p = &Plot{X: uint(x), Y: uint(y), TerrainType: DefaultOceanTerrain, PlotType: PlotOcean}
			}
			plots = append(plots, p)
		}
	}
	// Plots outside of the old grid (a broken file) keep their relative order after the grid
	sort.SliceStable(plots, func(i, j int) bool {
		return plots[i].X < plots[j].X || plots[i].X == plots[j].X && plots[i].Y < plots[j].Y
	})
	m.Plots = plots

	for i, p := range m.Players {
		if isEmptySlot(p) {
			continue
		}
		p.StartingX += west
		p.StartingY += south
		if !p.RandomStartLocation && !inside(p.StartingX, p.StartingY) {
			result.Starts = append(result.Starts, i)
		}
	}

	signs := m.Signs[:0]
	for _, s := range m.Signs {
		s.PlotX += west
		s.PlotY += south
		if inside(s.PlotX, s.PlotY) {
			signs = append(signs, s)
		} else {
			result.Signs++
		}
	}
	m.Signs = signs

	m.Map.GridWidth, m.Map.GridHeight = uint64(width), uint64(height)
	return result, nil
}

// ResizeMap adds or removes columns and rows on each side of the map, see WbMap.Resize.
func (a *App) ResizeMap(west, east, north, south int) (*ResizeResult, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return nil, errors.New("no map loaded")
	}
	if west == 0 && east == 0 && north == 0 && south == 0 {
		a.mu.Unlock()
		return &ResizeResult{Starts: []int{}}, nil
	}
	before := a.wbMap.snapshot()
	result, err := a.wbMap.Resize(west, east, north, south)
	if err != nil {
		before.restore(a.wbMap)
		a.mu.Unlock()
		return nil, err
	}
	width, height := a.wbMap.Map.GridWidth, a.wbMap.Map.GridHeight
	a.history.push(snapshotEntry(fmt.Sprintf("resize:%d,%d", width, height), before, a.wbMap.snapshot()))
	a.mu.Unlock()

	a.setDirty(true)
	ConsoleWrite("Resized the map to %dx%d (removed %d units, %d cities, %d signs)",
		width, height, result.Units, result.Cities, result.Signs)
	return result, nil
}
