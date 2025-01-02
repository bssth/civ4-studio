package editor

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
)

var (
	_ WbStructPackable   = &WbMap{}
	_ WbStructPackable   = &Game{}
	_ WbStructPackable   = &Team{}
	_ WbStructPackable   = &Player{}
	_ WbStructPackable   = &Plot{}
	_ WbStructPackable   = &MapProps{}
	_ WbStructSubSection = &City{}
	_ WbStructSubSection = &Unit{}
)

// WbStructPackable is an interface that should be implemented by all structs that are used to pack/unpack data from a map
type WbStructPackable interface {
	// Unpack is a method that should be implemented by all structs that are used to unpack data from a map[string]string
	// Firstly it is parsed by ParseWbMap function to key-value map, and then it is unpacked by the struct
	// To avoid using reflection, the Unpack method should be implemented manually, parameter by parameter
	Unpack(map[string]string) error
	// ToWbFormat is a method that should be implemented by all structs that are used to pack data back to WorldBuilder format
	ToWbFormat() []byte
}

// WbStructSubSection is an interface that should be implemented by all structs that are used to add subsections to the map
// For example, City and Unit structs are subsections of the Plot struct
type WbStructSubSection interface {
	AddAsSubsection(generator *SimpleGenerator)
}

// WbMap is a struct that represents a map parsed from WorldBuilder format.
// It's a root struct that contains all other map data
type WbMap struct {
	Version int
	Game    *Game
	Teams   []*Team
	Map     *MapProps
	Players []*Player
	Plots   []*Plot
}

// Unpack from WbMap is not used, parser unpacks it manually.
// Anyway it is implemented to satisfy WbStructPackable interface and is usable to get the version of the map
func (m *WbMap) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "Version":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.Version = i
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (m *WbMap) ToWbFormat() []byte {
	buf := bytes.NewBuffer(nil)
	buf.Write([]byte(fmt.Sprintf("Version=%d\n", m.Version)))
	buf.Write(m.Game.ToWbFormat())
	for _, team := range m.Teams {
		buf.Write(team.ToWbFormat())
	}
	for _, player := range m.Players {
		buf.Write(player.ToWbFormat())
	}
	buf.Write(m.Map.ToWbFormat())
	for _, plot := range m.Plots {
		buf.Write(plot.ToWbFormat())
	}
	return buf.Bytes()
}

/*
	Note: descriptions of the following structs are taken from http://modiki.civfanatics.com
*/

type Game struct {
	// The Era value is which era the game/scenario will start in. This value can be
	// any of the Civilization IV eras that are defined in the file "CIV4EraInfos.xml" (found in your Civilization 4 directory\Assets\XML\GameInfo)
	Era string
	// The Speed value is the speed of the game. This is where you can set
	// if the game is normal speed, epic speed, etc. as defined in the file "CIV4GameSpeedInfo.xml" (found in your Civilization 4 directory\Assets\XML\GameInfo)
	Speed string
	// The Calendar value is the calendar used in the scenario. This is what determines the date as displayed on the main screen and the time jump between turns.
	// This is defined in the file "CIV4BasicInfos.xml" (found in your Civilization 4 directory\Assets\XML\BasicInfos)
	Calendar string
	// Victory is all the different victory types in the scenario. These are defined in the file "CIV4VictoryInfo.xml"
	// (found in your Civilization 4 directory\Assets\XML\GameInfo), and you may specify any or all of them.
	// Each victory must be specified on its own line. By defining a victory
	// type it means the player/AI has the possibility of winning the game that way.
	// The default is that no victory types are set (and no defined victory types means that all are possible).
	// But if you define any, even just one, then all others are locked out
	Victory []string
	// GameTurn is the game turn that the scenario begins on.
	// Every scenario starts turn 0 (the first turn is defined as zero [0]) but you can start the scenario on a different turn number by defining this variable.
	// EG: WW2 started September 1939. You define the calendar as CALENDAR_MONTHS.
	// The first turn is always the first calendar segment (in this case January). To specify September, you would set GameTurn=8.
	GameTurn uint
	// MaxCityElimination is the number of cities a multi-player player can lose before losing the game.
	// EG: "MaxCityElimination=3" means that each player will lose the game if they lose 3 cities.
	MaxCityElimination uint
	// NumAdvancedStartPoints is the number of points you get to pre-build your civ in advanced start mode
	// If you set it to 0, the game defaults to the normal Settler+Warrior/Scout, but higher than that, and it defaults to advanced start.
	// 600 is the standard value, but if you're using the advanced start feature, experiment with it a bit to find a good balance.
	// Source: https://forums.civfanatics.com/threads/numadvancedstartpoints.325926/
	NumAdvancedStartPoints uint
	// The TargetScore value determines the score a player must achieve to win the game.
	// For example, the Desert War scenario uses TargetScore=6, as there are 6 objective cities in the game. If one team holds all 6 cities,
	// then they win the game. By itself you can define the actual score a player must achieve (in the score list on the right of the interface),
	// but coupled with python, this can be an extremely powerful scoring utility. You must have Victory=VICTORY_SCORE (or all victory conditions available)
	// for this method to work.
	TargetScore uint
	// The StartYear value determines the physical date that the game begins in.
	// EG: WW2 starts in 1939, so you would set "StartYear=1939." To specify a BC date, use a negative number. The default StartYear value is -4000 (4000 BC).
	StartYear int
	// The Description is what the title suggests: it's the text displayed in the scenario menu when the scenario is selected.
	// This text is usually a short description or summary of the scenario that displays under the map window.
	Description string
	// The ModPath is the path to the folder containing your modified files. Only set this if you have also modified XML or python files.
	// Otherwise, save your WBS file to PublicMaps (found in your Civilization 4 directory\PublicMaps) and leave this line blank.
	// If set, it will force the scenario to use the settings in the mod folder defined rather than the default settings of Civilization 4.
	ModPath string
	// Tutorial: This is the setting to turn on the tutorial
	// By default this setting is 0, which means do not turn the tutorial on. Setting this value to 1 will turn on the Civilization 4 tutorial.
	Tutorial bool
	// The Option value is the selected game options (not player options) in the scenario. These options are defined in
	// the file "CIV4GameOptionInfos.xml" (found in your Civilization 4 directory\Assets\XML\GameInfo) and you can have any number of these options set in the scenario.
	Option []string
	// The MPOption (short for "Multi-Player Option") value is the selected multi-player options in the scenario. These options are defined in the
	// "CIV4MPOptionInfos.xml" file (found in your Civilization IV directory\Assets\XML\GameInfo),
	// while the Option value can have any number in your scenario. The default is that no options are specified.
	MPOption []string
	// ForceControl is the specified options that cannot be changed in the scenario. By setting these options, they appear grayed out in the scenario setup menu so the player cannot change them. The default is that no forced options specified.
	// These values are defined in the file "CIV4ForceControlInfos.xml" (found in your Civilization 4 directory\Assets\XML\GameInfo)
	ForceControl []string
	// The MaxTurns value is the maximum number of turns in the scenario. This must be set higher than GameTurn (it obviously can't start after the end).
	// EG: You have a scenario that you want to run for 300 years, and your calendar is set to CALENDAR_YEARS.
	// Setting MaxTurns=300 will end the game with score victory after turn 299 (remember that 0 is the first turn).
	MaxTurns uint
}

