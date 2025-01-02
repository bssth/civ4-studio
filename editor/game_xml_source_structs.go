package editor

import "encoding/xml"

// This file contains the structs for representing game's XML files.
// All structs are auto generated from sample XML files.

type CivXmlType string

type Civ4GameText struct {
	XMLName xml.Name `xml:"Civ4GameText"`
	Text    string   `xml:",chardata"`
	TEXT    []struct {
		Text        string `xml:",chardata"`
		Tag         string `xml:"Tag"`
		English     string `xml:"English"`
		French      string `xml:"French"`
		German      string `xml:"German"`
		Italian     string `xml:"Italian"`
		Spanish     string `xml:"Spanish"`
		Polish      string `xml:"Polish"`
		Russian     string `xml:"Russian"`
		Czech       string `xml:"Czech"`
		Danish      string `xml:"Danish"`
		Greek       string `xml:"Greek"`
		Brazilian   string `xml:"Brazilian"`
		ChineseSimp string `xml:"ChineseSimp"`
		Korean      string `xml:"Korean"`
		Ukrainian   string `xml:"Ukrainian"`
		Arabic      string `xml:"Arabic"`
		Turkish     string `xml:"Turkish"`
		Bulgarian   string `xml:"Bulgarian"`
		Finnish     string `xml:"Finnish"`
		Dutch       string `xml:"Dutch"`
		Hungarian   string `xml:"Hungarian"`
		Japanese    string `xml:"Japanese"`
		Portuguese  string `xml:"Portuguese"`
		Catalan     string `xml:"Catalan"`
		ChineseTrad string `xml:"ChineseTrad"`
	} `xml:"TEXT"`
}

type Civ4Defines struct {
	XMLName xml.Name `xml:"Civ4Defines"`
	Text    string   `xml:",chardata"`
	Xmlns   string   `xml:"xmlns,attr"`
	Define  []struct {
		Text            string `xml:",chardata"`
		DefineName      string `xml:"DefineName"`
		IDefineIntVal   string `xml:"iDefineIntVal"`
		DefineTextVal   string `xml:"DefineTextVal"`
		FDefineFloatVal string `xml:"fDefineFloatVal"`
	} `xml:"Define"`
}

type Civ4EraInfos struct {
	XMLName  xml.Name `xml:"Civ4EraInfos"`
	Text     string   `xml:",chardata"`
	Xmlns    string   `xml:"xmlns,attr"`
	EraInfos struct {
		Text    string `xml:",chardata"`
		EraInfo []struct {
			Text                    string `xml:",chardata"`
			Type                    string `xml:"Type"`
			Description             string `xml:"Description"`
			Strategy                string `xml:"Strategy"`
			BNoGoodies              string `xml:"bNoGoodies"`
			BNoAnimals              string `xml:"bNoAnimals"`
			BNoBarbUnits            string `xml:"bNoBarbUnits"`
			BNoBarbCities           string `xml:"bNoBarbCities"`
			IAdvancedStartPoints    string `xml:"iAdvancedStartPoints"`
			IStartingUnitMultiplier string `xml:"iStartingUnitMultiplier"`
			IStartingDefenseUnits   string `xml:"iStartingDefenseUnits"`
			IStartingWorkerUnits    string `xml:"iStartingWorkerUnits"`
			IStartingExploreUnits   string `xml:"iStartingExploreUnits"`
			IStartingGold           string `xml:"iStartingGold"`
			IFreePopulation         string `xml:"iFreePopulation"`
			IStartPercent           string `xml:"iStartPercent"`
			IGrowthPercent          string `xml:"iGrowthPercent"`
			ITrainPercent           string `xml:"iTrainPercent"`
			IConstructPercent       string `xml:"iConstructPercent"`
			ICreatePercent          string `xml:"iCreatePercent"`
			IResearchPercent        string `xml:"iResearchPercent"`
			IBuildPercent           string `xml:"iBuildPercent"`
			IImprovementPercent     string `xml:"iImprovementPercent"`
			IGreatPeoplePercent     string `xml:"iGreatPeoplePercent"`
			ICulturePercent         string `xml:"iCulturePercent"`
			IAnarchyPercent         string `xml:"iAnarchyPercent"`
			IEventChancePerTurn     string `xml:"iEventChancePerTurn"`
			ISoundtrackSpace        string `xml:"iSoundtrackSpace"`
			BFirstSoundtrackFirst   string `xml:"bFirstSoundtrackFirst"`
			EraInfoSoundtracks      struct {
				Text              string   `xml:",chardata"`
				EraInfoSoundtrack []string `xml:"EraInfoSoundtrack"`
			} `xml:"EraInfoSoundtracks"`
			CitySoundscapes struct {
				Text           string `xml:",chardata"`
				CitySoundscape []struct {
					Text             string `xml:",chardata"`
					CitySizeType     string `xml:"CitySizeType"`
					SoundscapeScript string `xml:"SoundscapeScript"`
				} `xml:"CitySoundscape"`
			} `xml:"CitySoundscapes"`
			AudioUnitVictoryScript string `xml:"AudioUnitVictoryScript"`
			AudioUnitDefeatScript  string `xml:"AudioUnitDefeatScript"`
		} `xml:"EraInfo"`
	} `xml:"EraInfos"`
}

