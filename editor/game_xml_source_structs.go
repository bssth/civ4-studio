package editor

import (
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"
)

// This file contains the structs for representing game's XML files.

// Civ4GameText is a language file (Assets/XML/Text/*.xml). Every <TEXT> has a <Tag> and one element per
// language: <English>, <French>, <German>, <Italian>, <Spanish>, and others added by localizations.
type Civ4GameText struct {
	TEXT []struct {
		Tag       string        `xml:"Tag"`
		Languages []langElement `xml:",any"`
	} `xml:"TEXT"`
}

// langElement is the text of one language, the language is the element name
type langElement struct {
	XMLName xml.Name
	langText
}

// langText is either a plain string or a <Text> element with gender/plural attributes, e.g.
// <English><Text>American</Text><Gender>Male</Gender><Plural>0</Plural></English>
type langText struct {
	Value string `xml:",chardata"`
	Text  string `xml:"Text"`
}

func (t langText) String() string {
	if t.Text != "" {
		return t.Text
	}
	return t.Value
}

type Civ4Defines struct {
	Define []struct {
		DefineName      string `xml:"DefineName"`
		IDefineIntVal   string `xml:"iDefineIntVal"`
		DefineTextVal   string `xml:"DefineTextVal"`
		FDefineFloatVal string `xml:"fDefineFloatVal"`
	} `xml:"Define"`
}

// civ4InfoFile matches any info file of the form <Root><Category><Entry>...</Entry></Category></Root>.
// Only fields used by the editor are decoded, everything else is skipped.
type civ4InfoFile struct {
	Categories []struct {
		XMLName xml.Name
		Entries []civ4InfoEntry `xml:",any"`
	} `xml:",any"`
}

type civ4InfoEntry struct {
	// Value is used by plain lists like <ArtStyleTypes><ArtStyleType>ARTSTYLE_EUROPE</ArtStyleType></ArtStyleTypes>
	Value              string   `xml:",chardata"`
	Type               string   `xml:"Type"`
	Description        string   `xml:"Description"`
	ShortDescription   string   `xml:"ShortDescription"`
	Adjective          string   `xml:"Adjective"`
	DefaultPlayerColor string   `xml:"DefaultPlayerColor"`
	ArtStyleType       string   `xml:"ArtStyleType"`
	Playable           string   `xml:"bPlayable"`
	Era                string   `xml:"Era"`
	CivicOptionType    string   `xml:"CivicOptionType"`
	Leaders            []string `xml:"Leaders>Leader>LeaderName"`
	GridWidth          int      `xml:"iGridWidth"`
	GridHeight         int      `xml:"iGridHeight"`
	ColorTypePrimary   string   `xml:"ColorTypePrimary"`
	Water              string   `xml:"bWater"`
	Red                string   `xml:"fRed"`
	Green              string   `xml:"fGreen"`
	Blue               string   `xml:"fBlue"`
}

// rgb converts color components (0..1 floats) to "#rrggbb", empty if there is no color
func (e civ4InfoEntry) rgb() string {
	if e.Red == "" && e.Green == "" && e.Blue == "" {
		return ""
	}
	component := func(s string) int {
		f, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
		return int(math.Round(math.Max(0, math.Min(1, f)) * 255))
	}
	return fmt.Sprintf("#%02x%02x%02x", component(e.Red), component(e.Green), component(e.Blue))
}
