package editor

import (
	"encoding/xml"
	"strconv"
)

const XmlDir = "Assets/XML"
const XmlExt = "xml"

// LoadAllXML loads all XML files recursively from the game directory.
// XML file type is automatically detected and assigned to the appropriate global variable.
// progressHandler is a manual callback function that is called after each file is parsed (for UI updates).
func LoadAllXML(progressHandler func(string)) error {
	var decoder *xml.Decoder
	var xmlType CivXmlType

	files, err := GetFilesFromGameDirsRecursive(XmlDir, XmlExt)
	if err != nil {
		ConsoleWrite(err.Error())
		return err
	}

	if progressHandler == nil {
		progressHandler = func(s string) {}
	}
	counter := make(map[CivXmlType]int32)

	for _, f := range files {
		progressHandler("Parsing XML: " + f)

		decoder, xmlType, err = ParseXMLFromFile(f)
		if err != nil {
			continue
		}

		switch xmlType {
		case "Schema":
			break

		case "Civ4Defines":
			definesStruct := &Civ4Defines{}
			err = decoder.Decode(definesStruct)
			if err != nil {
				break
			}

			counter[xmlType] += AssignGlobalDefines(definesStruct)

		case "Civ4GameText":
			textStruct := &Civ4GameText{}
			err = decoder.Decode(textStruct)
			if err != nil {
				break
			}

			for _, text := range textStruct.TEXT {
				LangStrings[text.Tag] = text.English
				// @todo multiple languages
				counter[xmlType]++
			}

		case "Civ4EraInfos":
			eraStruct := &Civ4EraInfos{}
			err = decoder.Decode(eraStruct)
			if err != nil {
				break
			}

			for _, era := range eraStruct.EraInfos.EraInfo {
				EraInfos[era.Type] = &TypeInfo{
					Type:        era.Type,
					Description: era.Description,
				}
				counter[xmlType]++
			}

		case "Civ4GameSpeedInfo":
			speedStruct := &Civ4GameSpeedInfo{}
			err = decoder.Decode(speedStruct)
			if err != nil {
				break
			}

			for _, speed := range speedStruct.GameSpeedInfos.GameSpeedInfo {
				SpeedInfos[speed.Type] = &TypeInfo{
					Type:        speed.Type,
					Description: speed.Description,
				}
				counter[xmlType]++
			}

		case "Civ4CalendarInfos":
			calendarStruct := &Civ4CalendarInfos{}
			err = decoder.Decode(calendarStruct)
			if err != nil {
				break
			}

			for _, calendar := range calendarStruct.CalendarInfos.CalendarInfo {
				CalendarInfos[calendar.Type] = &TypeInfo{
					Type:        calendar.Type,
					Description: calendar.Description,
				}
				counter[xmlType]++
			}

		case "Civ4GameOptionInfos":
			gameOptionsStruct := &Civ4GameOptionInfos{}
			err = decoder.Decode(gameOptionsStruct)
			if err != nil {
				break
			}

			for _, option := range gameOptionsStruct.GameOptionInfos.GameOptionInfo {
				GameOptionInfos[option.Type] = &TypeInfo{
					Type:        option.Type,
					Description: option.Description,
				}
				counter[xmlType]++
			}

		case "Civ4MPOptionInfos":
			gameMpOptionsStruct := &Civ4MPOptionInfos{}
			err = decoder.Decode(gameMpOptionsStruct)
			if err != nil {
				break
			}

			for _, option := range gameMpOptionsStruct.MPOptionInfos.MPOptionInfo {
				GameMPInfos[option.Type] = &TypeInfo{
					Type:        option.Type,
					Description: option.Description,
				}
				counter[xmlType]++
			}

		case "Civ4ForceControlInfos":
			forceControlStruct := &Civ4ForceControlInfos{}
			err = decoder.Decode(forceControlStruct)
			if err != nil {
				break
			}

			for _, option := range forceControlStruct.ForceControlInfos.ForceControlInfo {
				ForceControlInfos[option.Type] = &TypeInfo{
					Type:        option.Type,
					Description: option.Description,
				}
				counter[xmlType]++
			}

		case "Civ4VictoryInfo":
			victoryStruct := &Civ4VictoryInfo{}
			err = decoder.Decode(victoryStruct)
			if err != nil {
				break
			}

			for _, option := range victoryStruct.VictoryInfos.VictoryInfo {
				VictoryInfos[option.Type] = &TypeInfo{
					Type:        option.Type,
					Description: option.Description,
				}
				counter[xmlType]++
			}
		}

		if err != nil {
			ConsoleWrite("Cannot parse %s: %s (%s)", xmlType, err.Error(), f)
			continue
		}
	}

	for civXmlType, cnt := range counter {
		ConsoleWrite("Loaded %d %s", cnt, civXmlType)
	}
	return nil
}

// AssignGlobalDefines assigns all defines from the Civ4Defines struct to the global variables.
func AssignGlobalDefines(definesStruct *Civ4Defines) int32 {
	var counter int32 = 0
	for _, define := range definesStruct.Define {
		if define.IDefineIntVal != "" {
			GlobalIntDefines[define.DefineName], _ = strconv.Atoi(define.IDefineIntVal)
		} else if define.FDefineFloatVal != "" {
			GlobalFloatDefines[define.DefineName], _ = strconv.ParseFloat(define.FDefineFloatVal, 64)
		} else if define.DefineTextVal != "" {
			GlobalStringDefines[define.DefineName] = define.DefineTextVal
		} else {
			continue
		}

		counter++
	}

	return counter
}