func (g *Game) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "Era":
			g.Era = v
		case "Speed":
			g.Speed = v
		case "Calendar":
			g.Calendar = v
		case "Victory":
			g.Victory = append(g.Victory, v)
		case "GameTurn":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.GameTurn = uint(i)
		case "MaxCityElimination":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.MaxCityElimination = uint(i)
		case "NumAdvancedStartPoints":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.NumAdvancedStartPoints = uint(i)
		case "TargetScore":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.TargetScore = uint(i)
		case "StartYear":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.StartYear = i
		case "Description":
			g.Description = v
		case "ModPath":
			g.ModPath = v
		case "Tutorial":
			g.Tutorial = v == "1"
		case "Option":
			g.Option = append(g.Option, v)
		case "MPOption":
			g.MPOption = append(g.MPOption, v)
		case "ForceControl":
			g.ForceControl = append(g.ForceControl, v)
		case "MaxTurns":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			g.MaxTurns = uint(i)
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (g *Game) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	generator.StartSection(BeginGame, EndGame)
	generator.AddKeyValue("Era", g.Era)
	generator.AddKeyValue("Speed", g.Speed)
	generator.AddKeyValue("Calendar", g.Calendar)
	generator.AddKeyValueArray("Victory", g.Victory)
	generator.AddKeyValueUint("GameTurn", uint64(g.GameTurn))
	generator.AddKeyValueUint("MaxCityElimination", uint64(g.MaxCityElimination))
	generator.AddKeyValueUint("NumAdvancedStartPoints", uint64(g.NumAdvancedStartPoints))
	generator.AddKeyValueUint("TargetScore", uint64(g.TargetScore))
	generator.AddKeyValueInt("StartYear", g.StartYear)
	generator.AddKeyValueString("Description", g.Description)
	generator.AddKeyValueString("ModPath", g.ModPath)
	generator.AddKeyValueBool("Tutorial", g.Tutorial)
	generator.AddKeyValueArray("Option", g.Option)
	generator.AddKeyValueArray("MPOption", g.MPOption)
	generator.AddKeyValueArray("ForceControl", g.ForceControl)
	generator.AddKeyValueUint("MaxTurns", uint64(g.MaxTurns))
	generator.EndSection()
	return generator.Bytes()
}

