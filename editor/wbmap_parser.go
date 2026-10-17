package editor

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	versionPrefix  = "Version="
	defaultVersion = 11
)

const (
	BeginGame   = "BeginGame"
	EndGame     = "EndGame"
	BeginPlayer = "BeginPlayer"
	EndPlayer   = "EndPlayer"
	NonePlayer  = "NONE"
	BeginTeam   = "BeginTeam"
	EndTeam     = "EndTeam"
	BeginMap    = "BeginMap"
	EndMap      = "EndMap"
	BeginPlot   = "BeginPlot"
	EndPlot     = "EndPlot"
	BeginUnit   = "BeginUnit"
	EndUnit     = "EndUnit"
	BeginCity   = "BeginCity"
	EndCity     = "EndCity"
	BeginSign   = "BeginSign"
	EndSign     = "EndSign"
)

const (
	stateGlobal = iota
	stateInsideGame
	stateInsideTeam
	stateInsidePlayer
	stateInsideMap
	stateInsidePlot
	stateInsideCity
	stateInsideUnit
	stateInsideSign
)

// sections are the Begin and End markers of the parser states
var sections = map[int][2]string{
	stateInsideGame:   {BeginGame, EndGame},
	stateInsideTeam:   {BeginTeam, EndTeam},
	stateInsidePlayer: {BeginPlayer, EndPlayer},
	stateInsideMap:    {BeginMap, EndMap},
	stateInsidePlot:   {BeginPlot, EndPlot},
	stateInsideCity:   {BeginCity, EndCity},
	stateInsideUnit:   {BeginUnit, EndUnit},
	stateInsideSign:   {BeginSign, EndSign},
}

// textKeys hold free text that may contain commas, so the value spans up to the end of the line
var textKeys = []string{
	"Description", "CivDesc", "CivShortDesc", "CivAdjective", "LeaderName",
	"CityName", "CityList", "Landmark", "ScriptData", "caption", "TeamReveal",
}

