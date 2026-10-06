package editor

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"math/rand"
	"strings"
)

// maxImageDataURL limits the size of an imported picture (as a data URL)
const maxImageDataURL = 40 << 20

// ImageImportOptions are the settings of the image import. Mode "height" reads a height map (bright is high),
// "colors" reads the colors of an ordinary map picture. Land, Hills, Peaks, Forests, Rivers and Resources work
// as in the terrain generator; Land is used only for height maps.
type ImageImportOptions struct {
	Mode      string `json:"mode"`
	Invert    bool   `json:"invert"`
	Seed      int64  `json:"seed"`
	Land      int    `json:"land"`
	Hills     int    `json:"hills"`
	Peaks     int    `json:"peaks"`
	Forests   int    `json:"forests"`
	Rivers    int    `json:"rivers"`
	Resources bool   `json:"resources"`
}

// decodeImageDataURL decodes a "data:image/...;base64," URL of a PNG, JPEG or GIF picture
func decodeImageDataURL(dataURL string) (image.Image, error) {
	if len(dataURL) > maxImageDataURL {
		return nil, errors.New("the image is too big")
	}
	comma := strings.Index(dataURL, ",")
	if !strings.HasPrefix(dataURL, "data:image/") || comma < 0 || !strings.HasSuffix(dataURL[:comma], ";base64") {
		return nil, errors.New("the image must be a base64 data URL")
	}
	data, err := base64.StdEncoding.DecodeString(dataURL[comma+1:])
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("can not read the image: %w", err)
	}
	return img, nil
}

type rgb struct{ r, g, b float64 }

// sampleImage scales a picture to w x h plots: the average color of the part of the picture over each plot,
// indexed by y*w+x with y growing northwards (the top row of the picture is the north)
func sampleImage(img image.Image, w, h int) []rgb {
	bounds := img.Bounds()
	iw, ih := bounds.Dx(), bounds.Dy()
	result := make([]rgb, w*h)
	for y := 0; y < h; y++ {
		row := h - 1 - y
		y0, y1 := row*ih/h, max(row*ih/h+1, (row+1)*ih/h)
		for x := 0; x < w; x++ {
			x0, x1 := x*iw/w, max(x*iw/w+1, (x+1)*iw/w)
			// Up to 8x8 samples per plot are enough for an average
			sx, sy := max(1, (x1-x0)/8), max(1, (y1-y0)/8)
			var sum rgb
			count := 0.0
			for py := y0; py < y1; py += sy {
				for px := x0; px < x1; px += sx {
					r, g, b, _ := img.At(bounds.Min.X+px, bounds.Min.Y+py).RGBA()
					sum.r += float64(r >> 8)
					sum.g += float64(g >> 8)
					sum.b += float64(b >> 8)
					count++
				}
			}
			result[y*w+x] = rgb{sum.r / count, sum.g / count, sum.b / count}
		}
	}
	return result
}

// paletteEntry is a color of a map picture and what it means
type paletteEntry struct {
	color    rgb
	terrain  string
	plotType uint
	feature  string
}

// mapPalette are the usual colors of map pictures: water in blues, lowlands in greens and yellows,
// hills in browns, mountains in grays and snow in white
var mapPalette = []paletteEntry{
	{rgb{15, 40, 100}, "TERRAIN_OCEAN", PlotOcean, ""},
	{rgb{40, 90, 180}, "TERRAIN_OCEAN", PlotOcean, ""},
	{rgb{90, 160, 220}, "TERRAIN_COAST", PlotOcean, ""},
	{rgb{160, 200, 235}, "TERRAIN_COAST", PlotOcean, ""},
	{rgb{80, 160, 60}, "TERRAIN_GRASS", PlotLand, ""},
	{rgb{120, 190, 90}, "TERRAIN_GRASS", PlotLand, ""},
	{rgb{30, 95, 40}, "TERRAIN_GRASS", PlotLand, "FEATURE_FOREST"},
	{rgb{170, 170, 80}, "TERRAIN_PLAINS", PlotLand, ""},
	{rgb{230, 210, 140}, "TERRAIN_DESERT", PlotLand, ""},
	{rgb{200, 170, 100}, "TERRAIN_DESERT", PlotLand, ""},
	{rgb{150, 150, 125}, "TERRAIN_TUNDRA", PlotLand, ""},
	{rgb{245, 245, 245}, "TERRAIN_SNOW", PlotLand, ""},
	{rgb{140, 105, 65}, "TERRAIN_PLAINS", PlotHills, ""},
	{rgb{105, 100, 95}, "TERRAIN_TUNDRA", PlotPeak, ""},
}

func nearestPalette(c rgb) paletteEntry {
	best, bestDistance := mapPalette[0], math.Inf(1)
	for _, p := range mapPalette {
		dr, dg, db := c.r-p.color.r, c.g-p.color.g, c.b-p.color.b
		// Weighted like the eye sees differences
		if d := 2*dr*dr + 4*dg*dg + 3*db*db; d < bestDistance {
			best, bestDistance = p, d
		}
	}
	return best
}

