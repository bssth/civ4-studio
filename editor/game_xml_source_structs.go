package editor

import "encoding/xml"

// This file contains the structs for representing game's XML files.

// Civ4GameText is a language file (Assets/XML/Text/*.xml)
type Civ4GameText struct {
	TEXT []struct {
		Tag     string   `xml:"Tag"`
		English langText `xml:"English"`
	} `xml:"TEXT"`
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
}
