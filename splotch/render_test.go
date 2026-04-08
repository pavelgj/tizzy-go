package splotch

import (
	"github.com/gdamore/tcell/v2"
)

func renderToScreen(s tcell.SimulationScreen, layout LayoutResult, focusedID string, componentStates map[string]any) {
	w, h := s.Size()
	grid := NewGrid(w, h)
	Render(grid, layout, focusedID, componentStates)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			cell := grid.Cells[y][x]
			s.SetContent(x, y, cell.Rune, nil, cell.Style)
		}
	}
}