type Team struct {
	// The TeamID value is the unique identifier for the team. Usually the numbers are issued in sequence starting from 0.
	TeamID uint
	// The Tech value is the list of technologies that the team begins with.
	// By defining a list of technologies here, you control how much prior knowledge each team will begin with.
	Tech []string
	// ContactWithTeam defines the amount of diplomatic contacts the team has. Each diplomatic contact is defined separately.
	ContactWithTeam []uint
	// AtWar is the list of teams that this team begins at war with.
	AtWar []uint
	// PermanentWarPeace is the list of teams that the status of war/peace cannot be changed for.
	// EG: Team 0 has AtWar=1 and PermanentWarPeace=1 and PermanentWarPeace=2.
	// This means team 0 cannot sue for peace with team 1 and cannot declare war on team 2.
	// This would come in handy for a scenario such as a WW2 scenario, i.e. so Germany is always at war with the US, UK, and USSR.
	PermanentWarPeace []uint
	// OpenBordersWithTeam is the list of teams that an open border agreement exists with at the start of the game/scenario. This can be cancelled later in the game.
	OpenBordersWithTeam []uint
	// DefensivePactWithTeam is the list of teams that a defensive pact exists with at the start of the game/scenario.
	// This can be cancelled later unless the "PermanentWarPeace" value is defined.
	DefensivePactWithTeam []uint
	// ProjectType defines the projects that exist in the team. This way you can define if a team project exists in a team at the start of the game/scenario.
	// These values are defined in the file "CIV4ProjectInfo.xml" (found in your Civilization 4 directory\Assets\XML\GameInfo)
	ProjectType []string
	// RevealMap defines the state of the team knowing the whole map at the start of the game.
	// Valid options are 0 (don't know map) and 1 (knows map). If left out, then the default value is 0.
	RevealMap bool
}

func (t *Team) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "TeamID":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.TeamID = uint(i)
		case "Tech":
			t.Tech = append(t.Tech, v)
		case "ContactWithTeam":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.ContactWithTeam = append(t.ContactWithTeam, uint(i))
		case "AtWar":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.AtWar = append(t.AtWar, uint(i))
		case "PermanentWarPeace":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.PermanentWarPeace = append(t.PermanentWarPeace, uint(i))
		case "OpenBordersWithTeam":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.OpenBordersWithTeam = append(t.OpenBordersWithTeam, uint(i))
		case "DefensivePactWithTeam":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			t.DefensivePactWithTeam = append(t.DefensivePactWithTeam, uint(i))
		case "ProjectType":
			t.ProjectType = append(t.ProjectType, v)
		case "RevealMap":
			t.RevealMap = v == "1"
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (t *Team) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	generator.StartSection(BeginTeam, EndTeam)
	generator.AddKeyValueUint("TeamID", uint64(t.TeamID))
	generator.AddKeyValueArray("Tech", t.Tech)
	generator.AddKeyValueUintArray("ContactWithTeam", t.ContactWithTeam)
	generator.AddKeyValueUintArray("AtWar", t.AtWar)
	generator.AddKeyValueUintArray("PermanentWarPeace", t.PermanentWarPeace)
	generator.AddKeyValueUintArray("OpenBordersWithTeam", t.OpenBordersWithTeam)
	generator.AddKeyValueUintArray("DefensivePactWithTeam", t.DefensivePactWithTeam)
	generator.AddKeyValueArray("ProjectType", t.ProjectType)
	generator.AddKeyValueBool("RevealMap", t.RevealMap)
	generator.EndSection()
	return generator.Bytes()
}

type Player struct {
	// The CivDesc (short for "Civilization Description") is the descriptive name to give the civilization. This is the name that civ takes on in the game.
	CivDesc string
	// CivShortDesc (short for "Civilization Short Description") is the short descriptive name to give the civilization.
	// This is the short name that civilization is known as.
	CivShortDesc string
	// The LeaderName value defines the name of the leader of that civilization. This name is what the leader (king, queen, emperor, etc.) is called.
	LeaderName string
	// CivAdjective defines the descriptive name of the civ. This can be determined by
	// completing this sentence: "I am _____ who leads the _____ which is populated by the _____ people."
	// The last blank would be your value for CivAdjective.
	CivAdjective string
	// The FlagDecal value is the DDS file of the civilization's flag. These are defined in the folder "(your Civilization 4 directory)\Art\Interface\TeamColor\"
	FlagDecal string
	// The WhiteFlag value is the setting which gives the civilization's flag (what the units hold) a white background
	// or the background of the color of the civilization. Valid settings are 1 (use white) or 0 (use civilization's default colour).
	WhiteFlag bool
	// LeaderType defines the leader settings to use for this civ.
	// These values are defined in the file "CIV4LeaderHeadInfos.xml"
	LeaderType string
	// CivType defines the civilization to use for this player.
	// These values are defined in the file "CIV4CivilizationInfos.xml"
	CivType string
	// Team defines the team number that this civilization is part of.
	// The team settings are defined in the "BeginTeam" section, found above this section in the WBS file.
	// More than one civilization can be part of a team, and every civilization must be part of a team, even if it is by itself.
	Team uint
	// Handicap is the default handicap that the AI takes if no human takes this civilization.
	// These values are defined in the file "CIV4HandicapInfo.xml"
	Handicap string
	// Color defines the default color of the civilization. The color defines the civilization's border color,
	// the color of the name, etc. These values are defined in the file "CIV4PlayerColorInfos.xml"
	Color string
	// ArtStyle is the style of art that the civilization uses. This value defines building graphics and tile improvements.
	// These values are defined in the file "GlobalTypes.xml" (found in your Civilization 4 directory\Assets\XML)
	ArtStyle string
	// PlayableCiv is the setting to turn on whether this civilization can be played by a human or not.
	// Valid values are 0 (AI only) or 1 (playable by human).
	PlayableCiv bool
	// MinorNationStatus is the setting to determine if a civilization is a minor nation in relation to diplomacy.
	// Valid values are 0 (full power civilization, such as the Aztec) or 1 (minor nation civilization that won't do diplomacy with anyone).
	MinorNationStatus bool
	// The StartingGold value is the amount of gold that each civilization starts with in a scenario. The StartingGold
	// value can be set to any numerical value
	// (i.e. if you wanted each civilization to start with 100 gold, the line of the file would look like "StartingGold= 100").
	StartingGold int
	// RandomStartLocation is the setting to determine if the civilization starts in a random location on the map.
	// If you change the relevant statement to false,
	// then the fixed starting positions should be applied (if they are still present; you can also search for them in the file).
	// Source: https://forums.civfanatics.com/threads/fixed-starting-locations-with-random-civ-world-builder.367650/
	RandomStartLocation bool
	// StartingX determines the X-axis location of the starting plot of this civilization.
	// This value is only valid if there are no cities on the map for this civilization.
	StartingX int
	// StartingY determines the Y-axis location of the starting plot of this civilization.
	// This value, like the StartingX value, is only valid if there are no cities on the map for this civilization.
	StartingY int
	// StateReligion defines the State Religion that the civilization starts the game with.
	// These values are defined in the file "CIV4ReligionInfo.xml"
	StateReligion string
	// The StartingEra value defines the era that the civilization begins the game in regard to graphics.
	// These values are defined in the file "CIV4EraInfos.xml"
	StartingEra string
	// CityList is the name of cities that the civilization has available when founding new cities.
	CityList []string
	// E.G. CivicOption=XXXX where XXXX is the civic category. Civic options are defined in the file "CIV4CivicOptionInfos.xml"
	CivicOption []string
	// E.G. Civic=YYYY where YYYY is the actual civic. Civics are defined in the file "CIV4CivicInfos.xml"
	Civic []string
	// EG: AttituedPlayer=XXX where XXX is the player number affected.
	AttitudePlayer []uint
	// EG: AttitudeExtra=YYY where YYY is the amount to change diplomatic attitude towards the player defined in "AttitudePlayer."
	AttitudeExtra []int
}

