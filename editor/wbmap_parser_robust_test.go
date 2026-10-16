package editor

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestParseByteOrderMark(t *testing.T) {
	source := "\xEF\xBB\xBF" + cityUnitSignMap
	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	data, replaced := wb.FileBytes()
	if !bytes.HasPrefix(data, utf8BOM) || replaced != 0 {
		t.Error("a file with a byte order mark is saved with it")
	}
	if !bytes.Equal(normalizeLines(data[3:]), normalizeLines([]byte(cityUnitSignMap))) {
		t.Error("the map must not change")
	}
}

func normalizeLines(data []byte) []byte {
	return []byte(strings.Join(normalizeWb(string(data)), "\n"))
}

func TestParseLongLine(t *testing.T) {
	script := strings.Repeat("x", 200_000)
	source := strings.Replace(cityUnitSignMap, "\tSomeModKey=42", "\tScriptData="+script, 1)
	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	if wb.Plots[0].ScriptData != script {
		t.Errorf("script data of %d characters is cut to %d", len(script), len(wb.Plots[0].ScriptData))
	}
}

func TestParseKeepsUnknownParts(t *testing.T) {
	source := strings.NewReplacer(
		"EndGame\n", "EndGame\nBeginModData\n\tBeginModData\n\t\tDepth=2\n\tEndModData\n\tValue=1\nEndModData\n",
		"\tSomeModKey=42\n", "\tSomeModKey=42\n\tBeginStarSystem\n\t\tPlanets=3\n\tEndStarSystem\n\tA free text line\n\tModFlag\n\tFoo=1, Bar=2\n",
		"\tCityPopulation=3\n", "\tCityPopulation=3\n\tisModCapital\n",
	).Replace(cityUnitSignMap) + "ModTrailer=yes\n"
	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	out := string(wb.ToWbFormat())
	for _, want := range []string{
		"EndGame\nBeginModData\nBeginModData\nDepth=2\nEndModData\nValue=1\nEndModData\nBeginPlayer",
		"\tBeginStarSystem\n\tPlanets=3\n\tEndStarSystem\n",
		"\tA free text line\n", "\tModFlag\n", "\tFoo=1, Bar=2\n", "\t\tisModCapital\n",
		"EndSign\nModTrailer=yes\n",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("the output has no %q:\n%s", want, out)
		}
	}
	// The kept parts are read again the same way
	again, err := ParseWbMap(strings.NewReader(out))
	if err != nil || string(again.ToWbFormat()) != out {
		t.Errorf("the second round trip differs: %v", err)
	}
}

func TestParseLatin1(t *testing.T) {
	// Latin-1 bytes as the game writes them: ü is 0xFC, é is 0xE9
	source := strings.NewReplacer("CityName=Rome, the Eternal", "CityName=Z\xFCrich", "Hello, world", "Caf\xE9").Replace(cityUnitSignMap)
	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	if name := wb.Plots[0].Cities[0].CityName; name != "Zürich" {
		t.Errorf("city name = %q", name)
	}
	// Names survive the way to the interface and back
	data, _ := json.Marshal(wb.Plots[0])
	var plot Plot
	if err := json.Unmarshal(data, &plot); err != nil || plot.Cities[0].CityName != "Zürich" {
		t.Errorf("JSON round trip: %q, %v", plot.Cities[0].CityName, err)
	}
	out, replaced := wb.FileBytes()
	if replaced != 0 || !bytes.Contains(out, []byte("CityName=Z\xFCrich")) || !bytes.Contains(out, []byte("caption=Caf\xE9")) {
		t.Errorf("the file must stay Latin-1 (replaced %d)", replaced)
	}
	if problems := wb.Validate(nil); hasProblemCode(problems, "map.notLatin1") {
		t.Error("Latin-1 names are fine")
	}

	wb.Plots[0].Cities[0].CityName = "Москва"
	out, replaced = wb.FileBytes()
	if replaced != 6 || !bytes.Contains(out, []byte("CityName=??????")) {
		t.Errorf("characters out of Latin-1 are saved as '?': %d", replaced)
	}
	if !hasProblemCode(wb.Validate(nil), "map.notLatin1") {
		t.Error("the check must warn about characters the game can not read")
	}
}

func hasProblemCode(problems []Problem, code string) bool {
	for _, p := range problems {
		if p.Code == code {
			return true
		}
	}
	return false
}

