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
		// Non-blocking send: drop the line if no one is reading fast enough
		// rather than freezing the whole pipeline.
		select {
		case consoleChannel <- line:
		default:
		}
	}

	log.Print(line)
}