func (p *Player) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	generator.StartSection(BeginPlayer, EndPlayer)
	generator.AddKeyValueString("CivDesc", p.CivDesc)
	generator.AddKeyValueString("CivShortDesc", p.CivShortDesc)
	generator.AddKeyValueString("LeaderName", p.LeaderName)
	generator.AddKeyValueString("CivAdjective", p.CivAdjective)
	generator.AddKeyValueString("FlagDecal", p.FlagDecal)
	generator.AddKeyValueBool("WhiteFlag", p.WhiteFlag)
	generator.AddKeyValueString("LeaderType", p.LeaderType)
	generator.AddKeyValueString("CivType", p.CivType)
	generator.AddKeyValueUint("Team", uint64(p.Team))
	generator.AddKeyValueString("Handicap", p.Handicap)
	generator.AddKeyValueString("Color", p.Color)
	generator.AddKeyValueString("ArtStyle", p.ArtStyle)
	generator.AddKeyValueBool("PlayableCiv", p.PlayableCiv)
	generator.AddKeyValueBool("MinorNationStatus", p.MinorNationStatus)
	generator.AddKeyValueInt("StartingGold", p.StartingGold)
	generator.AddKeyValueBool("RandomStartLocation", p.RandomStartLocation)
	generator.AddCommaSeparatedValues(
		fmt.Sprintf("StartingX=%d", p.StartingX),
		fmt.Sprintf("StartingY=%d", p.StartingY),
	)
	generator.AddKeyValueString("StateReligion", p.StateReligion)
	generator.AddKeyValueString("StartingEra", p.StartingEra)
	generator.AddKeyValueArray("CityList", p.CityList)
	generator.AddKeyValueArray("CivicOption", p.CivicOption)
	generator.AddKeyValueArray("Civic", p.Civic)
	generator.AddKeyValueUintArray("AttitudePlayer", p.AttitudePlayer)
	generator.AddKeyValueIntArray("AttitudeExtra", p.AttitudeExtra)
	generator.EndSection()
	return generator.Bytes()
}

func (p *Player) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "CivDesc":
			p.CivDesc = v
		case "CivShortDesc":
			p.CivShortDesc = v
		case "LeaderName":
			p.LeaderName = v
		case "CivAdjective":
			p.CivAdjective = v
		case "FlagDecal":
			p.FlagDecal = v
		case "WhiteFlag":
			p.WhiteFlag = v == "1"
		case "LeaderType":
			p.LeaderType = v
		case "CivType":
			p.CivType = v
		case "Team":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.Team = uint(i)
		case "Handicap":
			p.Handicap = v
		case "Color":
			p.Color = v
		case "ArtStyle":
			p.ArtStyle = v
		case "PlayableCiv":
			p.PlayableCiv = v == "1"
		case "MinorNationStatus":
			p.MinorNationStatus = v == "1"
		case "StartingGold":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.StartingGold = i
		case "RandomStartLocation":
			p.RandomStartLocation = v == "1"
		case "StartingX":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.StartingX = i
		case "StartingY":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.StartingY = i
		case "StateReligion":
			p.StateReligion = v
		case "StartingEra":
			p.StartingEra = v
		case "CityList":
			p.CityList = append(p.CityList, v)
		case "CivicOption":
			p.CivicOption = append(p.CivicOption, v)
		case "Civic":
			p.Civic = append(p.Civic, v)
		case "AttitudePlayer":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.AttitudePlayer = append(p.AttitudePlayer, uint(i))
		case "AttitudeExtra":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.AttitudeExtra = append(p.AttitudeExtra, i)
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