func TestParseUTF8StaysUTF8(t *testing.T) {
	source := strings.Replace(cityUnitSignMap, "Rome, the Eternal", "Zürich", 1)
	wb, err := ParseWbMap(strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	wb.Plots[0].Cities[0].CityName = "Москва"
	out, replaced := wb.FileBytes()
	if replaced != 0 || !bytes.Contains(out, []byte("CityName=Москва")) {
		t.Error("a UTF-8 file is saved as UTF-8")
	}
}

func TestParseErrors(t *testing.T) {
	cut := cityUnitSignMap[:strings.Index(cityUnitSignMap, "EndCity")]
	cases := []struct {
		name, source, code, section string
		line                        int
	}{
		{"binary", "CIV4\x00\x01\x02", "notText", "", 0},
		{"empty", "\n\n# only a comment\n", "empty", "", 0},
		{"not a map", "\nHello there\nVersion=11\n", "notMap", "", 2},
		{"bad version", "Version=eleven\n", "badVersion", "", 1},
		{"bad value", strings.Replace(cityUnitSignMap, "x=1,y=2", "x=one,y=2", 1), "badValue", BeginPlot, 38},
		{"cut off", cut, "unclosed", BeginCity, 51},
		{"unclosed mod section", "Version=11\nBeginGame\nEndGame\nBeginModData\nA=1\n", "unclosed", "BeginModData", 4},
		{"no game", "Version=11\nBeginMap\nEndMap\n", "noGame", "", 0},
	}
	for _, c := range cases {
		_, err := ParseWbMap(strings.NewReader(c.source))
		var fe *MapFileError
		if !errors.As(err, &fe) {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if fe.Code != c.code || fe.Section != c.section || fe.Line != c.line {
			t.Errorf("%s: %+v", c.name, fe)
		}
		if !strings.HasPrefix(fe.Error(), "parse error: ") {
			t.Errorf("%s: message %q", c.name, fe.Error())
		}
	}
}

func TestOpenMapErrors(t *testing.T) {
	dir := t.TempDir()
	app := NewApp()
	missing := filepath.Join(dir, "missing.CivBeyondSwordWBSave")
	if err := app.OpenMap(missing); err == nil {
		t.Fatal("a missing file must fail")
	}
	if e := app.TakeOpenError(); e == nil || e.Code != "read" || e.Path != missing {
		t.Errorf("open error = %+v", e)
	}
	if app.TakeOpenError() != nil {
		t.Error("the error is taken once")
	}

	bad := filepath.Join(dir, "bad.CivWarlordsWBSave")
	if err := os.WriteFile(bad, []byte("Version=11\nBeginGame\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.OpenMap(bad); err == nil {
		t.Fatal("a cut off file must fail")
	}
	good := filepath.Join(dir, "good.Civ4WorldBuilderSave")
	if err := os.WriteFile(good, []byte(cityUnitSignMap), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := app.OpenMap(good); err != nil {
		t.Fatal(err)
	}
	if app.TakeOpenError() != nil {
		t.Error("opening a map forgets the previous error")
	}
}

func TestWbSaveExtensions(t *testing.T) {
	for _, path := range []string{"a.CivBeyondSwordWBSave", "b.civwarlordswbsave", "c.Civ4WorldBuilderSave"} {
		if !isWbSave(path) {
			t.Errorf("%s is a WorldBuilder save", path)
		}
	}
	if isWbSave("d.CivBeyondSwordSave") {
		t.Error("a saved game is not a WorldBuilder save")
	}
	if f := saveFilters("x/map.CivWarlordsWBSave"); f[0].Pattern != "*.CivWarlordsWBSave" || len(f) != 3 {
		t.Errorf("filters = %+v", f)
	}
	if f := saveFilters(""); f[0].Pattern != "*"+wbSaveExt {
		t.Errorf("a new map is saved for Beyond the Sword: %+v", f)
	}
}

// Wails takes any returned value of an error type for a failed call, even a nil pointer
func TestBoundMethodsReturnNoErrorTypes(t *testing.T) {
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	app := reflect.TypeOf(NewApp())
	for i := 0; i < app.NumMethod(); i++ {
		m := app.Method(i)
		for j := 0; j < m.Type.NumOut(); j++ {
			if out := m.Type.Out(j); out != errorType && out.Implements(errorType) {
				t.Errorf("%s returns %s, an error type", m.Name, out)
			}
		}
	}
}
