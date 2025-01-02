package editor

type TypeInfo struct {
	Type        string
	Description string
}

var LangStrings = make(map[string]string)
var GlobalStringDefines = make(map[string]string)
var GlobalIntDefines = make(map[string]int)
var GlobalFloatDefines = make(map[string]float64)
var EraInfos = make(map[string]*TypeInfo)
var SpeedInfos = make(map[string]*TypeInfo)
var CalendarInfos = make(map[string]*TypeInfo)
var GameOptionInfos = make(map[string]*TypeInfo)
var GameMPInfos = make(map[string]*TypeInfo)
var ForceControlInfos = make(map[string]*TypeInfo)
var VictoryInfos = make(map[string]*TypeInfo)

// GetLangString returns language string by key. If it's not found, returns key itself
func GetLangString(key string) string {
	if key == "" {
		return ""
	}

	// Support keys with and without TXT_KEY_ prefix as automatic fallback
	for _, variant := range []string{key, "TXT_KEY_" + key} {
		value, ok := LangStrings[variant]
		if ok {
			return value
		}
	}

	return key
}

func GetEraNames() []string {
	var result []string
	for _, era := range EraInfos {
		result = append(result, era.Type)
	}
	return result
}

func GetSpeedNames() []string {
	var result []string
	for _, speed := range SpeedInfos {
		result = append(result, speed.Type)
	}
	return result
}

func GetCalendarNames() []string {
	var result []string
	for _, calendar := range CalendarInfos {
		result = append(result, calendar.Type)
	}
	return result
}