type Plot struct {
	// X: x=XXX,y=YYY where XXX = the column (vertical) that the plot resides in on the map and YYY = the row (horizontal)
	// that the plot resides in on the map. Columns begin at 0 (zero) on the left edge of the map and increase by 1 to the right.
	// Rows begin at 0 (zero) on the bottom edge of the map and increase by 1 upward.
	// Bottom left plot is 0,0 and top right plot is (mapwidth - 1),(mapheight - 1).
	X uint // @todo export lowercase
	// Y: x=XXX,y=YYY where XXX = the column (vertical) that the plot resides in on the map and YYY = the row (horizontal)
	// that the plot resides in on the map. Columns begin at 0 (zero) on the left edge of the map and increase by 1 to the right.
	// Rows begin at 0 (zero) on the bottom edge of the map and increase by 1 upward.
	// Bottom left plot is 0,0 and top right plot is (mapwidth - 1),(mapheight - 1).
	Y uint // @todo export lowercase
	// The setting to point to a sign or landmark on the map. The value is the text to be displayed by the landmark. EG: Landmark=This is a landmark!
	Landmark string
	// The ScriptData is a pointer to a plot script. In the WBS it is possible to assign a script to a city.
	// This reference does not go into these scripts (which are used by python).
	ScriptData string
	// IsNOfRiver and IsWOfRiver are the two settings to place a river through this plot.
	// They do not require a value being flags to the graphics engine. isNOfRiver (is north of river)
	// defines the river as being on the bottom edge of the plot and isWOfRiver (is west of river) defines the river as being on the right edge of the plot.
	// These settings MUST be used in conjunction with RiverNSDirection or RiverWEDirection.
	IsNOfRiver bool
	// IsNOfRiver and IsWOfRiver are the two settings to place a river through this plot.
	// They do not require a value being flags to the graphics engine. isNOfRiver (is north of river)
	// defines the river as being on the bottom edge of the plot and isWOfRiver (is west of river) defines the river as being on the right edge of the plot.
	// These settings MUST be used in conjunction with RiverNSDirection or RiverWEDirection.
	IsWOfRiver bool
	// RiverNSDirection= and RiverWEDirection= the direction that the water flows along the river.
	// Valid values are 0=north, 1=east, 2=south, 3=west. These settings MUST be used in conjunction with isNOfRiver or isWOfRiver
	RiverNSDirection int
	// RiverNSDirection= and RiverWEDirection= the direction that the water flows along the river.
	// Valid values are 0=north, 1=east, 2=south, 3=west. These settings MUST be used in conjunction with isNOfRiver or isWOfRiver
	RiverWEDirection int
	// StartingPlot is the flag used by the Civ4 engine to define a civilizations starting location. This will assign a random civ from the scenario at this location.
	// If you wish to specify that a civ gets the same starting location each game then define it through that civ's BeginPlayer section.
	StartingPlot bool
	// The setting to place a bonus at this plot.
	// Resources are classed as bonuses (but bonuses are not just resources).
	// These values are defined in CIV4BonusInfos.xml. EG: BonusType=BONUS_WHEAT
	BonusType string
	// The setting to place an improvement at this plot. These are defined in CIV4ImprovementInfos.xml. EG: ImprovementType=IMPROVEMENT_MINE
	ImprovementType string
	// FeatureType=XXX, FeatureVariety=YYY where XXX is the terrain feature to place on this plot and YYY is which variety of the valid terrain feature to place.
	// Forests is an example of a terrain feature, while the FeatureVariety will determine which version of the forest is placed (pines, hardwood, etc.)
	// These are defined in CIV4FeatureInfos.xml. EG: FeatureType=FEATURE_FOREST, FeatureVariety=1
	FeatureType []string
	// FeatureType=XXX, FeatureVariety=YYY where XXX is the terrain feature to place on this plot and YYY is which variety of the valid terrain feature to place.
	// Forests is an example of a terrain feature, while the FeatureVariety will determine which version of the forest is placed (pines, hardwood, etc.)
	// These are defined in CIV4FeatureInfos.xml. EG: FeatureType=FEATURE_FOREST, FeatureVariety=1
	FeatureVariety []string
	// The setting to place a particular transportation type in the plot.
	// Routes are also important as they define trade routes too. These settings are defined in CIV4RouteInfos.xml. EG: RouteType=ROUTE_RAILROAD
	RouteType string
	// The base terrain type of the plot. These values are defined in CIV4TerrainInfos.xml.
	// EVERY plot will have a TerrainType setting. EG: TerrainType=TERRAIN_GRASS
	TerrainType string
	// The setting which determines the height of the plot. This basically determines if the plot is below sea level, a hill, a mountain or flat terrain.
	// Valid values are: 0=Peak(mountain), 1=Hills, 2=Flat and 3=Sea(land below sea level).
	// Source: Dale's "In depth look at the WBS file" CivFanatics archive: https://forums.civfanatics.com/threads/in-depth-look-at-the-wbs-file.135669/
	PlotType uint
	// Units is the list of units that are placed on this plot. These units are defined in the "BeginUnit" section.
	Units []*Unit
	// Cities is the list of cities that are placed on this plot. These cities are defined in the "BeginCity" section.
	Cities []*City
	// The list of teams that this plot is revealed to at the start of the game.
	// The teams in this list will be able to view the plot, but fog of war may still be over the plot.
	// The list is simply a list of the team numbers seperated by a comma. The list MUST end with a comma. EG: TeamReveal=TeamReveal=0,1,2,3,
	TeamReveal []uint
}

