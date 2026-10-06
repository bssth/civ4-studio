package editor

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Region is a rectangle of plots. X is the western column, Y the southern row; on a map wrapping
// east-west the rectangle may cross the seam (X+Width goes past the last column).
type Region struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// regionClip is a copied rectangle, plots are stored by their position inside of it (x*height+y)
type regionClip struct {
	width, height int
	plots         []*Plot
}

// ClipboardInfo tells the size of the copied region, zero when nothing is copied
type ClipboardInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
	Cities int `json:"cities"`
	Units  int `json:"units"`
}

// regionCells returns m.Plots indexes of a region (-1 for positions without a plot), column by column.
// Columns wrap on a map wrapping east-west, rows and columns outside of the map are -1.
func (m *WbMap) regionCells(r Region) ([]int, error) {
	if m.Map == nil {
		return nil, errors.New("the map has no BeginMap section")
	}
	if r.Width <= 0 || r.Height <= 0 {
		return nil, errors.New("the region is empty")
	}
	w, h := int(m.Map.GridWidth), int(m.Map.GridHeight)
	if r.Width > w || r.Height > h {
		return nil, fmt.Errorf("the region %dx%d is bigger than the map", r.Width, r.Height)
	}
	index := make(map[PlotXY]int, len(m.Plots))
	for i, p := range m.Plots {
		index[PlotXY{int(p.X), int(p.Y)}] = i
	}
	cells := make([]int, 0, r.Width*r.Height)
	for dx := 0; dx < r.Width; dx++ {
		x := r.X + dx
		if m.Map.WrapX != 0 {
			x = ((x % w) + w) % w
		}
		for dy := 0; dy < r.Height; dy++ {
			i, ok := index[PlotXY{x, r.Y + dy}]
			if !ok {
				i = -1
			}
			cells = append(cells, i)
		}
	}
	return cells, nil
}

// CopyRegion keeps the plots of a region in the clipboard of the editor.
func (a *App) CopyRegion(r Region) (ClipboardInfo, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.wbMap == nil {
		return ClipboardInfo{}, errors.New("no map loaded")
	}
	cells, err := a.wbMap.regionCells(r)
	if err != nil {
		return ClipboardInfo{}, err
	}
	clip := &regionClip{width: r.Width, height: r.Height, plots: make([]*Plot, len(cells))}
	for k, i := range cells {
		if i >= 0 {
			clip.plots[k] = clonePlot(a.wbMap.Plots[i])
		}
	}
	a.clipboard = clip
	return a.clipboardInfo(), nil
}

// GetClipboard returns the size of the copied region.
func (a *App) GetClipboard() ClipboardInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.clipboardInfo()
}

func (a *App) clipboardInfo() ClipboardInfo {
	c := a.clipboard
	if c == nil {
		return ClipboardInfo{}
	}
	info := ClipboardInfo{Width: c.width, Height: c.height}
	for _, p := range c.plots {
		if p != nil {
			info.Cities += len(p.Cities)
			info.Units += len(p.Units)
		}
	}
	return info
}

// pasteInto copies the land of src into dst: terrain, height, features, resources, improvements,
// routes and rivers. With assets cities and units are replaced too.
func pasteInto(dst, src *Plot, assets bool) {
	dst.TerrainType, dst.PlotType = src.TerrainType, src.PlotType
	dst.FeatureType = append([]string(nil), src.FeatureType...)
	dst.FeatureVariety = append([]string(nil), src.FeatureVariety...)
	dst.BonusType, dst.ImprovementType, dst.RouteType = src.BonusType, src.ImprovementType, src.RouteType
	dst.IsNOfRiver, dst.IsWOfRiver = src.IsNOfRiver, src.IsWOfRiver
	dst.RiverNSDirection, dst.RiverWEDirection = src.RiverNSDirection, src.RiverWEDirection
	if assets {
		c := clonePlot(src)
		dst.Cities, dst.Units = c.Cities, c.Units
	}
}

