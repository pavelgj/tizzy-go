package splotch

import (
	"testing"
)

func TestMenuBarLayout(t *testing.T) {
	mb := NewMenuBar(Style{FillWidth: true}, []Menu{
		{Title: "File"},
		{Title: "Edit"},
	})

	res := Layout(mb, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	if res.W != 80 {
		t.Errorf("Expected width 80, got %d", res.W)
	}
	if res.H != 1 {
		t.Errorf("Expected height 1, got %d", res.H)
	}
}

func TestMenuBarLayoutAutoWidth(t *testing.T) {
	mb := NewMenuBar(Style{}, []Menu{
		{Title: "File"}, // len 4 + 4 = 8
		{Title: "Edit"}, // len 4 + 4 = 8
	})

	res := Layout(mb, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	expectedW := 16
	if res.W != expectedW {
		t.Errorf("Expected width %d, got %d", expectedW, res.W)
	}
}
