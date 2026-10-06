package editor

import (
	"encoding/json"
	"errors"
	"maps"
	"slices"
	"time"
)

// maxHistory limits the number of steps that can be undone
const maxHistory = 100

// mergeWindow is how long edits of the same section are merged into one step,
// so typing a name or dragging a slider is undone at once
const mergeWindow = 1500 * time.Millisecond

// historyEntry is one undoable change of the map
type historyEntry struct {
	label string
	undo  func(m *WbMap)
	redo  func(m *WbMap)
	// merge allows to join the entry with the next one of the same label made within mergeWindow
	merge bool
	at    time.Time
}

// History keeps undo and redo stacks of map edits. It is reset when another map is opened.
type History struct {
	undo []historyEntry
	redo []historyEntry
	// sealed forbids merging the next entry into the last one (e.g. after undo)
	sealed bool
	now    func() time.Time
}

func (h *History) clock() time.Time {
	if h.now != nil {
		return h.now()
	}
	return time.Now()
}

func (h *History) push(entry historyEntry) {
	entry.at = h.clock()
	if n := len(h.undo); n > 0 && entry.merge && !h.sealed {
		last := &h.undo[n-1]
		if last.merge && last.label == entry.label && entry.at.Sub(last.at) < mergeWindow {
			// Keep the state before the first edit and the state after the latest one
			last.redo, last.at = entry.redo, entry.at
			h.redo = nil
			return
		}
	}
	h.sealed = false
	h.undo = append(h.undo, entry)
	if len(h.undo) > maxHistory {
		h.undo = h.undo[len(h.undo)-maxHistory:]
	}
	h.redo = nil
}

func (h *History) reset() {
	h.undo, h.redo, h.sealed = nil, nil, false
}

// HistoryState tells which steps can be undone and redone, empty labels mean none.
// Labels are codes the frontend turns into text: "paint:<plots>", "plot:<x>,<y>", "start:<player>",
// "game", "map", "teams", "players", "swap:<a>,<b>", "clear:<player>", "signs:<x>,<y>", "resize:<width>,<height>",
// "paste:<plots>", "clearArea:<plots>".
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

// cloneValue makes a deep copy of a map section through JSON. Sections other than plots
// have no unexported fields, so nothing is lost.
func cloneValue[T any](v T) T {
	var c T
	data, err := json.Marshal(v)
	if err == nil {
		err = json.Unmarshal(data, &c)
	}
	if err != nil {
		panic("can not copy a map section: " + err.Error())
	}
	return c
}

// sectionEntry restores a section of the map, set must store a copy of the value it gets
func sectionEntry[T any](label string, before, after T, set func(m *WbMap, v T)) historyEntry {
	before, after = cloneValue(before), cloneValue(after)
	return historyEntry{
		label: label,
		merge: true,
		undo:  func(m *WbMap) { set(m, cloneValue(before)) },
		redo:  func(m *WbMap) { set(m, cloneValue(after)) },
	}
}

// mapSnapshot holds the parts of the map changed by player operations
type mapSnapshot struct {
	props   *MapProps
	players []*Player
	teams   []*Team
	plots   []*Plot
	signs   []*Sign
}

func (m *WbMap) snapshot() mapSnapshot {
	s := mapSnapshot{props: cloneValue(m.Map), players: cloneValue(m.Players), teams: cloneValue(m.Teams), signs: cloneValue(m.Signs)}
	s.plots = make([]*Plot, len(m.Plots))
	for i, p := range m.Plots {
		s.plots[i] = clonePlot(p)
	}
	return s
}

func (s mapSnapshot) restore(m *WbMap) {
	c := mapSnapshot{props: cloneValue(s.props), players: cloneValue(s.players), teams: cloneValue(s.teams), signs: cloneValue(s.signs)}
	c.plots = make([]*Plot, len(s.plots))
	for i, p := range s.plots {
		c.plots[i] = clonePlot(p)
	}
	m.Map, m.Players, m.Teams, m.Plots, m.Signs = c.props, c.players, c.teams, c.plots, c.signs
}

func snapshotEntry(label string, before, after mapSnapshot) historyEntry {
	return historyEntry{label: label, undo: before.restore, redo: after.restore}
}

// HistoryState returns labels of the steps that can be undone and redone.
func (a *App) HistoryState() HistoryState {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.history.state()
}

// Undo reverts the last map edit.
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
	a.history.sealed = true
	state := a.history.state()
	a.mu.Unlock()

	a.setDirty(true)
	return state, nil
}