// PasteRegion pastes the copied region with its south-western plot at x, y as one undoable step.
// Parts outside of the map are skipped. Returns the number of changed plots.
func (a *App) PasteRegion(x, y int, assets bool) (int, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return 0, errors.New("no map loaded")
	}
	clip := a.clipboard
	if clip == nil {
		a.mu.Unlock()
		return 0, errors.New("nothing is copied")
	}
	cells, err := a.wbMap.regionCells(Region{X: x, Y: y, Width: clip.width, Height: clip.height})
	if err != nil {
		a.mu.Unlock()
		return 0, err
	}
	var indexes []int
	var before, after []*Plot
	for k, i := range cells {
		src := clip.plots[k]
		if i < 0 || src == nil {
			continue
		}
		p := a.wbMap.Plots[i]
		old := clonePlot(p)
		pasteInto(p, src, assets)
		if string(old.ToWbFormat()) != string(p.ToWbFormat()) {
			indexes = append(indexes, i)
			before = append(before, old)
			after = append(after, clonePlot(p))
		}
	}
	if len(indexes) > 0 {
		a.history.push(plotsEntry(fmt.Sprintf("paste:%d", len(indexes)), indexes, before, after))
	}
	a.mu.Unlock()

	if len(indexes) > 0 {
		a.setDirty(true)
	}
	return len(indexes), nil
}

// ClearRegion removes units and/or cities from the plots of a region as one undoable step.
// Returns the number of removed units and cities.
func (a *App) ClearRegion(r Region, units, cities bool) (ClipboardInfo, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return ClipboardInfo{}, errors.New("no map loaded")
	}
	cells, err := a.wbMap.regionCells(r)
	if err != nil {
		a.mu.Unlock()
		return ClipboardInfo{}, err
	}
	var removed ClipboardInfo
	var indexes []int
	var before, after []*Plot
	for _, i := range cells {
		if i < 0 {
			continue
		}
		p := a.wbMap.Plots[i]
		if !(units && len(p.Units) > 0 || cities && len(p.Cities) > 0) {
			continue
		}
		old := clonePlot(p)
		if units {
			removed.Units += len(p.Units)
			p.Units = nil
		}
		if cities {
			removed.Cities += len(p.Cities)
			p.Cities = nil
		}
		indexes = append(indexes, i)
		before = append(before, old)
		after = append(after, clonePlot(p))
	}
	if len(indexes) > 0 {
		a.history.push(plotsEntry(fmt.Sprintf("clearArea:%d", len(indexes)), indexes, before, after))
	}
	a.mu.Unlock()

	if len(indexes) > 0 {
		a.setDirty(true)
	}
	return removed, nil
}

// writeDataURL saves a "data:image/png;base64,..." URL as a file
func writeDataURL(path, dataURL string) error {
	const prefix = "data:image/png;base64,"
	if !strings.HasPrefix(dataURL, prefix) {
		return errors.New("the image must be a PNG data URL")
	}
	data, err := base64.StdEncoding.DecodeString(dataURL[len(prefix):])
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// ExportImage asks where to save a PNG image of the map (a data URL drawn by the frontend) and saves it.
// Returns the path, empty if the user cancelled.
func (a *App) ExportImage(dataURL string) (string, error) {
	a.mu.Lock()
	current := a.filePath
	a.mu.Unlock()
	name := "map.png"
	dir := ""
	if current != "" {
		dir = filepath.Dir(current)
		name = strings.TrimSuffix(filepath.Base(current), filepath.Ext(current)) + ".png"
	}
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:            "Export map image",
		DefaultDirectory: dir,
		DefaultFilename:  name,
		Filters:          []runtime.FileFilter{{DisplayName: "PNG image (*.png)", Pattern: "*.png"}},
	})
	if err != nil || path == "" {
		return "", err
	}
	if filepath.Ext(path) == "" {
		path += ".png"
	}
	if err := writeDataURL(path, dataURL); err != nil {
		return "", err
	}
	ConsoleWrite("Exported the map image to %s", path)
	return path, nil
}
