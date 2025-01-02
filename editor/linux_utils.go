//go:build !windows

package editor

import "errors"

func RunElevated(exe string, argv []string) error {
	return errors.New("elevated execution is not supported on this platform")
}