type Civ4GameSpeedInfo struct {
	XMLName        xml.Name `xml:"Civ4GameSpeedInfo"`
	Text           string   `xml:",chardata"`
	Xmlns          string   `xml:"xmlns,attr"`
	GameSpeedInfos struct {
		Text          string `xml:",chardata"`
		GameSpeedInfo []struct {
			Text                        string `xml:",chardata"`
			Type                        string `xml:"Type"`
			Description                 string `xml:"Description"`
			Help                        string `xml:"Help"`
			IGrowthPercent              string `xml:"iGrowthPercent"`
			ITrainPercent               string `xml:"iTrainPercent"`
			IConstructPercent           string `xml:"iConstructPercent"`
			ICreatePercent              string `xml:"iCreatePercent"`
			IResearchPercent            string `xml:"iResearchPercent"`
			IBuildPercent               string `xml:"iBuildPercent"`
			IImprovementPercent         string `xml:"iImprovementPercent"`
			IGreatPeoplePercent         string `xml:"iGreatPeoplePercent"`
			ICulturePercent             string `xml:"iCulturePercent"`
			IAnarchyPercent             string `xml:"iAnarchyPercent"`
			IBarbPercent                string `xml:"iBarbPercent"`
			IFeatureProductionPercent   string `xml:"iFeatureProductionPercent"`
			IUnitDiscoverPercent        string `xml:"iUnitDiscoverPercent"`
			IUnitHurryPercent           string `xml:"iUnitHurryPercent"`
			IUnitTradePercent           string `xml:"iUnitTradePercent"`
			IUnitGreatWorkPercent       string `xml:"iUnitGreatWorkPercent"`
			IGoldenAgePercent           string `xml:"iGoldenAgePercent"`
			IHurryPercent               string `xml:"iHurryPercent"`
			IHurryConscriptAngerPercent string `xml:"iHurryConscriptAngerPercent"`
			IInflationPercent           string `xml:"iInflationPercent"`
			IInflationOffset            string `xml:"iInflationOffset"`
			IVictoryDelayPercent        string `xml:"iVictoryDelayPercent"`
			GameTurnInfos               struct {
				Text         string `xml:",chardata"`
				GameTurnInfo []struct {
					Text               string `xml:",chardata"`
					IMonthIncrement    string `xml:"iMonthIncrement"`
					ITurnsPerIncrement string `xml:"iTurnsPerIncrement"`
				} `xml:"GameTurnInfo"`
			} `xml:"GameTurnInfos"`
		} `xml:"GameSpeedInfo"`
	} `xml:"GameSpeedInfos"`
}

