package editor

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// TerrainOptions are the settings of the terrain generator. Percentages are of all plots for Land,
// of land plots for Hills, Peaks and Forests.
type TerrainOptions struct {
	Seed       int64 `json:"seed"`
	Land       int   `json:"land"`
	Continents int   `json:"continents"`
	Hills      int   `json:"hills"`
	Peaks      int   `json:"peaks"`
	Forests    int   `json:"forests"`
	Rivers     int   `json:"rivers"`
	Resources  bool  `json:"resources"`
}

// TerrainResult tells what was generated
type TerrainResult struct {
	Land      int `json:"land"`
	Rivers    int `json:"rivers"`
	Resources int `json:"resources"`
}

func (o *TerrainOptions) validate() error {
	switch {
	case o.Land < 5 || o.Land > 95:
		return errors.New("land must be within 5..95%")
	case o.Continents < 1 || o.Continents > 20:
		return errors.New("continents must be within 1..20")
	case o.Hills < 0 || o.Peaks < 0 || o.Hills+o.Peaks > 80:
		return errors.New("hills and peaks must be within 0..80% of land together")
	case o.Forests < 0 || o.Forests > 100:
		return errors.New("forests must be within 0..100%")
	case o.Rivers < 0 || o.Rivers > 500:
		return errors.New("rivers must be within 0..500")
	}
	return nil
}

// periodicNoise is value noise that repeats every width plots along x, so a wrapping map has no seam
type periodicNoise struct {
	cells, rows int
	values      []float64
}

func newPeriodicNoise(rng *rand.Rand, width, height int, scale float64) *periodicNoise {
	cells := max(1, int(math.Round(float64(width)/scale)))
	rows := max(1, int(math.Round(float64(height)/scale))) + 1
	n := &periodicNoise{cells: cells, rows: rows, values: make([]float64, cells*(rows+1))}
	for i := range n.values {
		n.values[i] = rng.Float64()
	}
	return n
}

func smooth(t float64) float64 { return t * t * (3 - 2*t) }

// at returns the noise for x in 0..1 (one period) and y in 0..1
func (n *periodicNoise) at(x, y float64) float64 {
	fx, fy := x*float64(n.cells), y*float64(n.rows-1)
	x0, y0 := int(math.Floor(fx)), int(math.Floor(fy))
	tx, ty := smooth(fx-float64(x0)), smooth(fy-float64(y0))
	v := func(ix, iy int) float64 {
		ix = ((ix % n.cells) + n.cells) % n.cells
		iy = min(max(iy, 0), n.rows)
		return n.values[iy*n.cells+ix]
	}
	top := v(x0, y0)*(1-tx) + v(x0+1, y0)*tx
	bottom := v(x0, y0+1)*(1-tx) + v(x0+1, y0+1)*tx
	return top*(1-ty) + bottom*ty
}

// fractal sums octaves of noise from big shapes to small details, the result is about 0..1
func fractal(rng *rand.Rand, width, height int, scale float64, octaves int) func(x, y int) float64 {
	layers := make([]*periodicNoise, octaves)
	for i := range layers {
		layers[i] = newPeriodicNoise(rng, width, height, scale/math.Pow(2, float64(i)))
	}
	return func(x, y int) float64 {
		sum, weight, amp := 0.0, 0.0, 1.0
		for _, l := range layers {
			sum += amp * l.at((float64(x)+0.5)/float64(width), (float64(y)+0.5)/float64(height))
			weight += amp
			amp /= 2
		}
		return sum / weight
	}
}

type terrainGrid struct {
	w, h  int
	wrapX bool
}

func (g terrainGrid) index(x, y int) (int, bool) {
	if g.wrapX {
		x = ((x % g.w) + g.w) % g.w
	}
	if x < 0 || y < 0 || x >= g.w || y >= g.h {
		return 0, false
	}
	return y*g.w + x, true
}

// distance between two plots, across the seam of a wrapping map
func (g terrainGrid) distance(x0, y0, x1, y1 int) float64 {
	dx := math.Abs(float64(x0 - x1))
	if g.wrapX {
		dx = math.Min(dx, float64(g.w)-dx)
	}
	return math.Hypot(dx, float64(y0-y1))
}