func (p *Plot) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "x":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.X = uint(i)
		case "y":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.Y = uint(i)
		case "Landmark":
			p.Landmark = v
		case "ScriptData":
			p.ScriptData = v
		case "isNOfRiver":
			p.IsNOfRiver = v == "1"
		case "isWOfRiver":
			p.IsWOfRiver = v == "1"
		case "RiverNSDirection":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.RiverNSDirection = i
		case "RiverWEDirection":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.RiverWEDirection = i
		case "StartingPlot":
			p.StartingPlot = v == "1"
		case "BonusType":
			p.BonusType = v
		case "ImprovementType":
			p.ImprovementType = v
		case "FeatureType":
			p.FeatureType = append(p.FeatureType, v)
		case "FeatureVariety":
			p.FeatureVariety = append(p.FeatureVariety, v)
		case "RouteType":
			p.RouteType = v
		case "TerrainType":
			p.TerrainType = v
		case "PlotType":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.PlotType = uint(i)
		case "TeamReveal":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			p.TeamReveal = append(p.TeamReveal, uint(i))
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (p *Plot) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	generator.StartSection(BeginPlot, EndPlot)
	generator.AddCommaSeparatedValues(fmt.Sprintf("x=%d", p.X), fmt.Sprintf("y=%d", p.Y))
	generator.AddKeyValueString("Landmark", p.Landmark)
	generator.AddKeyValueString("ScriptData", p.ScriptData)
	if p.IsNOfRiver {
		generator.AddLine("isNOfRiver")
		generator.AddKeyValueInt("RiverWEDirection", p.RiverWEDirection)
	}
	if p.IsWOfRiver {
		generator.AddLine("isWOfRiver")
		generator.AddKeyValueInt("RiverNSDirection", p.RiverNSDirection)
	}
	generator.AddKeyValueBool("StartingPlot", p.StartingPlot)
	generator.AddKeyValueString("BonusType", p.BonusType)
	generator.AddKeyValueString("ImprovementType", p.ImprovementType)

	if len(p.FeatureType) > 0 {
		for idx, s := range p.FeatureType {
			generator.AddCommaSeparatedValues(fmt.Sprintf("FeatureType=%s", s), fmt.Sprintf("FeatureVariety=%s", p.FeatureVariety[idx]))
		}
	}

	generator.AddKeyValueString("RouteType", p.RouteType)
	generator.AddKeyValueString("TerrainType", p.TerrainType)
	generator.AddKeyValueUint("PlotType", uint64(p.PlotType))
	for _, unit := range p.Units {
		unit.AddAsSubsection(generator)
	}
	for _, city := range p.Cities {
		city.AddAsSubsection(generator)
	}
	generator.AddKeyValueUintArray("TeamReveal", p.TeamReveal)
	generator.EndSection()
	return generator.Bytes()
}

