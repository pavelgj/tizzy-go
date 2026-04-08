package tizzy

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutTable(t *testing.T) {
	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"1", "Alice"},
		{"2", "Bob"},
	}
	table := NewTable(Style{}, headers, rows)
	res := Layout(table, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 8 || res.H != 4 {
		t.Errorf("Expected W=8, H=4, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderTable(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	headers := []string{"ID", "Name"}
	rows := [][]string{
		{"1", "Alice"},
	}
	table := NewTable(Style{}, headers, rows)
	layout := Layout(table, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	expectedRow0 := "ID Name"
	for i, r := range expectedRow0 {
		str, _, _ := s.Get(i, 0)
		if str != string(r) {
			t.Errorf("Row 0: Expected '%c' at %d,0, got '%s'", r, i, str)
		}
	}

	expectedRow1 := "--+-----"
	for i, r := range expectedRow1 {
		str, _, _ := s.Get(i, 1)
		if str != string(r) {
			t.Errorf("Row 1: Expected '%c' at %d,1, got '%s'", r, i, str)
		}
	}

	expectedRow2 := "1  Alice"
	for i, r := range expectedRow2 {
		str, _, _ := s.Get(i, 2)
		if str != string(r) {
			t.Errorf("Row 2: Expected '%c' at %d,2, got '%s'", r, i, str)
		}
	}
}
