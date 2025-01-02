package editor

import (
	"os"
	"testing"
)

func TestParseWbMap(t *testing.T) {
	file, err := os.Open("../" + testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	defer file.Close()

	wb, err := ParseWbMap(file)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if wb.Game == nil {
		t.Fatalf("Failed to parse game section")
	}

	if len(wb.Players) == 0 {
		t.Fatalf("Failed to parse players section")
	}

	if len(wb.Teams) == 0 {
		t.Fatalf("Failed to parse teams section")
	}

	if len(wb.Plots) == 0 {
		t.Fatalf("Failed to parse plots section")
	}
}
