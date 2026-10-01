package editor

import "testing"

func TestAppTitle(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })

	Version = "dev"
	if got := AppTitle(); got != "Civ4 Studio (development build)" {
		t.Errorf("got %q", got)
	}
	Version = "v1.2.3"
	if got := AppTitle(); got != "Civ4 Studio v1.2.3" {
		t.Errorf("got %q", got)
	}
}
