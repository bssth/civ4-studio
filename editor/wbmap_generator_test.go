package editor

import (
	"os"
	"testing"
)

const expectedResultLength = 1900000

func TestSimpleGenerator(t *testing.T) {
	file, err := os.Open("../" + testFilePath)
	if err != nil {
		t.Fatalf("Failed to open test file: %v", err)
	}

	defer file.Close()

	wb, err := ParseWbMap(file)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(wb.ToWbFormat()) < expectedResultLength {
		t.Fatalf("Failed to convert WbMap to WorldBuilder format: too short (%d bytes, >%d expected)", len(wb.ToWbFormat()), expectedResultLength)
	}
}