type MapProps struct {
	// GridWidth: the grid width value determines the width of the map in number of plots/tiles.
	// NOTE: The grid width begins at zero so the first column of plots will be 0 NOT 1. However, you still define grid width in real terms.
	GridWidth uint64
	// GridHeight: The grid height value determines the height of the map in number of plots/tiles.
	// NOTE: The grid height begins at zero so the first row of plots will be 0 NOT 1. However, you still define grid height in real terms.
	GridHeight uint64
	// TopLatitude: The top latitude determines the northernmost point on the map. The maximum (and default) for this value is 90 (the North Pole)
	// this will show the northern ice cap. To remove the northern ice cap (normally for maps focusing on a specific area, or fantasy maps),
	// simply reduce this value (usually to 60). This value must be positive.
	TopLatitude int64
	// BottomLatitude:The bottom latitude determines the southernmost point on the map.
	// The minimum (and default) for this value is -90 (the South Pole); this will show the southern ice cap.
	// To remove the southern ice cap (normally for maps focusing on a specific area, or fantasy maps),
	// simply reduce this value (usually to -60). This value must be negative.
	BottomLatitude int64
	// WrapX: The wrapping setting on the x-axis (horizontal). By default, this is 1 (map wraps at the left-right edges).
	// If you set this value to 0 then the x-axis does not wrap.
	// Combined with the y wrap setting you can create flat maps, doughnut maps or maps that wrap left-right or top-bottom.
	WrapX int
	// WrapY: The wrapping setting on the y-axis (vertical). By default this is 0 (no wrapping at the top-bottom edges). This setting works the same as wrap X.
	WrapY int
	// WorldSize: The map size setting of the scenario. This is usually set when you setup the map in the WBS in-game. However you may want to change it. These values are defined in CIV4WorldInfo.xml.
	WorldSize string
	// Climate: The climate setting of the game. These are the same as setting in a new game setup from the main menu. These values are defined in CIV4ClimateInfo.xml.
	Climate string
	// SeaLevel: The sea level setting of the game. These are the same as setting in a new game setup from the main menu. These values are defined in CIV4SeaLevelInfo.xml.
	SeaLevel string
	// NumPlotsWritten: The total number of plots in the game. This value is derived by multiplying the values from grid width and grid height above. EG: grid width=50 and grid height=50 then num plots written=2500 (50 * 50).
	NumPlotsWritten uint64
	// NumSignsWritten: The total number of player-specified signs in the game.
	// @todo find more information about this
	NumSignsWritten uint64
	// RandomizeResources: The setting to randomize resources on the map.
	// @todo find more information about this
	RandomizeResources bool
}

func (m *MapProps) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "grid width":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.GridWidth = uint64(i)
		case "grid height":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.GridHeight = uint64(i)
		case "top latitude":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.TopLatitude = int64(i)
		case "bottom latitude":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.BottomLatitude = int64(i)
		case "wrap X":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.WrapX = i
		case "wrap Y":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.WrapY = i
		case "world size":
			m.WorldSize = v
		case "climate":
			m.Climate = v
		case "sealevel":
			m.SeaLevel = v
		case "num plots written":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.NumPlotsWritten = uint64(i)
		case "num signs written":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			m.NumSignsWritten = uint64(i)
		case "Randomize Resources":
			m.RandomizeResources = v == "1"
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (m *MapProps) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	generator.StartSection(BeginMap, EndMap)
	generator.AddKeyValueUint("grid width", m.GridWidth)
	generator.AddKeyValueUint("grid height", m.GridHeight)
	generator.AddKeyValueInt64("top latitude", m.TopLatitude)
	generator.AddKeyValueInt64("bottom latitude", m.BottomLatitude)
	generator.AddKeyValueInt("wrap X", m.WrapX)
	generator.AddKeyValueInt("wrap Y", m.WrapY)
	generator.AddKeyValueString("world size", m.WorldSize)
	generator.AddKeyValueString("climate", m.Climate)
	generator.AddKeyValueString("sealevel", m.SeaLevel)
	generator.AddKeyValueUint("num plots written", m.NumPlotsWritten)
	generator.AddKeyValueUint("num signs written", m.NumSignsWritten)
	generator.AddKeyValueBool("Randomize Resources", m.RandomizeResources)
	generator.EndSection()
	return generator.Bytes()
}

type City struct {
	// CityOwner: the city owner. Similar to unit owner it is the value between 0 and 17 of the player who owns this city.
	CityOwner uint
	// The name of the city. This can be any value. EG: CityName=My City
	CityName string
	// The starting population in the city. CityPopulation is how many population points the city starts with.
	CityPopulation uint
	// ProductionUnit: the unit that the city is building at game start.
	// Only one Production type is used (the first one in the city definition). These values are defined in CIV4UnitInfos.xml
	ProductionUnit string
	// ProductionBuilding: the building that the city is building at game start.
	// Only one Production type is used (the first one in the city definition). These values are defined in CIV4BuildingInfos.xml
	ProductionBuilding string
	// ProductionProject: the project that the city is building at game start.
	// Only one Production type is used (the first one in the city definition). These values are defined in CIV4ProjectInfo.xml
	ProductionProject string
	// ProductionProcess: the process (science/wealth/culture) that the city is building at game start.
	// Only one Production type is used (the first one in the city definition). These values are defined in CIV4ProcessInfo.xml
	ProductionProcess string
	// BuildingType: the buildings that the city already has at game start.
	// Any number of BuildingTypes can be defined on separate lines. These values are defined in CIV4BuildingInfos.xml
	BuildingType string
	// The religions that the city has at game start
	// Any number of religions can be defined on separate lines. These values are defined in CIV4ReligionInfos.xml
	ReligionType string
	// HolyCityReligionType: the Holy City of the defined religions. Any number of these can be defined on separate lines.
	// These values are defined in CIV4ReligionInfos.xml
	HolyCityReligionType string
	// ScriptData: any scripts assigned to the city. This analysis does not go into these scripts.
	ScriptData string
	// The starting culture that the city has. Key is the player number and value is the amount of culture.
	// EG: PlayerCulture[3]=100 means this city begins with 100 points of player 3's culture.
	// You can define a culture level for any number of players.
	PlayerCulture map[uint]uint64
}

var playerCultureRegex = regexp.MustCompile("`Player([0-9]+)Culture`")

