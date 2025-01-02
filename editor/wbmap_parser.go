package editor

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
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
)

func ParseWbMap(reader io.Reader) (*WbMap, error) {
	fileScanner := bufio.NewScanner(reader)
	fileScanner.Split(bufio.ScanLines)

	var parsed map[string]string
	var err error
	var game *Game
	var mapProps *MapProps
	var lastTeam *Team
	var lastPlayer *Player
	var lastPlot *Plot
	var lastCity *City
	var lastUnit *Unit
	wbMap := &WbMap{Version: defaultVersion}

	line := 0
	parserState := stateGlobal

	ConsoleWrite("Parsing map contents...")

	for fileScanner.Scan() {
		line++
		content := fileScanner.Text()
		content = strings.Trim(content, " \t")
		if content == "" || strings.HasPrefix(content, "#") {
			continue
		}

		switch parserState {
		// Parser in global state (outside BeginX and other sections)
		case stateGlobal:
			switch {
			case strings.HasPrefix(content, versionPrefix):
				verStr := content[len(versionPrefix):]
				wbMap.Version, err = strconv.Atoi(verStr)
				if err != nil {
					return nil, createParserError("bad version number '%s'", line, verStr)
				}

			case content == BeginGame:
				parserState = stateInsideGame
				game = &Game{}

			case content == BeginTeam:
				parserState = stateInsideTeam
				lastTeam = &Team{}

			case content == BeginPlayer:
				parserState = stateInsidePlayer
				lastPlayer = &Player{}

			case content == BeginMap:
				parserState = stateInsideMap
				mapProps = &MapProps{}

			case content == BeginPlot:
				parserState = stateInsidePlot
				lastPlot = &Plot{}

			default:
				return nil, createParserError("cannot parse line in global context: '%s'", line, content)
			}

		// Parser between BeginMap and EndMap
		case stateInsideMap:
			if content == EndMap {
				wbMap.Map = mapProps
				parserState = stateGlobal
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = mapProps.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		case stateInsidePlot:
			if content == EndPlot {
				wbMap.Plots = append(wbMap.Plots, lastPlot)
				parserState = stateGlobal
				continue
			} else if content == BeginCity {
				parserState = stateInsideCity
				lastCity = &City{}
				continue
			} else if content == BeginUnit {
				parserState = stateInsideUnit
				lastUnit = &Unit{}
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = lastPlot.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser between BeginCity and EndCity
		case stateInsideCity:
			if content == EndCity {
				if lastPlot != nil {
					lastPlot.Cities = append(lastPlot.Cities, lastCity)
				}
				parserState = stateInsidePlot
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = lastCity.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser between BeginUnit and EndUnit
		case stateInsideUnit:
			if content == EndUnit {
				if lastPlot != nil {
					lastPlot.Units = append(lastPlot.Units, lastUnit)
				}
				parserState = stateInsidePlot
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = lastUnit.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser between BeginPlayer and EndPlayer
		case stateInsidePlayer:
			if content == EndPlayer {
				wbMap.Players = append(wbMap.Players, lastPlayer)
				parserState = stateGlobal
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = lastPlayer.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser between BeginTeam and EndTeam
		case stateInsideTeam:
			if content == EndTeam {
				wbMap.Teams = append(wbMap.Teams, lastTeam)
				parserState = stateGlobal
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = lastTeam.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser between BeginGame end EndGame
		case stateInsideGame:
			if content == EndGame {
				wbMap.Game = game
				parserState = stateGlobal
				continue
			}

			parsed, err = parseLine(content, line)
			if err != nil {
				return nil, err
			}

			err = game.Unpack(parsed)
			if err != nil {
				return nil, createParserError(err.Error(), line, content)
			}

		// Parser in unknown state (this should never happen)
		default:
			return nil, createParserError("internal error, unknown state %d", line, parserState)
		}
	}

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

	if wbMap.Game == nil {
		return nil, errors.New("no game info specified")
	}

	return wbMap, nil
}

func createParserError(err string, line int, p ...any) error {
	err = fmt.Sprintf(err, p...)

	return errors.New(
		fmt.Sprintf("parse error: %s (at line %d)", err, line),
	)
}

func parseLine(line string, lineNum int) (map[string]string, error) {
	kv := make(map[string]string)

	for _, contentPart := range strings.Split(line, ",") {
		contentPart = strings.Trim(contentPart, " ")
		if contentPart == "" {
			continue
		}

		key, value, err := parseKeyValue(contentPart)
		if err != nil {
			return nil, createParserError(err.Error(), lineNum, contentPart)
		}

		kv[key] = value
	}

	return kv, nil
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
