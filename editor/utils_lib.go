package editor

import (
	"os"
	"sort"
	"strconv"
)

// IsDir reports whether path exists and is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// ToInt converts a string to an integer
func ToInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

// ToUint converts a string to an unsigned integer
func ToUint(s string) uint {
	return uint(ToInt(s))
}

// IsInSlice checks if a string is in a slice
func IsInSlice(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}

	return false
}

// SwitchInSlice adds or removes a string from a slice depending on the "add" parameter
func SwitchInSlice(add bool, slice []string, value string) []string {
	if add {
		return AddToSlice(slice, value)
	} else {
		return RemoveFromSlice(slice, value)
	}
}

// AddToSlice adds a string to a slice if it's not already there
func AddToSlice(slice []string, value string) []string {
	for _, v := range slice {
		if v == value {
			return slice
		}
	}

	return append(slice, value)
}

// RemoveFromSlice removes a string from a slice
func RemoveFromSlice(slice []string, value string) []string {
	for i, v := range slice {
		if v == value {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}

// SortKeys returns the keys of a map sorted alphabetically
func SortKeys[T any](dict map[string]T) []string {
	keys := make([]string, 0)
	for k, _ := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// BoolToInt converts a boolean to an integer
func BoolToInt(b bool) int {
	if b {
		return 1
	} else {
		return 0
	}
}
