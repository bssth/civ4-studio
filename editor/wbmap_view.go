package editor

import (
	"bytes"
	"errors"
	"fmt"
)

// Plot flags in MapView.Flags
const (
	PlotFlagNOfRiver    = 1 << iota // river along the southern edge
	PlotFlagWOfRiver                // river along the eastern edge
	PlotFlagStart                   // StartingPlot
	PlotFlagImprovement             // has an improvement
	PlotFlagRoute                   // has a route
	PlotFlagLandmark                // has a landmark text
	PlotFlagSign                    // has a sign
)

// MapView is a compact, columnar form of the plots for drawing the map.
// Cell arrays are indexed by y*Width+x; type arrays hold indexes into dictionaries, -1 means none.
type MapView struct {
	Width     int      `json:"width"`
	Height    int      `json:"height"`
	Terrains  []string `json:"terrains"`
	Features  []string `json:"features"`
	Bonuses   []string `json:"bonuses"`
	Terrain   []int    `json:"terrain"`
	PlotType  []int    `json:"plot_type"`
	Feature   []int    `json:"feature"`
	Bonus     []int    `json:"bonus"`
	Flags     []int    `json:"flags"`
	CityOwner []int    `json:"city_owner"`
	UnitOwner []int    `json:"unit_owner"`
	UnitCount []int    `json:"unit_count"`
}

type dictionary struct {
	values []string
	index  map[string]int
}

func (d *dictionary) id(value string) int {
	if value == "" || value == NonePlayer {
		return -1
	}
	if d.index == nil {
		d.index = make(map[string]int)
	}
	if i, ok := d.index[value]; ok {
		return i
	}
	d.index[value] = len(d.values)
	d.values = append(d.values, value)
	return len(d.values) - 1
}

func (d *dictionary) list() []string {
	if d.values == nil {
		return []string{}
	}
	return d.values
}

func filled(n, value int) []int {
	result := make([]int, n)
	for i := range result {
		result[i] = value
	}
	return result
}

// View builds a MapView. Plots outside of the grid are skipped (see Stats for such problems).
func (m *WbMap) View() *MapView {
	view := &MapView{}
	if m.Map != nil {
		view.Width, view.Height = int(m.Map.GridWidth), int(m.Map.GridHeight)
	}
	n := view.Width * view.Height
	view.Terrain, view.Feature, view.Bonus = filled(n, -1), filled(n, -1), filled(n, -1)
	view.PlotType = filled(n, PlotOcean)
	view.Flags = make([]int, n)
	view.CityOwner, view.UnitOwner = filled(n, -1), filled(n, -1)
	view.UnitCount = make([]int, n)

	var terrains, features, bonuses dictionary
	for _, p := range m.Plots {
		x, y := int(p.X), int(p.Y)
		if x >= view.Width || y >= view.Height {
			continue
		}
		i := y*view.Width + x
		view.Terrain[i] = terrains.id(p.TerrainType)
		view.PlotType[i] = int(p.PlotType)
		if len(p.FeatureType) > 0 {
			view.Feature[i] = features.id(p.FeatureType[0])
		}
		view.Bonus[i] = bonuses.id(p.BonusType)

		flags := 0
		if p.IsNOfRiver {
			flags |= PlotFlagNOfRiver
		}
		if p.IsWOfRiver {
			flags |= PlotFlagWOfRiver
		}
		if p.StartingPlot {
			flags |= PlotFlagStart
		}
		if p.ImprovementType != "" && p.ImprovementType != NonePlayer {
			flags |= PlotFlagImprovement
		}
		if p.RouteType != "" && p.RouteType != NonePlayer {
			flags |= PlotFlagRoute
		}
		if p.Landmark != "" {
			flags |= PlotFlagLandmark
		}
		view.Flags[i] = flags

		if len(p.Cities) > 0 {
			view.CityOwner[i] = int(p.Cities[0].CityOwner)
		}
		view.UnitCount[i] = len(p.Units)
		if len(p.Units) > 0 {
			view.UnitOwner[i] = p.Units[0].UnitOwner
		}
	}

	for _, sign := range m.Signs {
		if sign.PlotX >= 0 && sign.PlotY >= 0 && sign.PlotX < view.Width && sign.PlotY < view.Height {
			view.Flags[sign.PlotY*view.Width+sign.PlotX] |= PlotFlagSign
		}
	}

	view.Terrains, view.Features, view.Bonuses = terrains.list(), features.list(), bonuses.list()
	return view
}

