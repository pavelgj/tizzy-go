package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutScrollView(t *testing.T) {
	text := NewText(Style{}, "Hello")
	sv := NewScrollView(Style{Width: 10, Height: 5}, text)
	res := Layout(sv, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 10 || res.H != 5 {
		t.Errorf("Expected W=10, H=5, got W=%d, H=%d", res.W, res.H)
	}
	if len(res.Children) != 1 {
		t.Errorf("Expected 1 child, got %d", len(res.Children))
	}
	childRes := res.Children[0]
	if childRes.W != 5 {
		t.Errorf("Expected child W=5, got %d", childRes.W)
	}
}

func TestRenderScrollView(t *testing.T) {
	t1 := NewText(Style{}, "Line 1")
	t2 := NewText(Style{}, "Line 2")
	t3 := NewText(Style{}, "Line 3")
	box := NewBox(Style{FlexDirection: "column"}, t1, t2, t3)
	sv := NewScrollView(Style{Width: 10, Height: 2, ID: "sv"}, box)
	
	componentStates := map[string]any{
		"sv": &ScrollViewState{ScrollOffset: 1},
	}

	layout := Layout(sv, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	s := tcell.NewSimulationScreen("")
	s.Init()
	s.SetSize(10, 2)
	
	grid := NewGrid(10, 2)
	Render(grid, layout, "", componentStates)
	
	for y := 0; y < 2; y++ {
		for x := 0; x < 10; x++ {
			cell := grid.Cells[y][x]
			s.SetContent(x, y, cell.Rune, nil, cell.Style)
		}
	}

	expectedRow0 := "Line 2"
	for i, r := range expectedRow0 {
		mainc, _, _, _ := s.GetContent(i, 0)
		if mainc != r {
			t.Errorf("Row 0: Expected '%c' at %d,0, got '%c'", r, i, mainc)
		}
	}

	expectedRow1 := "Line 3"
	for i, r := range expectedRow1 {
		mainc, _, _, _ := s.GetContent(i, 1)
		if mainc != r {
			t.Errorf("Row 1: Expected '%c' at %d,1, got '%c'", r, i, mainc)
		}
	}
}
