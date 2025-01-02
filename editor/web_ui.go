package editor

import (
	"context"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"time"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	/*
			argsWithoutProg := os.Args[1:]

		    if len(argsWithoutProg) != 0 {
		    println("launchArgs", argsWithoutProg)
			}
	*/

	a.ctx = ctx

	console := GetConsoleChannel()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case line := <-console:
				runtime.EventsEmit(ctx, "console", line)
			}
		}
	}()

	config, fileExists := GetConfig()
	GlobalConfig = &config
	if !fileExists {
		ConsoleWrite("Config file not found, opening settings")
	} else {
		ConsoleWrite("Config file found, opening editor")
	}

	go func() {
		ConsoleWrite(time.Now().String())
		<-time.After(1 * time.Second)
	}()
}

func (a *App) WriteConsole(line string) {
	ConsoleWrite(line)
}

func (a *App) GetConfig() *Config {
	return GlobalConfig
}

func (a *App) SetConfig(config *Config) {
	GlobalConfig = config
	err := SaveConfig()
	if err != nil {
		ConsoleWrite(err.Error())
	}
}

func (a *App) GetModsList() []string {
	return GetModsList(GlobalConfig.GameDir)
}
