package editor

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// PlotXY is a plot coordinate
type PlotXY struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// PaintOp changes the same properties of many plots at once (a brush stroke).
// Nil fields are not changed; an empty string removes a feature, resource, improvement or route.
type PaintOp struct {
	Cells          []PlotXY `json:"cells"`
	Terrain        *string  `json:"terrain,omitempty"`
	PlotType       *int     `json:"plot_type,omitempty"`
	Feature        *string  `json:"feature,omitempty"`
	FeatureVariety int      `json:"feature_variety"`
	Bonus          *string  `json:"bonus,omitempty"`
	Improvement    *string  `json:"improvement,omitempty"`
	Route          *string  `json:"route,omitempty"`
}

func (op *PaintOp) empty() bool {
	return op.Terrain == nil && op.PlotType == nil && op.Feature == nil &&
		op.Bonus == nil && op.Improvement == nil && op.Route == nil
}

// isWaterTerrain uses game data when it is loaded and the usual names otherwise
func isWaterTerrain(data *GameData, terrain string) bool {
	if data != nil {
		if info := data.Table(InfoTerrains).Get(terrain); info != nil {
			return info.Water
		}
	}
	switch terrain {
	case "TERRAIN_OCEAN", "TERRAIN_COAST", "TERRAIN_LAKE":
		return true
	}
	return strings.Contains(terrain, "OCEAN") || strings.Contains(terrain, "COAST")
}

// apply changes one plot and returns true if anything changed
func (op *PaintOp) apply(p *Plot, data *GameData) bool {
	before := string(p.ToWbFormat())

	if op.Terrain != nil {
		p.TerrainType = *op.Terrain
		// Keep the height consistent unless it is painted too
		if op.PlotType == nil {
			water := isWaterTerrain(data, *op.Terrain)
			if water && p.PlotType != PlotOcean {
				p.PlotType = PlotOcean
			} else if !water && p.PlotType == PlotOcean {
				p.PlotType = PlotLand
			}
		}
	}
	if op.PlotType != nil {
		p.PlotType = uint(*op.PlotType)
	}
	if op.Feature != nil {
		if *op.Feature == "" {
			p.FeatureType, p.FeatureVariety = nil, nil
		} else {
			p.FeatureType = []string{*op.Feature}
			p.FeatureVariety = []string{strconv.Itoa(op.FeatureVariety)}
		}
	}
	if op.Bonus != nil {
		p.BonusType = *op.Bonus
	}
	if op.Improvement != nil {
		p.ImprovementType = *op.Improvement
	}
	if op.Route != nil {
		p.RouteType = *op.Route
	}

	return before != string(p.ToWbFormat())
}

// Paint applies the operation to the plots under the given cells and returns indexes of
// changed plots with their states before and after the change.
func (m *WbMap) Paint(op *PaintOp, data *GameData) (indexes []int, before, after []*Plot) {
	index := make(map[PlotXY]int, len(m.Plots))
	for i, p := range m.Plots {
		index[PlotXY{int(p.X), int(p.Y)}] = i
	}

	seen := make(map[int]bool)
	for _, cell := range op.Cells {
		i, ok := index[cell]
		if !ok || seen[i] {
			continue
		}
		seen[i] = true
		old := clonePlot(m.Plots[i])
		if op.apply(m.Plots[i], data) {
			indexes = append(indexes, i)
			before = append(before, old)
			after = append(after, clonePlot(m.Plots[i]))
		}
	}
	return indexes, before, after
}

// PaintPlots applies a brush stroke as one undoable step. Returns the number of changed plots.
func (a *App) PaintPlots(op *PaintOp) (int, error) {
	if op == nil || op.empty() {
		return 0, errors.New("nothing to paint")
	}
	if op.PlotType != nil && (*op.PlotType < PlotPeak || *op.PlotType > PlotOcean) {
		return 0, fmt.Errorf("invalid plot type %d", *op.PlotType)
	}

	data := CurrentGameData()
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return 0, errors.New("no map loaded")
	}
	indexes, before, after := a.wbMap.Paint(op, data)
	if len(indexes) > 0 {
		a.history.push(plotsEntry(fmt.Sprintf("paint:%d", len(indexes)), indexes, before, after))
	}
	a.mu.Unlock()

	if len(indexes) > 0 {
		a.setDirty(true)
	}
	return len(indexes), nil
}