// Encodings of the text of a WorldBuilder save. The game reads and writes Latin-1 (CvWBDesc.py),
// other tools may write UTF-8.
const (
	encodingLatin1 = ""
	encodingUTF8   = "utf-8"
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// rawBlock is a part of a file the editor does not know, e.g. a section added by a mod. It is written back
// as it is, after the known sections of the kind it followed ("" for the beginning of the file).
type rawBlock struct {
	after string
	lines []string
}

// MapFileError tells why a file can not be read as a WorldBuilder save. Code and the fields let the
// frontend explain it in the interface language:
//   - notText: a binary file, e.g. a saved game instead of a WorldBuilder save
//   - empty: there is nothing in the file
//   - notMap: the first line is not "Version=..." or a section (Text)
//   - badVersion: the version number (Text) is not a number
//   - badValue: a line (Text) of a section has a value the editor can not read (Detail)
//   - unclosed: the section that starts at Line is not closed, the file is cut off
//   - noGame: there is no BeginGame section
//   - read: the file can not be read at all (Detail)
type MapFileError struct {
	Code    string `json:"code"`
	Line    int    `json:"line"`
	Text    string `json:"text"`
	Section string `json:"section"`
	Detail  string `json:"detail"`
	Path    string `json:"path"`
}

func (e *MapFileError) Error() string {
	var msg string
	switch e.Code {
	case "notText":
		msg = "the file is not a text WorldBuilder save; saved games can not be opened"
	case "empty":
		msg = "the file is empty"
	case "notMap":
		msg = fmt.Sprintf("the file does not look like a WorldBuilder save, line %d is '%s'", e.Line, e.Text)
	case "badVersion":
		msg = fmt.Sprintf("bad version number '%s' (at line %d)", e.Text, e.Line)
	case "badValue":
		msg = fmt.Sprintf("can not read '%s' in %s: %s (at line %d)", e.Text, e.Section, e.Detail, e.Line)
	case "unclosed":
		msg = fmt.Sprintf("%s at line %d is not closed, the file is cut off or damaged", e.Section, e.Line)
	case "noGame":
		msg = "no game info specified (BeginGame)"
	default:
		msg = e.Detail
	}
	return "parse error: " + msg
}

// shorten keeps error messages readable when a line is very long
func shorten(s string) string {
	const limit = 120
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	return string([]rune(s)[:limit]) + "…"
}

// decodeWbText returns the text of a file and how it is encoded. A UTF-8 byte order mark is dropped;
// text that is not valid UTF-8 is Latin-1, every byte is the character with the same code.
func decodeWbText(data []byte) (text, encoding string, bom bool, err error) {
	if bytes.HasPrefix(data, utf8BOM) {
		data, bom = data[len(utf8BOM):], true
	}
	if bytes.IndexByte(data, 0) >= 0 {
		return "", "", false, &MapFileError{Code: "notText"}
	}
	if utf8.Valid(data) {
		for _, b := range data {
			if b >= utf8.RuneSelf {
				return string(data), encodingUTF8, bom, nil
			}
		}
		if bom {
			return string(data), encodingUTF8, true, nil
		}
		return string(data), encodingLatin1, false, nil
	}
	var sb strings.Builder
	sb.Grow(len(data) + len(data)/8)
	for _, b := range data {
		sb.WriteRune(rune(b))
	}
	return sb.String(), encodingLatin1, false, nil
}

// FileBytes returns the map as it is written to a file, in the encoding of the file it was read from.
// Characters Latin-1 has not are written as '?', replaced counts them.
func (m *WbMap) FileBytes() (data []byte, replaced int) {
	text := m.ToWbFormat()
	if m.encoding == encodingUTF8 {
		if m.bom {
			return append(append([]byte{}, utf8BOM...), text...), 0
		}
		return text, 0
	}
	out := make([]byte, 0, len(text))
	for _, r := range string(text) {
		if r > 0xFF {
			r = '?'
			replaced++
		}
		out = append(out, byte(r))
	}
	return out, replaced
}

// notLatin1 counts texts of a Latin-1 map with characters the game can not save and returns the first of them
func (m *WbMap) notLatin1() (count int, example string) {
	if m.encoding == encodingUTF8 {
		return 0, ""
	}
	for _, line := range strings.Split(string(m.ToWbFormat()), "\n") {
		for _, r := range line {
			if r > 0xFF {
				if count == 0 {
					example = shorten(strings.TrimSpace(line))
				}
				count++
				break
			}
		}
	}
	return count, example
}

// isBeginMarker tells if a line starts a section, e.g. BeginPlot or BeginStarSystem of a mod
func isBeginMarker(line string) bool {
	return len(line) > len("Begin") && strings.HasPrefix(line, "Begin") && !strings.ContainsAny(line, " =,\t")
}

type wbParser struct {
	wb      *WbMap
	line    int
	state   int
	started bool
	// lines where the open sections start, by state
	starts map[int]int
	// the kind of the last finished top-level section, unknown blocks of the file follow it
	last string

	game     *Game
	mapProps *MapProps
	team     *Team
	player   *Player
	plot     *Plot
	city     *City
	unit     *Unit
	sign     *Sign

	// an unknown section being read: its lines, markers, nesting depth and where it goes
	raw                []string
	rawBegin, rawEnd   string
	rawDepth, rawStart int
	rawTarget          *[]string

	// how the file writes the flags the game reads as words, see WbMap.numericRandomize
	numericRandomize, numericStarts bool
}

func ParseWbMap(reader io.Reader) (*WbMap, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, &MapFileError{Code: "read", Detail: err.Error()}
	}
	text, encoding, bom, err := decodeWbText(data)
	if err != nil {
		return nil, err
	}

	ConsoleWrite("Parsing map contents...")
	p := &wbParser{wb: &WbMap{Version: defaultVersion, encoding: encoding, bom: bom}, starts: map[int]int{}}
	for i, content := range strings.Split(text, "\n") {
		p.line = i + 1
		content = strings.Trim(content, " \t\r")
		if content == "" || strings.HasPrefix(content, "#") {
			continue
		}
		if err := p.parse(content); err != nil {
			return nil, err
		}
	}
	return p.finish()
}

func (p *wbParser) parse(content string) error {
	if p.raw != nil {
		p.raw = append(p.raw, content)
		switch content {
		case p.rawBegin:
			p.rawDepth++
		case p.rawEnd:
			if p.rawDepth--; p.rawDepth == 0 {
				p.closeRaw()
			}
		}
		return nil
	}
	if !p.started {
		p.started = true
		if !strings.HasPrefix(content, versionPrefix) && !isBeginMarker(content) {
			return &MapFileError{Code: "notMap", Line: p.line, Text: shorten(content)}
		}
	}
	if p.state == stateGlobal {
		return p.parseGlobal(content)
	}
	return p.parseInSection(content)
}