// threshold returns the value above which share of values lie
func threshold(values []float64, share float64) float64 {
	sorted := append([]float64(nil), values...)
	sort.Float64s(sorted)
	i := int(float64(len(sorted)) * (1 - share))
	return sorted[min(max(i, 0), len(sorted)-1)]
}

// GenerateTerrain replaces terrain, height, features, rivers, resources and improvements of all plots.
// Cities, units, signs and start positions stay. The same options give the same terrain.
func (m *WbMap) GenerateTerrain(o TerrainOptions, data *GameData) (*TerrainResult, error) {
	if err := o.validate(); err != nil {
		return nil, err
	}
	if m.Map == nil || m.Map.GridWidth == 0 || m.Map.GridHeight == 0 {
		return nil, errors.New("the map has no size")
	}
	g := terrainGrid{w: int(m.Map.GridWidth), h: int(m.Map.GridHeight), wrapX: m.Map.WrapX != 0}
	if len(m.Plots) != g.w*g.h {
		return nil, fmt.Errorf("the map has %d plots instead of %d, create the plots first", len(m.Plots), g.w*g.h)
	}
	rng := rand.New(rand.NewSource(o.Seed))
	n := g.w * g.h

	// Height: fractal noise plus round continents, lowered near the poles. Continent centers are spread:
	// each one is the farthest of a few random candidates from the centers placed before.
	noise := fractal(rng, g.w, g.h, math.Max(4, float64(min(g.w, g.h))/4), 6)
	type center struct{ x, y, r float64 }
	centers := make([]center, 0, o.Continents)
	for len(centers) < o.Continents {
		r := math.Sqrt(float64(n)*float64(o.Land)/100/float64(o.Continents)/math.Pi) * (0.8 + rng.Float64()*0.4)
		best, bestDistance := center{}, -1.0
		for k := 0; k < 12; k++ {
			c := center{x: rng.Float64() * float64(g.w), y: float64(g.h) * (0.2 + rng.Float64()*0.6), r: r}
			d := math.Inf(1)
			for _, o := range centers {
				d = math.Min(d, g.distance(int(c.x), int(c.y), int(o.x), int(o.y))-o.r)
			}
			if d > bestDistance {
				best, bestDistance = c, d
			}
		}
		centers = append(centers, best)
	}
	height := make([]float64, n)
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			shape := 0.0
			for _, c := range centers {
				d := g.distance(x, y, int(c.x), int(c.y)) / (c.r * 1.25)
				shape = math.Max(shape, math.Max(0, 1-d*d))
			}
			edge := math.Min(float64(y), float64(g.h-1-y)) / float64(g.h)
			polar := math.Min(1, edge*8)
			height[y*g.w+x] = (noise(x, y)*0.6 + shape*0.5) * (0.4 + 0.6*polar)
		}
	}
	seaLevel := threshold(height, float64(o.Land)/100)
	land := make([]bool, n)
	// Mountains follow ridges of their own noise instead of the highest land, so they form chains
	ridges := fractal(rng, g.w, g.h, math.Max(3, float64(min(g.w, g.h))/6), 4)
	rough := make([]float64, n)
	for i, v := range height {
		r := 1 - math.Abs(ridges(i%g.w, i/g.w)-0.5)*2
		rough[i] = r*0.8 + rng.Float64()*0.2
		land[i] = v > seaLevel
	}
	plots, err := m.plotGrid(g)
	if err != nil {
		return nil, err
	}
	return m.applyTerrain(g, plots, rng, land, height, rough, o, data), nil
}

// plotGrid returns the plots indexed by y*width+x, an error if one is missing
func (m *WbMap) plotGrid(g terrainGrid) ([]*Plot, error) {
	plots := make([]*Plot, g.w*g.h)
	for _, p := range m.Plots {
		if i, ok := g.index(int(p.X), int(p.Y)); ok {
			plots[i] = p
		}
	}
	for i, p := range plots {
		if p == nil {
			return nil, fmt.Errorf("there is no plot %d,%d", i%g.w, i/g.w)
		}
	}
	return plots, nil
}

