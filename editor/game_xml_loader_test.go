package editor

import (
	"os"
	"path/filepath"
	"testing"
)

func writeTestFile(t *testing.T, path string, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// fakeGameInstall creates base game, BtS and mod folders with a few XML files
func fakeGameInstall(t *testing.T) (btsDir string) {
	root := t.TempDir()
	btsDir = filepath.Join(root, "Beyond the Sword")

	writeTestFile(t, filepath.Join(root, XmlDir, "Text", "CIV4GameTextInfos.xml"), `<?xml version="1.0" encoding="ISO-8859-1"?>
<Civ4GameText xmlns="http://www.firaxis.com">
	<TEXT><Tag>TXT_KEY_CIV_AMERICA_DESC</Tag><English>American Empire</English><French>Empire américain</French></TEXT>
	<TEXT><Tag>TXT_KEY_CIV_AMERICA_ADJECTIVE</Tag><English><Text>American</Text><Gender>Male</Gender><Plural>0</Plural></English></TEXT>
	<TEXT><Tag>TXT_KEY_LEADER_WASHINGTON</Tag><English>Washington</English></TEXT>
</Civ4GameText>`)
	writeTestFile(t, filepath.Join(root, XmlDir, "Technologies", "CIV4TechInfos.xml"), `<Civ4TechInfos xmlns="x-schema:CIV4TechnologiesSchema.xml">
	<TechInfos>
		<TechInfo><Type>TECH_MINING</Type><Description>TXT_KEY_TECH_MINING</Description><Era>ERA_ANCIENT</Era></TechInfo>
		<TechInfo><Type>TECH_OLD</Type><Description>TXT_KEY_TECH_OLD</Description><Era>ERA_ANCIENT</Era></TechInfo>
	</TechInfos>
</Civ4TechInfos>`)
	writeTestFile(t, filepath.Join(btsDir, XmlDir, "Civilizations", "CIV4CivilizationInfos.xml"), `<Civ4CivilizationInfos>
	<CivilizationInfos>
		<CivilizationInfo>
			<Type>CIVILIZATION_AMERICA</Type>
			<Description>TXT_KEY_CIV_AMERICA_DESC</Description>
			<ShortDescription>TXT_KEY_CIV_AMERICA_SHORT_DESC</ShortDescription>
			<Adjective>TXT_KEY_CIV_AMERICA_ADJECTIVE</Adjective>
			<DefaultPlayerColor>PLAYERCOLOR_WHITE</DefaultPlayerColor>
			<ArtStyleType>ARTSTYLE_EUROPEAN</ArtStyleType>
			<bPlayable>1</bPlayable>
			<Leaders>
				<Leader><LeaderName>LEADER_WASHINGTON</LeaderName><bLeaderAvailability>1</bLeaderAvailability></Leader>
				<Leader><LeaderName>LEADER_LINCOLN</LeaderName><bLeaderAvailability>1</bLeaderAvailability></Leader>
			</Leaders>
		</CivilizationInfo>
	</CivilizationInfos>
</Civ4CivilizationInfos>`)
	writeTestFile(t, filepath.Join(btsDir, XmlDir, "GlobalTypes.xml"), `<Civ4Types>
	<DomainTypes><DomainType>DOMAIN_SEA</DomainType></DomainTypes>
	<ArtStyleTypes><ArtStyleType>ARTSTYLE_EUROPEAN</ArtStyleType><ArtStyleType>ARTSTYLE_ASIAN</ArtStyleType></ArtStyleTypes>
</Civ4Types>`)
	writeTestFile(t, filepath.Join(btsDir, XmlDir, "GameInfo", "CIV4CivicInfos.xml"), `<Civ4CivicInfos>
	<CivicInfos>
		<CivicInfo><CivicOptionType>CIVICOPTION_GOVERNMENT</CivicOptionType><Type>CIVIC_DESPOTISM</Type><Description>TXT_KEY_CIVIC_DESPOTISM</Description></CivicInfo>
	</CivicInfos>
</Civ4CivicInfos>`)
	writeTestFile(t, filepath.Join(btsDir, BtsExe), "")
	writeTestFile(t, filepath.Join(btsDir, PublicMapsDir, ".keep"), "")

	// The mod replaces the tech file completely
	writeTestFile(t, filepath.Join(btsDir, ModsDir, "TestMod", XmlDir, "Technologies", "CIV4TechInfos.xml"), `<Civ4TechInfos>
	<TechInfos>
		<TechInfo><Type>TECH_MINING</Type><Description>TXT_KEY_TECH_MINING</Description><Era>ERA_ANCIENT</Era></TechInfo>
		<TechInfo><Type>TECH_STONE_TOOLS</Type><Description>TXT_KEY_TECH_STONE_TOOLS</Description><Era>ERA_PREHISTORIC</Era></TechInfo>
	</TechInfos>
</Civ4TechInfos>`)
	writeTestFile(t, filepath.Join(btsDir, ModsDir, "TestMod", XmlDir, "Broken.xml"), `<Civ4TechInfos><TechInfos><TechInfo>`)

	return btsDir
}

func withConfig(t *testing.T, config Config) {
	old := GlobalConfig
	GlobalConfig = &config
	t.Cleanup(func() { GlobalConfig = old })
}

func TestLoadAllXML(t *testing.T) {
	btsDir := fakeGameInstall(t)
	withConfig(t, Config{GameDir: btsDir, Mod: "TestMod"})

	data, err := LoadAllXML(nil)
	if err != nil {
		t.Fatal(err)
	}

	if got := data.Text("TXT_KEY_CIV_AMERICA_ADJECTIVE"); got != "American" {
		t.Errorf("nested English text: got %q", got)
	}
	if got := data.Text("TXT_KEY_CIV_AMERICA_DESC"); got != "American Empire" {
		t.Errorf("base game text: got %q", got)
	}

	civ := data.Table(InfoCivilizations).Get("CIVILIZATION_AMERICA")
	if civ == nil {
		t.Fatal("civilization not loaded")
	}
	if len(civ.Leaders) != 2 || civ.DefaultPlayerColor != "PLAYERCOLOR_WHITE" || !civ.Playable {
		t.Errorf("civilization parsed incorrectly: %+v", civ)
	}

	techs := data.Table(InfoTechs).All()
	if len(techs) != 2 || techs[1].Type != "TECH_STONE_TOOLS" || techs[1].Group != "ERA_PREHISTORIC" {
		t.Errorf("mod tech file must replace the base one: %+v", techs)
	}

	if styles := data.Table(InfoArtStyles).All(); len(styles) != 2 || styles[0].Type != "ARTSTYLE_EUROPEAN" {
		t.Errorf("art styles parsed incorrectly: %+v", styles)
	}

	if civic := data.Table(InfoCivics).Get("CIVIC_DESPOTISM"); civic == nil || civic.Group != "CIVICOPTION_GOVERNMENT" {
		t.Errorf("civic parsed incorrectly: %+v", civic)
	}
}

func TestGetCivilizationsAndOptions(t *testing.T) {
	btsDir := fakeGameInstall(t)
	withConfig(t, Config{GameDir: btsDir})

	data, err := LoadAllXML(nil)
	if err != nil {
		t.Fatal(err)
	}
	SetGameData(data)
	t.Cleanup(func() { SetGameData(NewGameData()) })

	app := NewApp()
	civs := app.GetCivilizations()
	if len(civs) != 1 || civs[0].Description != "American Empire" || civs[0].Adjective != "American" {
		t.Errorf("unexpected civilizations: %+v", civs)
	}

	options := app.GetOptions()
	if len(options["techs"]) != 2 {
		t.Errorf("without the mod base techs are expected: %+v", options["techs"])
	}
	if colors := options["artStyles"]; len(colors) != 2 || colors[1].Description != "Asian" {
		t.Errorf("types without description must be humanized: %+v", colors)
	}
}

func TestHumanizeType(t *testing.T) {
	if got := HumanizeType("PLAYERCOLOR_DARK_RED"); got != "Dark Red" {
		t.Errorf("got %q", got)
	}
	if got := HumanizeType("NONE"); got != "None" {
		t.Errorf("got %q", got)
	}
}