func (p *wbParser) parseGlobal(content string) error {
	switch {
	case strings.HasPrefix(content, versionPrefix):
		verStr := strings.TrimSpace(content[len(versionPrefix):])
		v, err := strconv.Atoi(verStr)
		if err != nil {
			return &MapFileError{Code: "badVersion", Line: p.line, Text: shorten(verStr)}
		}
		p.wb.Version = v
	case content == BeginGame:
		p.game = &Game{}
		p.enter(stateInsideGame)
	case content == BeginTeam:
		p.team = &Team{}
		p.enter(stateInsideTeam)
	case content == BeginPlayer:
		p.player = &Player{}
		p.enter(stateInsidePlayer)
	case content == BeginMap:
		p.mapProps = &MapProps{}
		p.enter(stateInsideMap)
	case content == BeginPlot:
		p.plot = &Plot{}
		p.enter(stateInsidePlot)
	case content == BeginSign:
		p.sign = &Sign{}
		p.enter(stateInsideSign)
	case isBeginMarker(content):
		p.startRaw(content, nil)
	default:
		// A line of a mod outside of sections is kept as it is
		p.wb.raw = append(p.wb.raw, rawBlock{after: p.last, lines: []string{content}})
	}
	return nil
}

func (p *wbParser) enter(state int) {
	p.state = state
	p.starts[state] = p.line
}

func (p *wbParser) parseInSection(content string) error {
	if content == sections[p.state][1] {
		p.leave()
		return nil
	}
	if p.state == stateInsidePlot && content == BeginCity {
		p.city = &City{}
		p.enter(stateInsideCity)
		return nil
	}
	if p.state == stateInsidePlot && content == BeginUnit {
		p.unit = &Unit{}
		p.enter(stateInsideUnit)
		return nil
	}
	extra := p.extra()
	if isBeginMarker(content) {
		// A section of a mod inside of a known one stays inside of it
		p.startRaw(content, extra)
		return nil
	}

	parsed, bare, ok := parseLine(content)
	if !ok {
		*extra = append(*extra, content)
		return nil
	}
	if v, ok := parsed["Randomize Resources"]; ok && p.state == stateInsideMap {
		p.numericRandomize = isWbNumber(v)
	}
	if v, ok := parsed["RandomStartLocation"]; ok && p.state == stateInsidePlayer && isWbNumber(v) {
		p.numericStarts = true
	}
	before := len(*extra)
	if err := p.target().Unpack(parsed); err != nil {
		return &MapFileError{Code: "badValue", Line: p.line, Text: shorten(content), Section: sections[p.state][0], Detail: err.Error()}
	}
	keepUnknown(extra, before, len(parsed), bare, content)
	return nil
}

// keepUnknown restores the form of the unknown keys of a line the struct put into Extra: a line with no known key
// is kept as it is, and keys without a value stay without it (the struct stores them as key=1)
func keepUnknown(extra *[]string, before, keys int, bare []string, content string) {
	added := (*extra)[before:]
	if len(added) == 0 {
		return
	}
	if len(added) == keys {
		*extra = append((*extra)[:before], content)
		return
	}
	for i, line := range added {
		for _, k := range bare {
			if line == k+"=1" {
				added[i] = k
			}
		}
	}
}

// leave finishes the current section
func (p *wbParser) leave() {
	switch p.state {
	case stateInsideGame:
		p.wb.Game, p.last = p.game, "game"
	case stateInsideTeam:
		p.wb.Teams, p.last = append(p.wb.Teams, p.team), "team"
	case stateInsidePlayer:
		p.wb.Players, p.last = append(p.wb.Players, p.player), "player"
	case stateInsideMap:
		p.wb.Map, p.last = p.mapProps, "map"
	case stateInsidePlot:
		p.wb.Plots, p.last = append(p.wb.Plots, p.plot), "plot"
	case stateInsideSign:
		p.wb.Signs, p.last = append(p.wb.Signs, p.sign), "sign"
	case stateInsideCity:
		p.plot.Cities = append(p.plot.Cities, p.city)
		p.state = stateInsidePlot
		return
	case stateInsideUnit:
		p.plot.Units = append(p.plot.Units, p.unit)
		p.state = stateInsidePlot
		return
	}
	p.state = stateGlobal
}

// target is the struct of the current section
func (p *wbParser) target() WbStructPackable {
	switch p.state {
	case stateInsideGame:
		return p.game
	case stateInsideTeam:
		return p.team
	case stateInsidePlayer:
		return p.player
	case stateInsideMap:
		return p.mapProps
	case stateInsidePlot:
		return p.plot
	case stateInsideCity:
		return p.city
	case stateInsideUnit:
		return p.unit
	default:
		return p.sign
	}
}

