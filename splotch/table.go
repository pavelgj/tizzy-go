package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// Table is a node that displays tabular data.
type Table struct {
	Style               Style
	Headers             []string
	Rows                [][]string
	ColWidths           []int // Optional, if empty calculate based on content
	CalculatedColWidths []int // Set during layout
}

// NewTable creates a new Table node.
func NewTable(style Style, headers []string, rows [][]string) *Table {
	return &Table{
		Style:   style,
		Headers: headers,
		Rows:    rows,
	}
}


// Layout calculates the layout for the Table.
func (n *Table) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	colWidths := make([]int, len(n.Headers))
	for i, h := range n.Headers {
		colWidths[i] = len(h)
	}

	for _, row := range n.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				if len(cell) > colWidths[i] {
					colWidths[i] = len(cell)
				}
			} else {
				colWidths = append(colWidths, len(cell))
			}
		}
	}

	if len(n.ColWidths) > 0 {
		for i, w := range n.ColWidths {
			if i < len(colWidths) {
				colWidths[i] = w
			}
		}
	}

	n.CalculatedColWidths = colWidths

	w := 0
	for _, cw := range colWidths {
		w += cw
	}
	if len(colWidths) > 1 {
		w += len(colWidths) - 1
	}

	h := len(n.Rows)
	if len(n.Headers) > 0 {
		h += 2
	}

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	if n.Style.FillWidth {
		w = c.MaxW - pad.Left - pad.Right - margin.Left - margin.Right - borderSize
		if w < 0 {
			w = 0
		}
	}

	layoutH := h + pad.Top + pad.Bottom + borderSize
	if n.Style.MaxHeight > 0 && layoutH > n.Style.MaxHeight {
		layoutH = n.Style.MaxHeight
	}

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}
}

// Render draws the Table to the grid.
func (n *Table) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)

	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	pad := n.Style.Padding
	curY := layout.Y + pad.Top + borderOffset

	// Draw headers
	if len(n.Headers) > 0 {
		curX := layout.X + pad.Left + borderOffset
		headerStyle := style.Bold(true)

		for i, h := range n.Headers {
			drawText(grid, curX, curY, h, headerStyle)
			for j := len(h); j < n.CalculatedColWidths[i]; j++ {
				grid.SetContent(curX+j, curY, ' ', headerStyle)
			}
			curX += n.CalculatedColWidths[i] + 1
		}
		curY++

		// Draw separator line
		curX = layout.X + pad.Left + borderOffset
		for i, cw := range n.CalculatedColWidths {
			for j := 0; j < cw; j++ {
				grid.SetContent(curX+j, curY, '-', style)
			}
			if i < len(n.CalculatedColWidths)-1 {
				grid.SetContent(curX+cw, curY, '+', style)
			}
			curX += cw + 1
		}
		curY++
	}

	// Draw rows
	for _, row := range n.Rows {
		curX := layout.X + pad.Left + borderOffset
		for i, cell := range row {
			if i < len(n.CalculatedColWidths) {
				drawText(grid, curX, curY, cell, style)
				for j := len(cell); j < n.CalculatedColWidths[i]; j++ {
					grid.SetContent(curX+j, curY, ' ', style)
				}
				curX += n.CalculatedColWidths[i] + 1
			}
		}
		curY++
	}
}

// GetStyle returns the style of the Table node.
func (n *Table) GetStyle() Style {
	return n.Style
}
