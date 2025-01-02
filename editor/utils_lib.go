package editor

import (
	"bufio"
	"encoding/xml"
	"golang.org/x/net/html/charset"
	"os"
	"regexp"
	"sort"
	"strconv"
)

// WaitCloseChan is a channel that will be closed when the UI window is closed
type WaitCloseChan <-chan bool

// xmlTypeRegex is used to determine the XML file type by reading the first line
var xmlTypeRegex = regexp.MustCompile("<([a-zA-Z0-9]+)[ |>]")

// ParseXMLFromFile opens a file and returns an XML decoder and the Civilization XML file type
func ParseXMLFromFile(path string) (decoder *xml.Decoder, tag CivXmlType, err error) {
	var xmlFile *os.File
	xmlFile, err = os.Open(path)
	if err != nil {
		ConsoleWrite("Error opening file: %s", err)
		return
	}

	// Read line-by-line to find the XML file type. It's needed to determine which struct to use
	// No need to parse the whole file, first line is enough
	sc := bufio.NewScanner(xmlFile)
	for sc.Scan() {
		matches := xmlTypeRegex.FindStringSubmatch(sc.Text())
		if len(matches) > 1 {
			tag = CivXmlType(matches[1])
			break
		}
	}

	xmlFile.Close()
	// Reopen file to parse it from the beginning
	xmlFile, err = os.Open(path)
	if err != nil {
		ConsoleWrite("Error reopening file: %s", err)
		return
	}

	decoder = xml.NewDecoder(xmlFile)
	decoder.CharsetReader = charset.NewReaderLabel // needed for non-UTF-8 files, sometimes it's something like "iso-8859-1"
	decoder.Strict = false                         // a bit faster and safer because some mod files are not strictly valid XML
	return
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
