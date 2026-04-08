package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// Render draws the layout tree to the grid.
func Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	if r, ok := layout.Node.(Renderable); ok {
		r.Render(grid, layout, focusedID, componentStates)
		return
	}

}

func drawBorder(grid *Grid, x, y, w, h int, title string, style tcell.Style) {
	// Top and bottom borders
	for i := 0; i < w; i++ {
		grid.SetContent(x+i, y, '─', style)
		grid.SetContent(x+i, y+h-1, '─', style)
	}

	// Draw title if present
	if title != "" && w > len(title)+4 {
		grid.SetContent(x+1, y, ' ', style)
		for i, r := range title {
			grid.SetContent(x+2+i, y, r, style)
		}
		grid.SetContent(x+2+len(title), y, ' ', style)
	}

	// Left and right borders
	for i := 0; i < h; i++ {
		grid.SetContent(x, y+i, '│', style)
		grid.SetContent(x+w-1, y+i, '│', style)
	}

	// Corners
	grid.SetContent(x, y, '┌', style)
	grid.SetContent(x+w-1, y, '┐', style)
	grid.SetContent(x, y+h-1, '└', style)
	grid.SetContent(x+w-1, y+h-1, '┘', style)
}

func drawText(grid *Grid, x, y int, text string, style tcell.Style) {
	col := 0
	for _, r := range text {
		grid.SetContent(x+col, y, r, style)
		col++
	}
}
func shiftLayout(res LayoutResult, dx, dy int) LayoutResult {
	res.X += dx
	res.Y += dy
	for i := range res.Children {
		res.Children[i] = shiftLayout(res.Children[i], dx, dy)
	}
	return res
}