// applyTerrain makes the plots from a land mask: hills and peaks where rough is highest, terrain by the latitude
// of the map and a moisture noise, ice in polar seas, forests, jungle and oases, then rivers running down height
// and resources. It replaces everything of the plots but cities, units and start positions.
func (m *WbMap) applyTerrain(g terrainGrid, plots []*Plot, rng *rand.Rand, land []bool, height, rough []float64,
	o TerrainOptions, data *GameData) *TerrainResult {
	var landRough []float64
	for i := range land {
		if land[i] {
			landRough = append(landRough, rough[i])
		}
	}
	peakLevel, hillLevel := math.Inf(1), math.Inf(1)
	if len(landRough) > 0 {
		if o.Peaks > 0 {
			peakLevel = threshold(landRough, float64(o.Peaks)/100)
		}
		if o.Hills+o.Peaks > 0 {
			hillLevel = threshold(landRough, float64(o.Hills+o.Peaks)/100)
		}
	}
	// Moisture decides between grassland and plains, desert and forests; climate zones get ragged borders
	moisture := fractal(rng, g.w, g.h, math.Max(3, float64(min(g.w, g.h))/6), 4)
	jitter := fractal(rng, g.w, g.h, math.Max(3, float64(min(g.w, g.h))/8), 3)
	top, bottom := float64(m.Map.TopLatitude), float64(m.Map.BottomLatitude)
	rowLatitude := func(y int) float64 { return math.Abs(bottom + (top-bottom)*(float64(y)+0.5)/float64(g.h)) }

	nearWater := func(x, y int) bool {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if i, ok := g.index(x+dx, y+dy); ok && !land[i] {
					return true
				}
			}
		}
		return false
	}
	nearLand := func(x, y int) bool {
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if i, ok := g.index(x+dx, y+dy); ok && land[i] {
					return true
				}
			}
		}
		return false
	}

	result := &TerrainResult{}
	for y := 0; y < g.h; y++ {
		for x := 0; x < g.w; x++ {
			lat := rowLatitude(y) + (jitter(x, y)-0.5)*16
			i := y*g.w + x
			p := plots[i]
			p.FeatureType, p.FeatureVariety = nil, nil
			p.BonusType, p.ImprovementType, p.RouteType = "", "", ""
			p.IsNOfRiver, p.IsWOfRiver, p.RiverNSDirection, p.RiverWEDirection = false, false, 0, 0
			wet := moisture(x, y) + (rng.Float64()-0.5)*0.12
			if !land[i] {
				p.PlotType = PlotOcean
				p.TerrainType = "TERRAIN_OCEAN"
				if nearLand(x, y) {
					p.TerrainType = "TERRAIN_COAST"
				}
				if lat > 72 && wet > 0.35 {
					p.FeatureType, p.FeatureVariety = []string{"FEATURE_ICE"}, []string{"0"}
				}
				continue
			}
			result.Land++
			switch v := rough[i]; {
			case v > peakLevel:
				p.PlotType = PlotPeak
			case v > hillLevel:
				p.PlotType = PlotHills
			default:
				p.PlotType = PlotLand
			}
			switch {
			case lat > 70:
				p.TerrainType = "TERRAIN_SNOW"
			case lat > 58:
				p.TerrainType = "TERRAIN_TUNDRA"
			case lat > 12 && lat < 38 && wet < 0.42:
				p.TerrainType = "TERRAIN_DESERT"
			case wet < 0.5:
				p.TerrainType = "TERRAIN_PLAINS"
			default:
				p.TerrainType = "TERRAIN_GRASS"
			}
			if p.PlotType == PlotPeak {
				continue
			}
			// Forests where it is wet enough; jungle near the equator
			forestChance := float64(o.Forests) / 100 * (0.4 + wet)
			switch p.TerrainType {
			case "TERRAIN_DESERT":
				if rng.Float64() < 0.04 && p.PlotType == PlotLand && !nearWater(x, y) {
					p.FeatureType, p.FeatureVariety = []string{"FEATURE_OASIS"}, []string{"0"}
				}
			case "TERRAIN_SNOW":
			default:
				if rng.Float64() < forestChance {
					feature := "FEATURE_FOREST"
					if lat < 18 && p.TerrainType == "TERRAIN_GRASS" && p.PlotType == PlotLand {
						feature = "FEATURE_JUNGLE"
					}
					p.FeatureType, p.FeatureVariety = []string{feature}, []string{fmt.Sprint(rng.Intn(3))}
				}
			}
		}
	}

	result.Rivers = generateRivers(g, rng, plots, land, height, o.Rivers)
	if o.Resources {
		result.Resources = placeResources(g, rng, plots, land, data)
	}
	return result
}