func (c *City) Unpack(packed map[string]string) error {
	for k, v := range packed {
		if playerCultureRegex.MatchString(k) {
			m := playerCultureRegex.FindStringSubmatch(k)
			if len(m) != 2 {
				return fmt.Errorf("invalid player culture key: %s", k)
			}

			i, err := strconv.Atoi(m[1])
			if err != nil {
				return err
			}

			numValue, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return err
			}

			c.PlayerCulture[uint(i)] = uint64(numValue)
			continue
		}

		switch k {
		case "CityOwner":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.CityOwner = uint(i)
		case "CityName":
			c.CityName = v
		case "CityPopulation":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			c.CityPopulation = uint(i)
		case "ProductionUnit":
			c.ProductionUnit = v
		case "ProductionBuilding":
			c.ProductionBuilding = v
		case "ProductionProject":
			c.ProductionProject = v
		case "ProductionProcess":
			c.ProductionProcess = v
		case "BuildingType":
			c.BuildingType = v
		case "ReligionType":
			c.ReligionType = v
		case "HolyCityReligionType":
			c.HolyCityReligionType = v
		case "ScriptData":
			c.ScriptData = v
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (c *City) AddAsSubsection(generator *SimpleGenerator) {
	generator.StartSection(BeginCity, EndCity)
	generator.AddKeyValueUint("CityOwner", uint64(c.CityOwner))
	generator.AddKeyValueString("CityName", c.CityName)
	generator.AddKeyValueUint("CityPopulation", uint64(c.CityPopulation))
	generator.AddKeyValueString("ProductionUnit", c.ProductionUnit)
	generator.AddKeyValueString("ProductionBuilding", c.ProductionBuilding)
	generator.AddKeyValueString("ProductionProject", c.ProductionProject)
	generator.AddKeyValueString("ProductionProcess", c.ProductionProcess)
	generator.AddKeyValueString("BuildingType", c.BuildingType)
	generator.AddKeyValueString("ReligionType", c.ReligionType)
	generator.AddKeyValueString("HolyCityReligionType", c.HolyCityReligionType)
	generator.AddKeyValueString("ScriptData", c.ScriptData)
	for k, v := range c.PlayerCulture {
		generator.AddKeyValueUint(fmt.Sprintf("Player%vCulture", k), v)
	}
	generator.EndSection()
}

func (c *City) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	c.AddAsSubsection(generator)
	return generator.Bytes()
}

type Unit struct {
	// UnitType that is at the plot. These values are defined in CIV4UnitInfos.xml.
	UnitType string
	// UnitOwner (the player number who owns this unit).
	// The first player is player 0 with the last possible player being player 17 (equals 18 players).
	UnitOwner int
	// Yhe experience Level of the unit. Each Level means one more promotion is possible.
	// EG: Level=0 means no promotions, Level=2 means 2 promotions.
	Level int
	// The actual Experience of the unit. This reflects how many points it has gained towards the next promotion level.
	Experience int
	// The promotions this unit has. You assign as many PromotionType lines as Levels given to the unit above.
	// These values are defined in CIV4PromotionInfos.xml.
	PromotionType string
	// The usage of the unit for the AI. Assigning the correct UnitAIType for a unit
	// is important as it tells the AI what the unit is used for.
	// EG: Settler units should get UnitAIType=UNITAI_SETTLE
	UnitAIType string
	// Damage: @todo find information about this key
	Damage uint
	// FacingDirection: 2 for east, 3 for south-east, 4 for south and so on.
	// Source: https://forums.civfanatics.com/threads/world-builder-assigning-colonist-professions.321004/
	FacingDirection int
}

func (u *Unit) Unpack(packed map[string]string) error {
	for k, v := range packed {
		switch k {
		case "UnitType":
			u.UnitType = v
		case "UnitOwner":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			u.UnitOwner = i
		case "Level":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			u.Level = i
		case "Experience":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			u.Experience = i
		case "PromotionType":
			u.PromotionType = v
		case "UnitAIType":
			u.UnitAIType = v
		case "Damage":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			u.Damage = uint(i)
		case "FacingDirection":
			i, err := strconv.Atoi(v)
			if err != nil {
				return err
			}
			u.FacingDirection = i
		default:
			return fmt.Errorf("unknown key: %s", k)
		}
	}

	return nil
}

func (u *Unit) AddAsSubsection(generator *SimpleGenerator) {
	generator.StartSection(BeginUnit, EndUnit)
	generator.AddCommaSeparatedValues(fmt.Sprintf("UnitType=%s", u.UnitType), fmt.Sprintf("UnitOwner=%d", u.UnitOwner))
	generator.AddCommaSeparatedValues(fmt.Sprintf("Level=%d", u.Level), fmt.Sprintf("Experience=%d", u.Experience))
	generator.AddKeyValueString("PromotionType", u.PromotionType)
	generator.AddKeyValueString("UnitAIType", u.UnitAIType)
	generator.AddKeyValueUint("Damage", uint64(u.Damage))
	generator.AddKeyValueInt("FacingDirection", u.FacingDirection)
	generator.EndSection()
}

func (u *Unit) ToWbFormat() []byte {
	generator := &SimpleGenerator{}
	u.AddAsSubsection(generator)
	return generator.Bytes()
}
