package editor

import (
	"fmt"
	"strings"
)

// maxSearchResults limits the search, a short query may match half of the map
const maxSearchResults = 50

// SearchResult is a found place: Kind is "city", "start", "sign" or "landmark"
type SearchResult struct {
	Kind   string `json:"kind"`
	Label  string `json:"label"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	Player int    `json:"player"`
}

// Search finds cities, start positions of players (by civilization, leader and type names),
// signs and landmarks whose names contain the query, ignoring case.
func (m *WbMap) Search(query string) []SearchResult {
	q := strings.ToLower(strings.TrimSpace(query))
	results := []SearchResult{}
	if q == "" {
		return results
	}
	matches := func(values ...string) bool {
		for _, v := range values {
			if v != "" && strings.Contains(strings.ToLower(v), q) {
				return true
			}
		}
		return false
	}
	add := func(r SearchResult) bool {
		results = append(results, r)
		return len(results) < maxSearchResults
	}

	for i, p := range m.Players {
		if isEmptySlot(p) || !matches(p.CivDesc, p.CivShortDesc, p.CivAdjective, p.LeaderName, p.CivType, p.LeaderType) {
			continue
		}
		name := p.CivDesc
		if name == "" || strings.HasPrefix(name, "TXT_KEY_") {
			name = HumanizeType(p.CivType)
		}
		if p.LeaderName != "" && !strings.HasPrefix(p.LeaderName, "TXT_KEY_") {
			name += fmt.Sprintf(" (%s)", p.LeaderName)
		}
		if !add(SearchResult{Kind: "start", Label: name, X: p.StartingX, Y: p.StartingY, Player: i}) {
			return results
		}
	}
	for _, p := range m.Plots {
		for _, c := range p.Cities {
			if matches(c.CityName) && !add(SearchResult{Kind: "city", Label: c.CityName, X: int(p.X), Y: int(p.Y), Player: int(c.CityOwner)}) {
				return results
			}
		}
		if matches(p.Landmark) && !add(SearchResult{Kind: "landmark", Label: p.Landmark, X: int(p.X), Y: int(p.Y), Player: -1}) {
			return results
		}
	}
	for _, s := range m.Signs {
		if matches(s.Caption) && !add(SearchResult{Kind: "sign", Label: s.Caption, X: s.PlotX, Y: s.PlotY, Player: s.PlayerType}) {
			return results
		}
	}
	return results
}

// SearchMap finds places on the map by name, see WbMap.Search.
func (a *App) SearchMap(query string) []SearchResult {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return []SearchResult{}
	}
	return a.wbMap.Search(query)
}
