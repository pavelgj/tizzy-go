package splotch

import "testing"

func TestFindFocusableIDs(t *testing.T) {
	root := NewBox(Style{},
		NewText(Style{ID: "t1", Focusable: true}, "Text 1"),
		NewBox(Style{ID: "b1", Focusable: true},
			NewText(Style{ID: "t2", Focusable: true}, "Text 2"),
		),
		NewText(Style{ID: "t3"}, "Not Focusable"), // Should not be found
	)

	ids := findFocusableIDs(root)

	expected := []string{"t1", "b1", "t2"}
	if len(ids) != len(expected) {
		t.Fatalf("Expected %d IDs, got %d", len(expected), len(ids))
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("Expected ID at %d to be %s, got %s", i, expected[i], id)
		}
	}
}

func TestNextPrevFocus(t *testing.T) {
	ids := []string{"1", "2", "3"}

	if nextFocus("1", ids) != "2" {
		t.Errorf("Expected next after 1 to be 2")
	}
	if nextFocus("3", ids) != "1" {
		t.Errorf("Expected next after 3 to be 1")
	}
	if nextFocus("", ids) != "1" {
		t.Errorf("Expected next after empty to be 1")
	}

	if prevFocus("2", ids) != "1" {
		t.Errorf("Expected prev before 2 to be 1")
	}
	if prevFocus("1", ids) != "3" {
		t.Errorf("Expected prev before 1 to be 3")
	}
	if prevFocus("", ids) != "3" {
		t.Errorf("Expected prev before empty to be 3")
	}
}
