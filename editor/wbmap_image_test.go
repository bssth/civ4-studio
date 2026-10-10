package editor

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func pngDataURL(t *testing.T, img image.Image) string {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// heightImage is 80x40 pixels: a bright disc (an island) on black, brightest in the middle
func heightImage() image.Image {
	img := image.NewGray(image.Rect(0, 0, 80, 40))
	for y := 0; y < 40; y++ {
		for x := 0; x < 80; x++ {
			dx, dy := float64(x-40)/30, float64(y-20)/15
			v := 1 - (dx*dx + dy*dy)
			if v < 0 {
				v = 0
			}
			img.SetGray(x, y, color.Gray{Y: uint8(v * 255)})
		}
	}
	return img
}

func TestSampleImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	// The top row of the picture is red, the bottom row blue
	for x := 0; x < 4; x++ {
		img.Set(x, 0, color.RGBA{R: 255, A: 255})
		img.Set(x, 1, color.RGBA{B: 255, A: 255})
	}
	c := sampleImage(img, 2, 2)
	if c[0].b != 255 || c[0].r != 0 || c[2].r != 255 {
		t.Errorf("the bottom of the map must be the bottom of the picture: %+v", c)
	}
	big := sampleImage(img, 8, 4)
	if len(big) != 32 || big[31].r != 255 {
		t.Errorf("a picture smaller than the map: %+v", big[31])
	}
}

func TestImportHeightMap(t *testing.T) {
	app := NewApp()
	app.wbMap = terrainTestMap(t, 40, 20)
	before := app.wbMap.ToWbFormat()
	o := ImageImportOptions{Mode: "height", Seed: 1, Land: 30, Hills: 15, Peaks: 5, Forests: 30, Rivers: 3, Resources: true}
	result, err := app.ImportImage(pngDataURL(t, heightImage()), o)
	if err != nil {
		t.Fatal(err)
	}
	m := app.wbMap
	if share := float64(result.Land) / 800; share < 0.28 || share > 0.32 {
		t.Errorf("land share %.2f", share)
	}
	// The middle is land and a peak or hill, the corners are water
	if p := m.plotAt(20, 10); p.PlotType == PlotOcean || p.PlotType == PlotLand {
		t.Errorf("the brightest plot must be high: %+v", p)
	}
	if m.plotAt(0, 0).PlotType != PlotOcean || m.plotAt(39, 19).PlotType != PlotOcean {
		t.Error("the dark corners must be water")
	}
	if s := app.HistoryState(); s.Undo != "import:height" {
		t.Errorf("undo label = %q", s.Undo)
	}
	app.Undo()
	if !bytes.Equal(before, app.wbMap.ToWbFormat()) {
		t.Error("undo must restore the map")
	}

	o.Invert = true
	if _, err := app.ImportImage(pngDataURL(t, heightImage()), o); err != nil {
		t.Fatal(err)
	}
	if app.wbMap.plotAt(20, 10).PlotType != PlotOcean || app.wbMap.plotAt(0, 0).PlotType == PlotOcean {
		t.Error("an inverted height map makes the bright middle water")
	}
}

func TestImportColors(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 40, 20))
	fill := func(x0, x1 int, c color.RGBA) {
		for y := 0; y < 20; y++ {
			for x := x0; x < x1; x++ {
				img.Set(x, y, c)
			}
		}
	}
	fill(0, 10, color.RGBA{30, 70, 160, 255})    // ocean
	fill(10, 20, color.RGBA{75, 155, 55, 255})   // grassland
	fill(20, 25, color.RGBA{235, 215, 145, 255}) // desert
	fill(25, 30, color.RGBA{25, 90, 35, 255})    // forest
	fill(30, 35, color.RGBA{100, 98, 96, 255})   // mountains
	fill(35, 40, color.RGBA{250, 250, 250, 255}) // snow
	wb := terrainTestMap(t, 40, 20)
	wb.Map.WrapX = 0
	result, err := wb.ImportImage(img, ImageImportOptions{Mode: "colors", Seed: 1, Rivers: 2}, nil)
	if err != nil {
		t.Fatal(err)
	}
	check := func(x int, terrain string, plotType uint, feature string) {
		p := wb.plotAt(x, 5)
		f := ""
		if len(p.FeatureType) > 0 {
			f = p.FeatureType[0]
		}
		if p.TerrainType != terrain || p.PlotType != plotType || f != feature {
			t.Errorf("column %d: %s %d %q, want %s %d %q", x, p.TerrainType, p.PlotType, f, terrain, plotType, feature)
		}
	}
	check(2, "TERRAIN_OCEAN", PlotOcean, "")
	check(9, "TERRAIN_COAST", PlotOcean, "")
	check(15, "TERRAIN_GRASS", PlotLand, "")
	check(22, "TERRAIN_DESERT", PlotLand, "")
	check(27, "TERRAIN_GRASS", PlotLand, "FEATURE_FOREST")
	check(32, "TERRAIN_TUNDRA", PlotPeak, "")
	check(37, "TERRAIN_SNOW", PlotLand, "")
	if result.Land != 600 {
		t.Errorf("land = %d", result.Land)
	}
}

func TestImportImageErrors(t *testing.T) {
	app := NewApp()
	app.wbMap = terrainTestMap(t, 10, 10)
	for _, url := range []string{"", "data:text/plain;base64,AAAA", "data:image/png;base64,!!!", "data:image/png;base64,AAAA"} {
		if _, err := app.ImportImage(url, ImageImportOptions{Mode: "height", Land: 30}); err == nil {
			t.Errorf("%q must be refused", url)
		}
	}
	if _, err := app.ImportImage(pngDataURL(t, heightImage()), ImageImportOptions{Mode: "magic", Land: 30}); err == nil {
		t.Error("an unknown mode must be refused")
	}
	if len(app.history.undo) != 0 {
		t.Error("refused imports are not steps")
	}
}