// FindPlot returns the index of the plot with given coordinates in m.Plots, -1 if there is none
func (m *WbMap) FindPlot(x, y int) int {
	for i, p := range m.Plots {
		if int(p.X) == x && int(p.Y) == y {
			return i
		}
	}
	return -1
}

// GetMapView returns plots in a compact form for drawing (nil if no map loaded).
func (a *App) GetMapView() *MapView {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	return a.wbMap.View()
}

// GetPlot returns the plot with given coordinates, including its units and cities (nil if there is none).
func (a *App) GetPlot(x, y int) *Plot {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	if i := a.wbMap.FindPlot(x, y); i >= 0 {
		return a.wbMap.Plots[i]
	}
	return nil
}

// SetPlot replaces the plot with the same coordinates.
func (a *App) SetPlot(plot *Plot) error {
	if plot == nil {
		return errors.New("plot is nil")
	}
	if len(plot.FeatureVariety) < len(plot.FeatureType) {
		return errors.New("every feature must have a variety")
	}
	for _, city := range plot.Cities {
		if city == nil {
			return errors.New("city is nil")
		}
	}
	for _, unit := range plot.Units {
		if unit == nil {
			return errors.New("unit is nil")
		}
		if unit.Damage > 100 {
			return errors.New("unit damage must be within 0..100")
		}
	}

	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	i := a.wbMap.FindPlot(int(plot.X), int(plot.Y))
	if i < 0 {
		a.mu.Unlock()
		return fmt.Errorf("there is no plot %d,%d", plot.X, plot.Y)
	}
	old := a.wbMap.Plots[i]
	// Not exported to the frontend, keep how the original file writes TeamReveal
	plot.teamRevealAsList = old.teamRevealAsList
	changed := !bytes.Equal(old.ToWbFormat(), plot.ToWbFormat())
	a.wbMap.Plots[i] = plot
	if changed {
		a.history.push(plotsEntry(fmt.Sprintf("plot:%d,%d", plot.X, plot.Y),
			[]int{i}, []*Plot{clonePlot(old)}, []*Plot{clonePlot(plot)}))
	}
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
	return nil
}

// SetPlayerStart moves the fixed starting position of a player and turns off its random start location.
func (a *App) SetPlayerStart(player, x, y int) error {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	if player < 0 || player >= len(a.wbMap.Players) {
		a.mu.Unlock()
		return fmt.Errorf("there is no player %d", player)
	}
	if m := a.wbMap.Map; m != nil && (x < 0 || y < 0 || uint64(x) >= m.GridWidth || uint64(y) >= m.GridHeight) {
		a.mu.Unlock()
		return fmt.Errorf("%d,%d is outside of the map", x, y)
	}
	p := a.wbMap.Players[player]
	changed := p.StartingX != x || p.StartingY != y || p.RandomStartLocation
	oldX, oldY, oldRandom := p.StartingX, p.StartingY, p.RandomStartLocation
	p.StartingX, p.StartingY, p.RandomStartLocation = x, y, false
	if changed {
		setStart := func(sx, sy int, random bool) func(m *WbMap) {
			return func(m *WbMap) {
				if player < len(m.Players) {
					q := m.Players[player]
					q.StartingX, q.StartingY, q.RandomStartLocation = sx, sy, random
				}
			}
		}
		a.history.push(historyEntry{
			label: fmt.Sprintf("start:%d", player),
			undo:  setStart(oldX, oldY, oldRandom),
			redo:  setStart(x, y, false),
		})
	}
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
	return nil
}