// ImportImage replaces the terrain of all plots with a picture, see ImageImportOptions.
// Cities, units, signs and start positions stay.
func (m *WbMap) ImportImage(img image.Image, o ImageImportOptions, data *GameData) (*TerrainResult, error) {
	if m.Map == nil || m.Map.GridWidth == 0 || m.Map.GridHeight == 0 {
		return nil, errors.New("the map has no size")
	}
	g := terrainGrid{w: int(m.Map.GridWidth), h: int(m.Map.GridHeight), wrapX: m.Map.WrapX != 0}
	plots, err := m.plotGrid(g)
	if err != nil {
		return nil, err
	}
	terrain := TerrainOptions{Seed: o.Seed, Land: o.Land, Continents: 1, Hills: o.Hills, Peaks: o.Peaks,
		Forests: o.Forests, Rivers: o.Rivers, Resources: o.Resources}
	if o.Mode == "colors" {
		terrain.Land = 50 // not used
	}
	if err := terrain.validate(); err != nil {
		return nil, err
	}
	colors := sampleImage(img, g.w, g.h)
	rng := rand.New(rand.NewSource(o.Seed))
	n := g.w * g.h

	switch o.Mode {
	case "height":
		height := make([]float64, n)
		for i, c := range colors {
			v := (0.299*c.r + 0.587*c.g + 0.114*c.b) / 255
			if o.Invert {
				v = 1 - v
			}
			// A tiny noise keeps flat areas of the picture from being cut by a straight threshold
			height[i] = v + rng.Float64()*0.002
		}
		seaLevel := threshold(height, float64(o.Land)/100)
		land := make([]bool, n)
		for i, v := range height {
			land[i] = v > seaLevel
		}
		// Mountains along ridges of a noise on the high parts, so the brightest spot is not one block of peaks
		ridges := fractal(rng, g.w, g.h, math.Max(3, float64(min(g.w, g.h))/8), 3)
		rough := make([]float64, n)
		for i, v := range height {
			r := 1 - math.Abs(ridges(i%g.w, i/g.w)-0.5)*2
			rough[i] = v + r*0.25
		}
		return m.applyTerrain(g, plots, rng, land, height, rough, terrain, data), nil

	case "colors":
		land := make([]bool, n)
		kinds := make([]paletteEntry, n)
		for i, c := range colors {
			kinds[i] = nearestPalette(c)
			land[i] = kinds[i].plotType != PlotOcean
		}
		result := &TerrainResult{}
		for i, p := range plots {
			k := kinds[i]
			p.FeatureType, p.FeatureVariety = nil, nil
			p.BonusType, p.ImprovementType, p.RouteType = "", "", ""
			p.IsNOfRiver, p.IsWOfRiver, p.RiverNSDirection, p.RiverWEDirection = false, false, 0, 0
			p.TerrainType, p.PlotType = k.terrain, k.plotType
			if !land[i] {
				// Coast is next to land, whatever the shade of blue
				p.TerrainType = "TERRAIN_OCEAN"
				x, y := i%g.w, i/g.w
				for dy := -1; dy <= 1 && p.TerrainType == "TERRAIN_OCEAN"; dy++ {
					for dx := -1; dx <= 1; dx++ {
						if j, ok := g.index(x+dx, y+dy); ok && land[j] {
							p.TerrainType = "TERRAIN_COAST"
							break
						}
					}
				}
				continue
			}
			result.Land++
			if k.feature != "" {
				p.FeatureType, p.FeatureVariety = []string{k.feature}, []string{fmt.Sprint(rng.Intn(3))}
			}
		}
		// Rivers run from inland towards the sea: the height is the distance to the nearest water
		result.Rivers = generateRivers(g, rng, plots, land, distanceToWater(g, land), o.Rivers)
		if o.Resources {
			result.Resources = placeResources(g, rng, plots, land, data)
		}
		return result, nil
	}
	return nil, fmt.Errorf("unknown import mode %q", o.Mode)
}

// distanceToWater is the number of steps from every plot to the nearest water plot, 0 for water
func distanceToWater(g terrainGrid, land []bool) []float64 {
	dist := make([]float64, len(land))
	var queue []int
	for i := range land {
		if land[i] {
			dist[i] = math.Inf(1)
		} else {
			queue = append(queue, i)
		}
	}
	for len(queue) > 0 {
		i := queue[0]
		queue = queue[1:]
		x, y := i%g.w, i/g.w
		for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
			if j, ok := g.index(x+d[0], y+d[1]); ok && dist[j] > dist[i]+1 {
				dist[j] = dist[i] + 1
				queue = append(queue, j)
			}
		}
	}
	for i, d := range dist {
		if math.IsInf(d, 1) {
			dist[i] = float64(g.w + g.h) // an island-less map: all land
		}
	}
	return dist
}

// ImportImage replaces the terrain with a picture (a data URL) as one undoable step, see WbMap.ImportImage.
func (a *App) ImportImage(dataURL string, o ImageImportOptions) (*TerrainResult, error) {
	img, err := decodeImageDataURL(dataURL)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return nil, errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	result, err := a.wbMap.ImportImage(img, o, CurrentGameData())
	if err != nil {
		before.restore(a.wbMap)
		a.mu.Unlock()
		return nil, err
	}
	a.history.push(snapshotEntry("import:"+o.Mode, before, a.wbMap.snapshot()))
	a.mu.Unlock()

	a.setDirty(true)
	ConsoleWrite("Imported terrain from an image (%s): %d land plots, %d rivers", o.Mode, result.Land, result.Rivers)
	return result, nil
}
