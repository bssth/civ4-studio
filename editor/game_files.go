package editor

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const BtsExe = "Civ4BeyondSword.exe"
const AssetsDir = "Assets"
const PublicMapsDir = "PublicMaps"
const PrivateMapsDir = "PrivateMaps"
const ModsDir = "Mods"

// CheckGameDirectory checks if game directory contains executable and required directories
func CheckGameDirectory(path string) error {
	for _, pathChecking := range []string{BtsExe, AssetsDir, PublicMapsDir} {
		if _, err := os.Stat(path + string(os.PathSeparator) + pathChecking); err != nil {
			return errors.New(pathChecking + " not found")
		}
	}

	return nil
}

// IsMod returns true if mod is enabled if configuration
func IsMod() bool {
	return GlobalConfig.Mod != ""
}

// GetExe returns path to game executable
func GetExe() string {
	return GlobalConfig.GameDir + string(os.PathSeparator) + BtsExe
}

// GetModDir returns a path to mod directory if mod is enabled, otherwise returns game directory
func GetModDir() string {
	return GlobalConfig.GameDir + string(os.PathSeparator) + GetRelativeModDir()
}

// GetRelativeModDir returns a path to mod directory relative to game directory
func GetRelativeModDir() string {
	if !IsMod() {
		return ""
	}

	return ModsDir + string(os.PathSeparator) + GlobalConfig.Mod
}

// GetPublicMapsDir returns a path to public maps directory
func GetPublicMapsDir() string {
	return GetModDir() + string(os.PathSeparator) + PublicMapsDir
}

// GetRootDirs returns list of directories where to search for game files (game directory + mod directory)
func GetRootDirs() []string {
	dirs := []string{GlobalConfig.GameDir}
	if IsMod() {
		dirs = append(dirs, GetModDir())
	}

	return dirs
}

// GetModsList returns list of mods in game directory
func GetModsList(path string) []string {
	var mods []string

	files, err := os.ReadDir(path + string(os.PathSeparator) + ModsDir)
	if err != nil {
		return mods
	}

	for _, file := range files {
		if file.IsDir() {
			mods = append(mods, file.Name())
		}
	}

	return mods
}

// GetFilesFromGameDirsRecursive returns a list of files by path (filepath or directory to scan) from all game directories (core game + mods) recursively
func GetFilesFromGameDirsRecursive(path string, ext string) (files []string, err error) {
	var currentErr error

	for _, dir := range GetRootDirs() {
		subDir := dir + string(os.PathSeparator) + path

		currentErr = filepath.Walk(subDir, func(walkFileName string, f os.FileInfo, err error) error {
			if !strings.HasSuffix(strings.ToLower(walkFileName), "."+ext) {
				return nil
			}

			files = append(files, walkFileName)
			return nil
		})

		if currentErr != nil {
			err = fmt.Errorf("%w\n%s", err, currentErr)
		}
	}

	return
}

// GetFilesFromGameDirs returns a list of files by path (filepath or directory to scan) from all game directories (core game + mods)
func GetFilesFromGameDirs(path string) []string {
	var files []string
	for _, dir := range GetRootDirs() {
		parent := dir + string(os.PathSeparator) + path
		if file, err := os.Stat(parent); err != nil || !file.IsDir() {
			if err != nil {
				return files
			}
			if !file.IsDir() {
				files = append(files, parent)
				continue
			}
		}

		readDir, err := os.ReadDir(parent)
		if err != nil {
			ConsoleWrite(err.Error())
			continue
		}

		for _, file := range readDir {
			if file.IsDir() {
				continue
			}

			files = append(files, parent+string(os.PathSeparator)+file.Name())
		}
	}

	return files
}

// LaunchGame starts the game with current configured mod
func LaunchGame(mapFileName string) error {
	if err := CheckGameDirectory(GlobalConfig.GameDir); err != nil {
		return err
	}

	var argv []string
	var err error
	if GlobalConfig.Mod != "" {
		argv = append(argv, "mod=\" "+GlobalConfig.Mod+"\"")
	}

	if mapFileName != "" {
		absPath, err := filepath.Abs(mapFileName)
		if err == nil {
			mapFileName = absPath
		}

		argv = append(argv, "/FXSLOAD=\\\"\""+absPath+"\"")
	}

	if runtime.GOOS != "windows" {
		var game *os.Process
		game, err = os.StartProcess(GetExe(), argv, &os.ProcAttr{
			Dir:   GlobalConfig.GameDir,
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Sys:   nil,
		})
		if err != nil {
			return err
		}

		go func() {
			wait, err := game.Wait()
			if err != nil {
				ConsoleWrite(err.Error())
				return
			}

			ConsoleWrite("Game exited with code %d", wait.ExitCode())
		}()
	} else {
		return RunElevated(GetExe(), argv)
	}

	return nil
}
