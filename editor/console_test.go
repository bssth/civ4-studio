package editor

import "testing"

func TestConsoleWrite(t *testing.T) {
	ch := GetConsoleChannel()
	ConsoleWrite("test")

	select {
	case msg := <-ch:
		if msg != "test\n" {
			t.Fatalf("Expected 'test\n', got '%s'", msg)
		}
	default:
		t.Fatal("Expected message on channel")
	}
}
