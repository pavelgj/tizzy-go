package tz

import "github.com/gdamore/tcell/v2"

// Cell represents a single character on the screen with its style.
type Cell struct {
	Rune  rune
	Style tcell.Style
}

// Grid represents a 2D array of cells.
type Grid struct {
	W     int
	H     int
	Cells [][]Cell
}

// NewGrid creates a new Grid with the given dimensions.
func NewGrid(w, h int) *Grid {
	cells := make([][]Cell, h)
	for i := range cells {
		cells[i] = make([]Cell, w)
		for j := range cells[i] {
			cells[i][j] = Cell{Rune: ' ', Style: tcell.StyleDefault}
		}
	}
	return &Grid{W: w, H: h, Cells: cells}
}

// SetContent sets the content of a cell at (x, y).
func (g *Grid) SetContent(x, y int, r rune, style tcell.Style) {
	if x >= 0 && x < g.W && y >= 0 && y < g.H {
		g.Cells[y][x] = Cell{Rune: r, Style: style}
	}
}
