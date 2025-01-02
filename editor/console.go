package editor

import (
	"fmt"
	"log"
)

var consoleChannel = make(chan string, 10)
var consoleChannelUsed = false

func GetConsoleChannel() <-chan string {
	if !consoleChannelUsed {
		consoleChannelUsed = true
	}

	return consoleChannel
}

func ConsoleWrite(line string, p ...any) {
	line = fmt.Sprintf(line+"\n", p...)
	if consoleChannelUsed {
		consoleChannel <- line
	}

	log.Print(line)
}
