package editor

import (
	"errors"
	"fmt"
	"strings"
)

// GetSigns returns all signs of the map (nil if no map loaded).
func (a *App) GetSigns() []*Sign {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return nil
	}
	if a.wbMap.Signs == nil {
		return []*Sign{}
	}
	return a.wbMap.Signs
}

// SetPlotSigns replaces the signs of the plot x, y. Other signs keep their order,
// the new ones take the place of the first replaced sign.
func (a *App) SetPlotSigns(x, y int, signs []*Sign) error {
	for _, sign := range signs {
		if sign == nil {
			return errors.New("sign is nil")
		}
		if strings.ContainsAny(sign.Caption, "\r\n") {
			return errors.New("a sign caption must be one line")
		}
		sign.PlotX, sign.PlotY = x, y
	}

	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return errors.New("no map loaded")
	}
	if !a.wbMap.inGrid(x, y) {
		a.mu.Unlock()
		return fmt.Errorf("%d,%d is outside of the map", x, y)
	}
	old := a.wbMap.Signs
	var result, current []*Sign
	inserted := false
	for _, sign := range old {
		if sign.PlotX == x && sign.PlotY == y {
			current = append(current, sign)
			if !inserted {
				result = append(result, signs...)
				inserted = true
			}
			continue
		}
		result = append(result, sign)
	}
	if !inserted {
		result = append(result, signs...)
	}
	changed := !sameWbFormat(current, signs)
	a.wbMap.Signs = result
	if changed {
		a.history.push(sectionEntry(fmt.Sprintf("signs:%d,%d", x, y), old, result,
			func(m *WbMap, v []*Sign) { m.Signs = v }))
	}
	a.mu.Unlock()

	if changed {
		a.setDirty(true)
	}
	return nil
}
