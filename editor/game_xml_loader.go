package editor

import (
	"encoding/xml"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
)

const XmlDir = "Assets/XML"
const XmlExt = ".xml"

// WarlordsDir is the Warlords expansion folder next to Beyond the Sword
const WarlordsDir = "Warlords"

const (
	xmlRootGameText = "civ4gametext"
	xmlRootDefines  = "civ4defines"
	xmlRootTypes    = "civ4types"
)

var knownCategories = func() map[string]bool {
	result := make(map[string]bool)
	for _, category := range KnownInfoCategories {
		result[category] = true
	}
	return result
}()

// GetXMLRootDirs returns directories to load XML from, in the order the game applies them:
// base game, Warlords, Beyond the Sword and the mod. BtS standalone installations keep part of
// the files (including most of the texts) in base game and Warlords folders.
func GetXMLRootDirs() []string {
	var dirs []string
	parent := filepath.Dir(filepath.Clean(GlobalConfig.GameDir))
	for _, dir := range []string{parent, filepath.Join(parent, WarlordsDir), GlobalConfig.GameDir} {
		if IsDir(filepath.Join(dir, XmlDir)) && !IsInSlice(dirs, dir) {
			dirs = append(dirs, dir)
		}
	}
	if IsMod() {
		dirs = append(dirs, GetModDir())
	}
	return dirs
}

// CollectXMLFiles returns XML files from all root dirs. A file with the same relative path
// in a later dir (e.g. in a mod) replaces the earlier one, the same way the game does it.
func CollectXMLFiles(rootDirs []string) ([]string, error) {
	var files []string
	index := make(map[string]int)

	for _, root := range rootDirs {
		base := filepath.Join(root, XmlDir)
		err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				if path == base && errors.Is(err, fs.ErrNotExist) {
					return nil
				}
				return err
			}
			if d.IsDir() || !strings.EqualFold(filepath.Ext(path), XmlExt) {
				return nil
			}

			rel, err := filepath.Rel(base, path)
			if err != nil {
				return err
			}
			key := strings.ToLower(filepath.ToSlash(rel))
			if i, ok := index[key]; ok {
				files[i] = path
			} else {
				index[key] = len(files)
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return files, nil
}

// LoadAllXML loads game XML files (texts, defines and info files) from the game and mod directories.
// progressHandler is called before each file is parsed (for UI updates), it may be nil.
func LoadAllXML(progressHandler func(string)) (*GameData, error) {
	files, err := CollectXMLFiles(GetXMLRootDirs())
	if err != nil {
		return nil, err
	}
	return LoadXMLFiles(files, progressHandler), nil
}

// LoadXMLFiles parses given files into a new GameData. Broken files are reported to the console and skipped.
func LoadXMLFiles(files []string, progressHandler func(string)) *GameData {
	if progressHandler == nil {
		progressHandler = func(string) {}
	}

	data := NewGameData()
	texts := 0
	for _, f := range files {
		progressHandler("Parsing XML: " + f)
		n, err := loadXMLFile(data, f)
		if err != nil {
			ConsoleWrite("Cannot parse %s: %s", f, err.Error())
		}
		texts += n
	}

	ConsoleWrite("Loaded %d texts", texts)
	for _, category := range KnownInfoCategories {
		if n := data.Table(category).Len(); n > 0 {
			ConsoleWrite("Loaded %d %s", n, category)
		}
	}
	return data
}

func loadXMLFile(data *GameData, path string) (int, error) {
	root, firstChild, err := peekXML(path)
	if err != nil {
		return 0, err
	}

	switch {
	case strings.EqualFold(root, xmlRootGameText):
		text := &Civ4GameText{}
		if err := decodeXMLFile(path, text); err != nil {
			return 0, err
		}
		for _, t := range text.TEXT {
			// @todo multiple languages
			data.Lang[t.Tag] = strings.TrimSpace(t.English.String())
		}
		return len(text.TEXT), nil

	case strings.EqualFold(root, xmlRootDefines):
		defines := &Civ4Defines{}
		if err := decodeXMLFile(path, defines); err != nil {
			return 0, err
		}
		assignGlobalDefines(data, defines)

	case strings.EqualFold(root, xmlRootTypes) || knownCategories[firstChild]:
		info := &civ4InfoFile{}
		if err := decodeXMLFile(path, info); err != nil {
			return 0, err
		}
		for _, category := range info.Categories {
			if knownCategories[category.XMLName.Local] {
				addInfoEntries(data.Infos[category.XMLName.Local], category.Entries)
			}
		}
	}

	return 0, nil
}

func addInfoEntries(table *InfoTable, entries []civ4InfoEntry) {
	for _, e := range entries {
		infoType := strings.TrimSpace(e.Type)
		if infoType == "" {
			infoType = strings.TrimSpace(e.Value)
		}
		if infoType == "" {
			continue
		}

		group := e.Era
		if e.CivicOptionType != "" {
			group = e.CivicOptionType
		}

		table.Set(&TypeInfo{
			Type:               infoType,
			Description:        e.Description,
			Group:              group,
			ShortDescription:   e.ShortDescription,
			Adjective:          e.Adjective,
			DefaultPlayerColor: e.DefaultPlayerColor,
			ArtStyleType:       e.ArtStyleType,
			Leaders:            e.Leaders,
			Playable:           e.Playable == "1",
		})
	}
}

func newXMLDecoder(r io.Reader) *xml.Decoder {
	decoder := xml.NewDecoder(r)
	decoder.CharsetReader = charset.NewReaderLabel // needed for non-UTF-8 files, sometimes it's something like "iso-8859-1"
	decoder.Strict = false                         // some mod files are not strictly valid XML
	return decoder
}

// peekXML returns names of the root element and its first child without reading the whole file
func peekXML(path string) (root string, firstChild string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	decoder := newXMLDecoder(f)
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return root, "", nil
		}
		if err != nil {
			return root, "", err
		}
		if start, ok := token.(xml.StartElement); ok {
			if root == "" {
				root = start.Name.Local
				continue
			}
			return root, start.Name.Local, nil
		}
	}
}

func decodeXMLFile(path string, v any) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return newXMLDecoder(f).Decode(v)
}

// assignGlobalDefines copies all defines from the Civ4Defines struct to game data
func assignGlobalDefines(data *GameData, definesStruct *Civ4Defines) {
	for _, define := range definesStruct.Define {
		if define.IDefineIntVal != "" {
			data.IntDefines[define.DefineName], _ = strconv.Atoi(define.IDefineIntVal)
		} else if define.FDefineFloatVal != "" {
			data.FloatDefines[define.DefineName], _ = strconv.ParseFloat(define.FDefineFloatVal, 64)
		} else if define.DefineTextVal != "" {
			data.StringDefines[define.DefineName] = define.DefineTextVal
		}
	}
}