// extra returns the lines unknown to the editor of the current section
func (p *wbParser) extra() *[]string {
	switch p.state {
	case stateInsideGame:
		return &p.game.Extra
	case stateInsideTeam:
		return &p.team.Extra
	case stateInsidePlayer:
		return &p.player.Extra
	case stateInsideMap:
		return &p.mapProps.Extra
	case stateInsidePlot:
		return &p.plot.Extra
	case stateInsideCity:
		return &p.city.Extra
	case stateInsideUnit:
		return &p.unit.Extra
	default:
		return &p.sign.Extra
	}
}

// startRaw starts reading an unknown section up to its End marker; target is where its lines go,
// nil for a top-level block
func (p *wbParser) startRaw(content string, target *[]string) {
	p.raw = []string{content}
	p.rawBegin, p.rawEnd = content, "End"+strings.TrimPrefix(content, "Begin")
	p.rawDepth, p.rawStart, p.rawTarget = 1, p.line, target
}

func (p *wbParser) closeRaw() {
	if p.rawTarget != nil {
		*p.rawTarget = append(*p.rawTarget, p.raw...)
	} else {
		p.wb.raw = append(p.wb.raw, rawBlock{after: p.last, lines: p.raw})
	}
	p.raw, p.rawTarget = nil, nil
}

func (p *wbParser) finish() (*WbMap, error) {
	wbMap := p.wb
	// Caveman2Cosmos writes Randomize Resources as a number and no RandomStartLocation. Both as numbers come
	// from Civ4 Studio 0.1.0, which wrote them so for every map: such a file gets the game's words back.
	wbMap.numericRandomize = p.numericRandomize && !p.numericStarts
	realPlayers, emptyPlayers := 0, 0
	for _, player := range wbMap.Players {
		if player.CivType == NonePlayer && player.LeaderType == NonePlayer {
			emptyPlayers++
		} else {
			realPlayers++
		}
	}

	ConsoleWrite("Loaded %d teams", len(wbMap.Teams))
	ConsoleWrite("Loaded %d players (+ %d player placeholders)", realPlayers, emptyPlayers)
	ConsoleWrite("Loaded %d plots", len(wbMap.Plots))
	ConsoleWrite("Loaded %d signs", len(wbMap.Signs))
	if len(wbMap.raw) > 0 {
		ConsoleWrite("Kept %d unknown parts of the file as they are", len(wbMap.raw))
	}

	switch {
	case !p.started:
		return nil, &MapFileError{Code: "empty"}
	case p.raw != nil:
		return nil, &MapFileError{Code: "unclosed", Line: p.rawStart, Section: p.rawBegin}
	case p.state != stateGlobal:
		return nil, &MapFileError{Code: "unclosed", Line: p.starts[p.state], Section: sections[p.state][0]}
	case wbMap.Game == nil:
		return nil, &MapFileError{Code: "noGame"}
	}
	return wbMap, nil
}

// parseLine splits a line into keys and values. Keys without a value (e.g. isNOfRiver) get "1" and are
// listed in bare. ok is false for a line that is not made of key=value pairs.
func parseLine(line string) (kv map[string]string, bare []string, ok bool) {
	kv = make(map[string]string)

	for _, key := range textKeys {
		if strings.HasPrefix(line, key+"=") {
			kv[key] = line[len(key)+1:]
			return kv, nil, true
		}
	}

	for _, contentPart := range strings.Split(line, ",") {
		contentPart = strings.Trim(contentPart, " ")
		if contentPart == "" {
			continue
		}

		key, value, err := parseKeyValue(contentPart)
		if err != nil {
			return nil, nil, false
		}
		if !strings.Contains(contentPart, "=") {
			bare = append(bare, key)
		}
		kv[key] = value
	}

	return kv, bare, true
}

func parseKeyValue(line string) (string, string, error) {
	if !strings.Contains(line, "=") && !strings.Contains(line, " ") {
		// Handling boolean keys without values (usually means TRUE)
		return line, "1", nil
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.New("unknown line format")
	}

	return parts[0], parts[1], nil
}

// parseWbBool reads a bool the game writes as a word ("false", "true") or a number ("0", "1")
func parseWbBool(v string) bool {
	return v == "1" || strings.EqualFold(v, "true")
}

func isWbNumber(v string) bool {
	_, err := strconv.Atoi(v)
	return err == nil
}
