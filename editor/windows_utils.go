//go:build windows

package editor

import (
	"golang.org/x/sys/windows"
	"os"
	"strings"
	"syscall"
)

// RunElevated runs an executable with elevated privileges. Windows only
func RunElevated(exe string, argv []string) error {
	exe = "\"" + exe + "\""
	verb := "runas"
	cwd, _ := os.Getwd()
	args := strings.Join(argv, " ")

	ConsoleWrite("Running %s %s", exe, args)

	verbPtr, _ := syscall.UTF16PtrFromString(verb)
	exePtr, _ := syscall.UTF16PtrFromString(exe)
	cwdPtr, _ := syscall.UTF16PtrFromString(cwd)
	argPtr, _ := syscall.UTF16PtrFromString(args)

	return windows.ShellExecute(0, verbPtr, exePtr, argPtr, cwdPtr, 1)
}
