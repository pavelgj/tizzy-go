package tz

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

	expectedRow0 := "ID Name "
	for i, r := range expectedRow0 {
		str, _, _ := s.Get(i, 0)
		if str != string(r) {
			t.Errorf("Row 0: Expected '%c' at %d,0, got '%s'", r, i, str)
		}
	}

	expectedRow1 := "────────"
	for i, r := range []rune(expectedRow1) {
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

func TestTableTruncation(t *testing.T) {
	headers := []string{"Name"}
	rows := [][]string{
		{"Very long content here"},
	}
	table := NewTable(Style{}, headers, rows)
	table.ColWidths = []int{8}
	layout := Layout(table, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	// Cell should be truncated to 8 chars: 7 chars + ellipsis
	expected := "Very lo…"
	for i, r := range []rune(expected) {
		str, _, _ := s.Get(i, 2)
		if str != string(r) {
			t.Errorf("Truncation: Expected '%c' at %d, got '%s'", r, i, str)
		}
	}
}

func TestTableAlignment(t *testing.T) {
	headers := []string{"L", "C", "R"}
	rows := [][]string{
		{"x", "x", "x"},
	}
	table := NewTable(Style{}, headers, rows)
	table.ColWidths = []int{5, 5, 5}
	table.ColAligns = []string{"left", "center", "right"}
	layout := Layout(table, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(40, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	// left:   "x    " (x at position 0)
	// center: "  x  " (x at position 2)
	// right:  "    x" (x at position 4)
	checkCell := func(x, y int, expected rune) {
		t.Helper()
		str, _, _ := s.Get(x, y)
		if str != string(expected) {
			t.Errorf("Alignment: Expected '%c' at %d,%d, got '%s'", expected, x, y, str)
		}
	}

	// Data row (row index 2 after header + separator)
	checkCell(0, 2, 'x') // left col: x at offset 0
	checkCell(8, 2, 'x') // center col: x at offset 2 within col (6 + 2)
	checkCell(16, 2, 'x') // right col: x at offset 4 within col (12 + 4)
}

func TestTableFillWidthDistribution(t *testing.T) {
	headers := []string{"A", "B"}
	rows := [][]string{{"1", "2"}}
	table := NewTable(Style{FillWidth: true}, headers, rows)
	layout := Layout(table, 0, 0, Constraints{MaxW: 20, MaxH: 100})

	// Total width should fill MaxW
	if layout.W != 20 {
		t.Errorf("Expected W=20, got W=%d", layout.W)
	}

	// Both columns should have received extra space
	total := 0
	for _, cw := range table.CalculatedColWidths {
		total += cw
	}
	// total + 1 separator = 20 (no border, no padding)
	if total+1 != 20 {
		t.Errorf("Expected column widths to sum to 19, got %d", total)
	}
}
