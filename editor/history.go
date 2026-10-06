package editor

import (
	"errors"
	"maps"
	"slices"
)

// maxHistory limits the number of steps that can be undone
const maxHistory = 100

// historyEntry is one undoable change of the map
type historyEntry struct {
	label string
	undo  func(m *WbMap)
	redo  func(m *WbMap)
}

// History keeps undo and redo stacks of map edits made on the World tab
// (painting, plot edits, start positions). It is reset when another map is opened.
type History struct {
	undo []historyEntry
	redo []historyEntry
}

func (h *History) push(entry historyEntry) {
	h.undo = append(h.undo, entry)
	if len(h.undo) > maxHistory {
		h.undo = h.undo[len(h.undo)-maxHistory:]
	}
	h.redo = nil
}

func (h *History) reset() {
	h.undo, h.redo = nil, nil
}

// HistoryState tells which steps can be undone and redone, empty labels mean none.
// Labels are codes the frontend turns into text: "paint:<plots>", "plot:<x>,<y>", "start:<player>".
type HistoryState struct {
	Undo string `json:"undo"`
	Redo string `json:"redo"`
}

func (h *History) state() HistoryState {
	var s HistoryState
	if len(h.undo) > 0 {
		s.Undo = h.undo[len(h.undo)-1].label
	}
	if len(h.redo) > 0 {
		s.Redo = h.redo[len(h.redo)-1].label
	}
	return s
}

// clonePlot makes a deep copy, so a stored history state is not changed by later edits
func clonePlot(p *Plot) *Plot {
	c := *p
	c.FeatureType = slices.Clone(p.FeatureType)
	c.FeatureVariety = slices.Clone(p.FeatureVariety)
	c.TeamReveal = slices.Clone(p.TeamReveal)
	c.Extra = slices.Clone(p.Extra)
	c.Units = nil
	for _, u := range p.Units {
		unit := *u
		unit.PromotionType = slices.Clone(u.PromotionType)
		unit.Extra = slices.Clone(u.Extra)
		c.Units = append(c.Units, &unit)
	}
	c.Cities = nil
	for _, city := range p.Cities {
		cc := *city
		cc.BuildingType = slices.Clone(city.BuildingType)
		cc.ReligionType = slices.Clone(city.ReligionType)
		cc.HolyCityReligionType = slices.Clone(city.HolyCityReligionType)
		cc.PlayerCulture = maps.Clone(city.PlayerCulture)
		cc.Extra = slices.Clone(city.Extra)
		c.Cities = append(c.Cities, &cc)
	}
	return &c
}

// plotsEntry restores plots at given indexes of m.Plots
func plotsEntry(label string, indexes []int, before, after []*Plot) historyEntry {
	restore := func(states []*Plot) func(m *WbMap) {
		return func(m *WbMap) {
			for i, idx := range indexes {
				if idx < len(m.Plots) {
					m.Plots[idx] = clonePlot(states[i])
				}
			}
		}
	}
	return historyEntry{label: label, undo: restore(before), redo: restore(after)}
}

// HistoryState returns labels of the steps that can be undone and redone.
func (a *App) HistoryState() HistoryState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.history.state()
}

// Undo reverts the last map edit made on the World tab.
func (a *App) Undo() (HistoryState, error) {
	return a.step(true)
}

// Redo applies the last undone map edit again.
func (a *App) Redo() (HistoryState, error) {
	return a.step(false)
}

func (a *App) step(undo bool) (HistoryState, error) {
	a.mu.Lock()
	from, to := &a.history.undo, &a.history.redo
	if !undo {
		from, to = to, from
	}
	if a.wbMap == nil || len(*from) == 0 {
		state := a.history.state()
		a.mu.Unlock()
		return state, errors.New("nothing to undo or redo")
	}
	entry := (*from)[len(*from)-1]
	*from = (*from)[:len(*from)-1]
	if undo {
		entry.undo(a.wbMap)
	} else {
		entry.redo(a.wbMap)
	}
	*to = append(*to, entry)
	state := a.history.state()
	a.mu.Unlock()

	a.setDirty(true)
	return state, nil
}
