package editor

import "testing"

func TestGetDefaultConfig(t *testing.T) {
	testConf := &Config{AutoSave: true}
	config := GetDefaultConfig()
	if config != *testConf {
		t.Error("GetDefaultConfig() did not return the expected default configuration")
	}
}
