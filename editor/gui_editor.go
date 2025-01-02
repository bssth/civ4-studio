package editor

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/bssth/civ4-studio/resources"
	"image/color"
	"os"
	"strconv"
	"strings"
)

// Editor is the main struct for the editor. It contains all data and methods to work with it
type Editor struct {
	FilePath string
	WbMap    *WbMap
}

const (
	SectionWelcome = iota
	SectionMapSettings
	SectionTeams
	SectionPlayers
)

func (e *Editor) ShowEditor() {
	currentSection := 0

	// Create new empty map if it's not set
	// @todo more real default values
	if e.WbMap == nil {
		e.WbMap = &WbMap{
			Version: 1,
			Game:    &Game{},
		}
	}

	// Firstly, we need to create whole interface and then update it with data
	editor := Application.NewWindow("Civ 4 Studio")
	editor.SetIcon(resources.IconResource)
	editor.Resize(fyne.NewSize(1024, 768))
	editor.CenterOnScreen()
	defer func() {
		if r := recover(); r != nil {
			ConsoleWrite(r.(string))
			Application.Quit()
		}
	}()

	progress := GuiProgressBar()
	body := container.NewVBox()

	// Open another editor section (usually by user), fill content, apply data and callbacks
	openSection := func(section int) {
		body.RemoveAll()
		currentSection = section

		switch section {
		case SectionMapSettings:
			if e.FilePath == "" {
				GuiNoMapLoaded(body)
				break
			}

			GuiSelectEntry(body, "Era", "Starting era", e.WbMap.Game.Era, GetEraNames(), func(s string) { e.WbMap.Game.Era = s })
			GuiSelectEntry(body, "Speed", "Game speed", e.WbMap.Game.Speed, GetSpeedNames(), func(s string) { e.WbMap.Game.Speed = s })
			GuiSelectEntry(body, "Calendar", "Calendar type", e.WbMap.Game.Calendar, GetCalendarNames(), func(s string) { e.WbMap.Game.Calendar = s })

			body.Add(widget.NewSeparator())
			GuiTextField(body, "GameTurn", "Starting turn", strconv.Itoa(int(e.WbMap.Game.GameTurn)), func(s string) { e.WbMap.Game.GameTurn = ToUint(s) })
			GuiTextField(body, "MaxCityElimination", "Maximum cities lost to lose game", strconv.Itoa(int(e.WbMap.Game.MaxCityElimination)), func(s string) { e.WbMap.Game.MaxCityElimination = ToUint(s) })
			GuiTextField(body, "NumAdvancedStartPoints", "Starting points", strconv.Itoa(int(e.WbMap.Game.NumAdvancedStartPoints)), func(s string) { e.WbMap.Game.NumAdvancedStartPoints = ToUint(s) })
			GuiTextField(body, "TargetScore", "Score to win", strconv.Itoa(int(e.WbMap.Game.TargetScore)), func(s string) { e.WbMap.Game.TargetScore = ToUint(s) })
			GuiTextField(body, "StartYear", "Starting year", strconv.Itoa(e.WbMap.Game.StartYear), func(s string) { e.WbMap.Game.StartYear = ToInt(s) })
			GuiTextField(body, "Description", "Description", e.WbMap.Game.Description, func(s string) { e.WbMap.Game.Description = s })
			GuiTextField(body, "ModPath", "Mod path", e.WbMap.Game.ModPath, func(s string) { e.WbMap.Game.ModPath = s })
			GuiTextField(body, "MaxTurns", "Turns to end game", strconv.Itoa(int(e.WbMap.Game.MaxTurns)), func(s string) { e.WbMap.Game.MaxTurns = ToUint(s) })
			GuiCheckbox(body, "Tutorial", e.WbMap.Game.Tutorial, func(b bool) { e.WbMap.Game.Tutorial = b })

			body.Add(widget.NewSeparator())
			// @todo checkboxes not working, T[key] is always the last value
			for _, key := range SortKeys(VictoryInfos) {
				GuiCheckbox(body, "Victory: "+GetLangString(VictoryInfos[key].Description), IsInSlice(e.WbMap.Game.Victory, VictoryInfos[key].Type),
					func(b bool) { e.WbMap.Game.Victory = SwitchInSlice(b, e.WbMap.Game.Victory, VictoryInfos[key].Type) })
			}
			body.Add(widget.NewSeparator())
			for _, key := range SortKeys(GameOptionInfos) {
				GuiCheckbox(body, "Game option: "+GetLangString(GameOptionInfos[key].Description), IsInSlice(e.WbMap.Game.Option, GameOptionInfos[key].Type),
					func(b bool) { e.WbMap.Game.Option = SwitchInSlice(b, e.WbMap.Game.Option, GameOptionInfos[key].Type) })
			}
			body.Add(widget.NewSeparator())
			for _, key := range SortKeys(GameMPInfos) {
				GuiCheckbox(body, "Multiplayer option: "+GetLangString(GameMPInfos[key].Description), IsInSlice(e.WbMap.Game.MPOption, GameMPInfos[key].Type),
					func(b bool) { e.WbMap.Game.MPOption = SwitchInSlice(b, e.WbMap.Game.MPOption, GameMPInfos[key].Type) })
			}
			body.Add(widget.NewSeparator())
			for _, key := range SortKeys(ForceControlInfos) {
				GuiCheckbox(body, "Make unchangeable: "+GetLangString(ForceControlInfos[key].Description), IsInSlice(e.WbMap.Game.ForceControl, ForceControlInfos[key].Type),
					func(b bool) {
						e.WbMap.Game.ForceControl = SwitchInSlice(b, e.WbMap.Game.ForceControl, ForceControlInfos[key].Type)
					})
			}

		case SectionTeams:
			if e.FilePath == "" {
				GuiNoMapLoaded(body)
				break
			}

			fallthrough
		case SectionPlayers:
			if e.FilePath == "" {
				GuiNoMapLoaded(body)
				break
			}

			fallthrough
		default:
			currentSection = 0
			body.Add(container.NewCenter(canvas.NewText("Start with selecting what to edit", color.White)))
		}
	}

	// Menu with sections to switch between
	menu := container.NewVBox(
		widget.NewButton("Map settings", func() { openSection(SectionMapSettings) }),
		widget.NewButton("Teams", func() { openSection(SectionTeams) }),
		widget.NewButton("Players", func() { openSection(SectionPlayers) }),
	)

	// Some kind of reactivity to update all content when something changes externally (like file loaded etc.)
	updateContent := func() {
		openSection(currentSection)
	}
	updateAll := func() {
		updateContent()
	}

	// Open map file, parse it and load into editor
	openFile := func() {
		dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			progress.Start("Loading and parsing map...")
			defer progress.Stop()
			if err != nil {
				ConsoleWrite(err.Error())
				dialog.ShowError(err, editor)
				return
			}
			if reader == nil {
				return
			}

			defer reader.Close()
			wbMap, err := ParseWbMap(reader)
			if err != nil {
				errorText := err.Error()
				ConsoleWrite(errorText)

				// Truncate error text if it's too long for dialog
				if len(errorText) > 100 {
					errorText = errorText[:100] + "..."
				}
				dialog.ShowError(errors.New(errorText), editor)
				return
			}

			e.WbMap, e.FilePath = wbMap, reader.URI().Path()
			currentSection = SectionWelcome // not using openSection() to prevent reload, we'll do it manually in the next line
			updateAll()
		}, editor).Show()
	}

	saveFile := func() {
		if e.WbMap == nil || e.WbMap.Game == nil {
			return
		}

		progress.Start("Saving...")
		defer progress.Stop()

		if e.FilePath == "" {
			dialog.NewFileSave(func(writer fyne.URIWriteCloser, err error) {
				if err != nil {
					ConsoleWrite(err.Error())
					dialog.ShowError(err, editor)
					return
				}
				if writer == nil {
					return
				}

				defer writer.Close()
				_, err = writer.Write(e.WbMap.ToWbFormat())
				if err != nil {
					ConsoleWrite(err.Error())
					dialog.ShowError(err, editor)
				}
			}, editor).Show()
		} else {
			err := os.WriteFile(e.FilePath, e.WbMap.ToWbFormat(), 0644)
			if err != nil {
				ConsoleWrite(err.Error())
				dialog.ShowError(err, editor)
				return
			}
		}
	}

	toolbar := widget.NewToolbar(
		widget.NewToolbarAction(theme.FolderOpenIcon(), openFile),
		widget.NewToolbarAction(theme.DocumentSaveIcon(), saveFile),
		widget.NewToolbarAction(theme.MediaPlayIcon(), func() {
			err := LaunchGame(e.FilePath)
			if err != nil {
				ConsoleWrite(err.Error())
				dialog.ShowError(err, editor)
				return
			}
		}),
		widget.NewToolbarSpacer(),
		widget.NewToolbarAction(theme.SettingsIcon(), func() {
			/*closed := ShowSettings()
			go func() {
				if <-closed {
					updateAll()
				}
			}()*/
		}),
		widget.NewToolbarAction(theme.ViewFullScreenIcon(), func() {
			editor.SetFullScreen(!editor.FullScreen())
		}),
	)

	window := container.NewBorder(toolbar, progress.GetBlock(), menu, nil, container.NewScroll(body))
	editor.SetContent(window)
	editor.Show()
	editor.SetMaster()

	// After creating interfaces and callbacks, we need to load game data and invite user to open map file
	go func() {
		progress.Start("Parsing game XML files")
		// Show what file is parsed at the moment
		parsingProgressHandler := func(s string) {
			// Remove game dir from path to make it shorter (better UX)
			s = strings.ReplaceAll(s, GlobalConfig.GameDir+"\\", "")
			s = strings.ReplaceAll(s, GlobalConfig.GameDir+string(os.PathSeparator), "")
			s = strings.ReplaceAll(s, GlobalConfig.GameDir, "")
			if len(s) > 170 {
				s = s[:170] + "..."
			}

			progress.Stop()
			progress.Start(s)
		}

		err := LoadAllXML(parsingProgressHandler)
		if err != nil {
			ConsoleWrite(err.Error())
			dialog.ShowError(err, editor)
			return
		}

		progress.Stop()

		// Select/open test file if not already set
		if e.FilePath == "" {
			// Open test file if it exists. It's developed for testing and debugging purposes
			if *DevMode && e.testFileExists() {
				e.FilePath, e.WbMap = e.openTestFile()
				if e.WbMap != nil {
					updateAll()
					return
				}
			}

			openFile()
		}
	}()
}

const testFilePath = "resources/test.CivBeyondSwordWBSave"

func (e *Editor) testFileExists() bool {
	_, err := os.Stat(testFilePath)
	return err == nil
}

func (e *Editor) openTestFile() (string, *WbMap) {
	file, err := os.Open(testFilePath)
	if err != nil {
		return "", nil
	}

	defer file.Close()

	wb, err := ParseWbMap(file)
	if err != nil {
		return "", nil
	}

	return testFilePath, wb
}
