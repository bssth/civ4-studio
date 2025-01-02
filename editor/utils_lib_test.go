package editor

import "testing"

func TestToInt(t *testing.T) {
	if ToInt("1") != 1 {
		t.Error("ToInt failed")
	}

	if ToInt("0") != 0 {
		t.Error("ToInt failed")
	}
}

func TestToUint(t *testing.T) {
	if ToUint("1") != 1 {
		t.Error("ToUint failed")
	}

	if ToUint("0") != 0 {
		t.Error("ToUint failed")
	}
}

func TestIsInSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if !IsInSlice(slice, "a") {
		t.Error("IsInSlice failed")
	}

	if IsInSlice(slice, "d") {
		t.Error("IsInSlice failed")
	}
}

func TestSwitchInSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if len(SwitchInSlice(true, slice, "d")) != 4 {
		t.Error("SwitchInSlice failed")
	}

	if len(SwitchInSlice(false, slice, "b")) != 2 {
		t.Error("SwitchInSlice failed")
	}
}

func TestAddToSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if len(AddToSlice(slice, "d")) != 4 {
		t.Error("AddToSlice failed")
	}

	if len(AddToSlice(slice, "b")) != 3 {
		t.Error("AddToSlice failed")
	}
}

func TestRemoveFromSlice(t *testing.T) {
	slice := []string{"a", "b", "c"}

	if len(RemoveFromSlice(slice, "b")) != 2 {
		t.Error("RemoveFromSlice failed")
	}

	if len(RemoveFromSlice(slice, "d")) != 3 {
		t.Error("RemoveFromSlice failed")
	}
}

func TestSortKeys(t *testing.T) {
	dict := map[string]int{"b": 2, "a": 1, "c": 3}
	keys := SortKeys(dict)

	if keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Error("SortKeys failed")
	}
}

func TestBoolToInt(t *testing.T) {
	if BoolToInt(true) != 1 {
		t.Error("BoolToInt failed")
	}

	if BoolToInt(false) != 0 {
		t.Error("BoolToInt failed")
	}
}