// generateRivers runs rivers along plot edges from high land downhill to the sea. Edges are walked between
// plot corners; a corner x, y is the south-western corner of plot x, y. Only rivers that reach water are kept.
func generateRivers(g terrainGrid, rng *rand.Rand, plots []*Plot, land []bool, height []float64, count int) int {
	// Plots around a corner: south-west, south-east, north-west, north-east
	around := func(cx, cy int) [4]int {
		var r [4]int
		for k, d := range [4][2]int{{-1, -1}, {0, -1}, {-1, 0}, {0, 0}} {
			if i, ok := g.index(cx+d[0], cy+d[1]); ok {
				r[k] = i
			} else {
				r[k] = -1
			}
		}
		return r
	}
	cornerHeight := func(cx, cy int) (float64, bool) {
		sum, wet := 0.0, false
		for _, i := range around(cx, cy) {
			if i < 0 || !land[i] {
				wet = true
				continue
			}
			sum += height[i]
		}
		return sum / 4, wet
	}
	used := make(map[[2]int]bool)
	var sources []int
	for i := range land {
		if land[i] && plots[i].PlotType == PlotHills {
			sources = append(sources, i)
		}
	}
	if len(sources) == 0 {
		for i := range land {
			if land[i] {
				sources = append(sources, i)
			}
		}
	}
	made := 0
	for attempt := 0; made < count && attempt < count*20 && len(sources) > 0; attempt++ {
		start := sources[rng.Intn(len(sources))]
		cx, cy := start%g.w+rng.Intn(2), start/g.w+rng.Intn(2)
		if _, wet := cornerHeight(cx, cy); wet || used[[2]int{cx, cy}] {
			continue
		}
		type step struct{ cx, cy, dx, dy int }
		var path []step
		visited := map[[2]int]bool{{cx, cy}: true}
		reached := false
		for len(path) < g.w+g.h {
			best, bestHeight := -1, math.Inf(1)
			moves := [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}}
			for k, d := range moves {
				nx, ny := cx+d[0], cy+d[1]
				if g.wrapX {
					nx = ((nx % g.w) + g.w) % g.w
				}
				if nx < 0 || ny < 0 || nx > g.w || ny > g.h || visited[[2]int{nx, ny}] || used[[2]int{nx, ny}] {
					continue
				}
				// Both plots along the edge must be land, or the river would run along the coast
				if !edgeOnLand(g, land, cx, cy, d[0], d[1]) {
					continue
				}
				hgt, _ := cornerHeight(nx, ny)
				hgt += rng.Float64() * 0.02
				if hgt < bestHeight {
					best, bestHeight = k, hgt
				}
			}
			if best < 0 {
				break
			}
			d := moves[best]
			path = append(path, step{cx, cy, d[0], d[1]})
			cx, cy = cx+d[0], cy+d[1]
			if g.wrapX {
				cx = ((cx % g.w) + g.w) % g.w
			}
			visited[[2]int{cx, cy}] = true
			if _, wet := cornerHeight(cx, cy); wet {
				reached = true
				break
			}
		}
		if !reached || len(path) < 3 {
			continue
		}
		for _, s := range path {
			markRiverEdge(g, plots, s.cx, s.cy, s.dx, s.dy)
			used[[2]int{s.cx, s.cy}] = true
		}
		made++
	}
	return made
}

// edgeOnLand tells if both plots along the edge from corner cx, cy in direction dx, dy are land
func edgeOnLand(g terrainGrid, land []bool, cx, cy, dx, dy int) bool {
	var a, b [2]int
	switch {
	case dx != 0: // horizontal edge: plots below and above
		x := cx
		if dx < 0 {
			x = cx - 1
		}
		a, b = [2]int{x, cy - 1}, [2]int{x, cy}
	default: // vertical edge: plots on the left and right
		y := cy
		if dy < 0 {
			y = cy - 1
		}
		a, b = [2]int{cx - 1, y}, [2]int{cx, y}
	}
	i, okA := g.index(a[0], a[1])
	j, okB := g.index(b[0], b[1])
	return okA && okB && land[i] && land[j]
}