type Civ4CalendarInfos struct {
	XMLName       xml.Name `xml:"Civ4CalendarInfos"`
	Text          string   `xml:",chardata"`
	Xmlns         string   `xml:"xmlns,attr"`
	CalendarInfos struct {
		Text         string `xml:",chardata"`
		CalendarInfo []struct {
			Text        string `xml:",chardata"`
			Type        string `xml:"Type"`
			Description string `xml:"Description"`
		} `xml:"CalendarInfo"`
	} `xml:"CalendarInfos"`
}

type Civ4GameOptionInfos struct {
	XMLName         xml.Name `xml:"Civ4GameOptionInfos"`
	Text            string   `xml:",chardata"`
	Xmlns           string   `xml:"xmlns,attr"`
	GameOptionInfos struct {
		Text           string `xml:",chardata"`
		GameOptionInfo []struct {
			Text        string `xml:",chardata"`
			Type        string `xml:"Type"`
			Description string `xml:"Description"`
			Help        string `xml:"Help"`
			BDefault    string `xml:"bDefault"`
			BVisible    string `xml:"bVisible"`
		} `xml:"GameOptionInfo"`
	} `xml:"GameOptionInfos"`
}

type Civ4MPOptionInfos struct {
	XMLName       xml.Name `xml:"Civ4MPOptionInfos"`
	Text          string   `xml:",chardata"`
	Xmlns         string   `xml:"xmlns,attr"`
	MPOptionInfos struct {
		Text         string `xml:",chardata"`
		MPOptionInfo []struct {
			Text        string `xml:",chardata"`
			Type        string `xml:"Type"`
			Description string `xml:"Description"`
			Help        string `xml:"Help"`
			BDefault    string `xml:"bDefault"`
		} `xml:"MPOptionInfo"`
	} `xml:"MPOptionInfos"`
}

type Civ4ForceControlInfos struct {
	XMLName           xml.Name `xml:"Civ4ForceControlInfos"`
	Text              string   `xml:",chardata"`
	Xmlns             string   `xml:"xmlns,attr"`
	ForceControlInfos struct {
		Text             string `xml:",chardata"`
		ForceControlInfo []struct {
			Text        string `xml:",chardata"`
			Type        string `xml:"Type"`
			Description string `xml:"Description"`
			Help        string `xml:"Help"`
			BDefault    string `xml:"bDefault"`
		} `xml:"ForceControlInfo"`
	} `xml:"ForceControlInfos"`
}

type Civ4VictoryInfo struct {
	XMLName      xml.Name `xml:"Civ4VictoryInfo"`
	Text         string   `xml:",chardata"`
	Xmlns        string   `xml:"xmlns,attr"`
	VictoryInfos struct {
		Text        string `xml:",chardata"`
		VictoryInfo []struct {
			Text                   string `xml:",chardata"`
			Type                   string `xml:"Type"`
			Description            string `xml:"Description"`
			Civilopedia            string `xml:"Civilopedia"`
			BTargetScore           string `xml:"bTargetScore"`
			BEndScore              string `xml:"bEndScore"`
			BConquest              string `xml:"bConquest"`
			BDiploVote             string `xml:"bDiploVote"`
			BPermanent             string `xml:"bPermanent"`
			IPopulationPercentLead string `xml:"iPopulationPercentLead"`
			ILandPercent           string `xml:"iLandPercent"`
			IMinLandPercent        string `xml:"iMinLandPercent"`
			IReligionPercent       string `xml:"iReligionPercent"`
			CityCulture            string `xml:"CityCulture"`
			INumCultureCities      string `xml:"iNumCultureCities"`
			ITotalCultureRatio     string `xml:"iTotalCultureRatio"`
			IVictoryDelayTurns     string `xml:"iVictoryDelayTurns"`
			VictoryMovie           string `xml:"VictoryMovie"`
		} `xml:"VictoryInfo"`
	} `xml:"VictoryInfos"`
}