// markRiverEdge sets the river flag of the plot that owns the edge: a horizontal edge is the southern edge
// of the plot above it (isNOfRiver), a vertical edge the eastern edge of the plot on its left (isWOfRiver)
func markRiverEdge(g terrainGrid, plots []*Plot, cx, cy, dx, dy int) {
	if dx != 0 {
		x := cx
		if dx < 0 {
			x = cx - 1
		}
		if i, ok := g.index(x, cy); ok {
			plots[i].IsNOfRiver = true
			plots[i].RiverWEDirection = 1 // east
			if dx < 0 {
				plots[i].RiverWEDirection = 3 // west
			}
		}
		return
	}
	y := cy
	if dy < 0 {
		y = cy - 1
	}
	if i, ok := g.index(cx-1, y); ok {
		plots[i].IsWOfRiver = true
		plots[i].RiverNSDirection = 0 // north
		if dy < 0 {
			plots[i].RiverNSDirection = 2 // south
		}
	}
}

// resourceRule places a resource of Beyond the Sword on fitting plots
type resourceRule struct {
	bonus    string
	terrains []string
	water    bool
	hills    bool
	feature  string
	weight   int
}

var resourceRules = []resourceRule{
	{bonus: "BONUS_FISH", water: true, terrains: []string{"TERRAIN_COAST"}, weight: 4},
	{bonus: "BONUS_CLAM", water: true, terrains: []string{"TERRAIN_COAST"}, weight: 2},
	{bonus: "BONUS_CRAB", water: true, terrains: []string{"TERRAIN_COAST"}, weight: 2},
	{bonus: "BONUS_WHALE", water: true, terrains: []string{"TERRAIN_OCEAN", "TERRAIN_COAST"}, weight: 1},
	{bonus: "BONUS_WHEAT", terrains: []string{"TERRAIN_PLAINS"}, weight: 3},
	{bonus: "BONUS_CORN", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS"}, weight: 3},
	{bonus: "BONUS_RICE", terrains: []string{"TERRAIN_GRASS"}, weight: 2},
	{bonus: "BONUS_COW", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS"}, weight: 3},
	{bonus: "BONUS_PIG", terrains: []string{"TERRAIN_GRASS"}, weight: 2},
	{bonus: "BONUS_SHEEP", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS"}, hills: true, weight: 2},
	{bonus: "BONUS_HORSE", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_TUNDRA"}, weight: 3},
	{bonus: "BONUS_DEER", terrains: []string{"TERRAIN_TUNDRA"}, feature: "FEATURE_FOREST", weight: 2},
	{bonus: "BONUS_FUR", terrains: []string{"TERRAIN_TUNDRA", "TERRAIN_SNOW"}, feature: "FEATURE_FOREST", weight: 2},
	{bonus: "BONUS_IVORY", terrains: []string{"TERRAIN_PLAINS", "TERRAIN_GRASS"}, weight: 1},
	{bonus: "BONUS_BANANA", terrains: []string{"TERRAIN_GRASS"}, feature: "FEATURE_JUNGLE", weight: 2},
	{bonus: "BONUS_DYE", terrains: []string{"TERRAIN_GRASS"}, feature: "FEATURE_JUNGLE", weight: 1},
	{bonus: "BONUS_SPICES", terrains: []string{"TERRAIN_GRASS"}, feature: "FEATURE_JUNGLE", weight: 1},
	{bonus: "BONUS_SILK", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS"}, feature: "FEATURE_FOREST", weight: 1},
	{bonus: "BONUS_WINE", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS"}, weight: 1},
	{bonus: "BONUS_INCENSE", terrains: []string{"TERRAIN_DESERT", "TERRAIN_PLAINS"}, weight: 1},
	{bonus: "BONUS_IRON", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_DESERT", "TERRAIN_TUNDRA"}, hills: true, weight: 2},
	{bonus: "BONUS_COPPER", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_DESERT", "TERRAIN_TUNDRA"}, hills: true, weight: 2},
	{bonus: "BONUS_STONE", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_DESERT"}, weight: 1},
	{bonus: "BONUS_MARBLE", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_DESERT"}, hills: true, weight: 1},
	{bonus: "BONUS_GOLD", terrains: []string{"TERRAIN_PLAINS", "TERRAIN_DESERT"}, hills: true, weight: 1},
	{bonus: "BONUS_SILVER", terrains: []string{"TERRAIN_TUNDRA", "TERRAIN_SNOW"}, hills: true, weight: 1},
	{bonus: "BONUS_GEMS", terrains: []string{"TERRAIN_GRASS"}, hills: true, weight: 1},
	{bonus: "BONUS_OIL", terrains: []string{"TERRAIN_DESERT", "TERRAIN_TUNDRA", "TERRAIN_SNOW"}, weight: 1},
	{bonus: "BONUS_COAL", terrains: []string{"TERRAIN_GRASS", "TERRAIN_PLAINS", "TERRAIN_TUNDRA"}, hills: true, weight: 1},
	{bonus: "BONUS_ALUMINUM", terrains: []string{"TERRAIN_DESERT", "TERRAIN_PLAINS"}, hills: true, weight: 1},
	{bonus: "BONUS_URANIUM", terrains: []string{"TERRAIN_DESERT", "TERRAIN_TUNDRA", "TERRAIN_PLAINS"}, weight: 1},
	{bonus: "BONUS_SUGAR", terrains: []string{"TERRAIN_GRASS"}, weight: 1},
}

// fits tells if a rule allows a resource on a plot
func (r resourceRule) fits(p *Plot) bool {
	if r.water != (p.PlotType == PlotOcean) || p.PlotType == PlotPeak {
		return false
	}
	if r.hills && p.PlotType != PlotHills {
		return false
	}
	if r.feature != "" && (len(p.FeatureType) == 0 || p.FeatureType[0] != r.feature) {
		return false
	}
	if r.feature == "" && len(p.FeatureType) > 0 && p.FeatureType[0] == "FEATURE_ICE" {
		return false
	}
	for _, t := range r.terrains {
		if t == p.TerrainType {
			return true
		}
	}
	return false
}

// placeResources puts about one resource on every ninth land plot and on a few coastal ones,
// keeping two plots between resources. With game data only resources defined there are used.
func placeResources(g terrainGrid, rng *rand.Rand, plots []*Plot, land []bool, data *GameData) int {
	var rules []resourceRule
	for _, r := range resourceRules {
		if data == nil || data.Table(InfoBonuses).Len() == 0 || data.Table(InfoBonuses).Get(r.bonus) != nil {
			rules = append(rules, r)
		}
	}
	if len(rules) == 0 {
		return 0
	}
	order := rng.Perm(len(plots))
	placed := 0
	taken := make([]bool, len(plots))
	for _, i := range order {
		p := plots[i]
		chance := 1.0 / 9
		if !land[i] {
			chance = 1.0 / 14
			if p.TerrainType != "TERRAIN_COAST" {
				chance = 1.0 / 60
			}
		}
		if rng.Float64() > chance || taken[i] {
			continue
		}
		total := 0
		var fitting []resourceRule
		for _, r := range rules {
			if r.fits(p) {
				fitting = append(fitting, r)
				total += r.weight
			}
		}
		if total == 0 {
			continue
		}
		pick := rng.Intn(total)
		for _, r := range fitting {
			if pick -= r.weight; pick < 0 {
				p.BonusType = r.bonus
				break
			}
		}
		placed++
		x, y := i%g.w, i/g.w
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if j, ok := g.index(x+dx, y+dy); ok {
					taken[j] = true
				}
			}
		}
	}
	return placed
}

// PlaceStarts puts the start positions of all players with a civilization on good land as far from each
// other as possible: flat or hilly grassland and plains, preferably at the coast. Returns the moved players.
func (m *WbMap) PlaceStarts(seed int64) ([]int, error) {
	if m.Map == nil || m.Map.GridWidth == 0 {
		return nil, errors.New("the map has no size")
	}
	g := terrainGrid{w: int(m.Map.GridWidth), h: int(m.Map.GridHeight), wrapX: m.Map.WrapX != 0}
	land := make([]bool, g.w*g.h)
	type candidate struct{ x, y int }
	var good []candidate
	for _, p := range m.Plots {
		if i, ok := g.index(int(p.X), int(p.Y)); ok {
			land[i] = p.PlotType != PlotOcean
		}
	}
	for _, p := range m.Plots {
		if (p.PlotType != PlotLand && p.PlotType != PlotHills) ||
			(p.TerrainType != "TERRAIN_GRASS" && p.TerrainType != "TERRAIN_PLAINS") || len(p.Cities) > 0 {
			continue
		}
		good = append(good, candidate{int(p.X), int(p.Y)})
	}
	var players []int
	for i, p := range m.Players {
		if !isEmptySlot(p) {
			players = append(players, i)
		}
	}
	if len(players) == 0 {
		return nil, errors.New("there are no players with a civilization")
	}
	if len(good) < len(players) {
		return nil, fmt.Errorf("only %d plots of grassland or plains for %d players", len(good), len(players))
	}
	// Land around a plot makes a better start: count land plots in the city radius
	score := func(c candidate) float64 {
		s := 0.0
		coast := false
		for dy := -2; dy <= 2; dy++ {
			for dx := -2; dx <= 2; dx++ {
				if j, ok := g.index(c.x+dx, c.y+dy); ok {
					if land[j] {
						s++
					} else if abs(dx) <= 1 && abs(dy) <= 1 {
						coast = true
					}
				}
			}
		}
		if coast {
			s += 4
		}
		return s
	}
	rng := rand.New(rand.NewSource(seed))
	sort.SliceStable(good, func(a, b int) bool { return score(good[a]) > score(good[b]) })
	// The better half of the plots, the first start is random among the best
	pool := good[:max(len(players), len(good)/2)]
	chosen := []candidate{pool[rng.Intn(min(len(pool), 10))]}
	for len(chosen) < len(players) {
		best, bestValue := -1, -1.0
		for k, c := range pool {
			d := math.Inf(1)
			for _, s := range chosen {
				d = math.Min(d, g.distance(c.x, c.y, s.x, s.y))
			}
			if v := d + score(c)*0.15; d > 0 && v > bestValue {
				best, bestValue = k, v
			}
		}
		if best < 0 {
			break
		}
		chosen = append(chosen, pool[best])
	}
	for k, i := range players {
		c := chosen[k%len(chosen)]
		p := m.Players[i]
		p.StartingX, p.StartingY, p.RandomStartLocation = c.x, c.y, false
	}
	return players, nil
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// GenerateTerrain fills the map with generated terrain as one undoable step, see WbMap.GenerateTerrain.
func (a *App) GenerateTerrain(o TerrainOptions) (*TerrainResult, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return nil, errors.New("no map loaded")
	}
	before := a.wbMap.snapshot()
	result, err := a.wbMap.GenerateTerrain(o, CurrentGameData())
	if err != nil {
		before.restore(a.wbMap)
		a.mu.Unlock()
		return nil, err
	}
	a.history.push(snapshotEntry(fmt.Sprintf("generate:%d", o.Seed), before, a.wbMap.snapshot()))
	a.mu.Unlock()

	a.setDirty(true)
	ConsoleWrite("Generated terrain with seed %d: %d land plots, %d rivers, %d resources", o.Seed, result.Land, result.Rivers, result.Resources)
	return result, nil
}

// PlaceStarts moves the start positions of all players as one undoable step, see WbMap.PlaceStarts.
func (a *App) PlaceStarts(seed int64) (int, error) {
	a.mu.Lock()
	if a.wbMap == nil {
		a.mu.Unlock()
		return 0, errors.New("no map loaded")
	}
	before := cloneValue(a.wbMap.Players)
	players, err := a.wbMap.PlaceStarts(seed)
	if err != nil {
		a.mu.Unlock()
		return 0, err
	}
	a.history.push(sectionEntry("starts", before, a.wbMap.Players, func(m *WbMap, v []*Player) { m.Players = v }))
	a.mu.Unlock()

	a.setDirty(true)
	return len(players), nil
}
