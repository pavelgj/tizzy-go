package tz

import (
	"github.com/gdamore/tcell/v2"
)

// Table is a node that displays tabular data.
type Table struct {
	Style               Style
	Headers             []string
	Rows                [][]string
	ColWidths           []int       // Optional, if empty calculate based on content
	ColAligns           []string    // Optional, "left", "center", "right" per column
	StripeBackground    tcell.Color // Optional, alternating row background color
	Dividers            bool        // Optional, draw vertical column dividers
	CalculatedColWidths []int       // Set during layout
}

// NewTable creates a new Table node.
func NewTable(style Style, headers []string, rows [][]string) *Table {
	return &Table{
		Style:   style,
		Headers: headers,
		Rows:    rows,
	}
}

// runeLen returns the number of runes (terminal columns) in a string.
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// Layout calculates the layout for the Table.
func (n *Table) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	colWidths := make([]int, len(n.Headers))
	for i, h := range n.Headers {
		colWidths[i] = runeLen(h)
	}

	for _, row := range n.Rows {
		for i, cell := range row {
			cw := runeLen(cell)
			if i < len(colWidths) {
				if cw > colWidths[i] {
					colWidths[i] = cw
				}
			} else {
				colWidths = append(colWidths, cw)
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

	sepWidth := 0
	if len(colWidths) > 1 {
		sepWidth = len(colWidths) - 1
	}

	naturalW := sepWidth
	for _, cw := range colWidths {
		naturalW += cw
	}

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	w := naturalW
	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	if n.Style.FillWidth {
		availW := c.MaxW - pad.Left - pad.Right - margin.Left - margin.Right - borderSize
		if availW < 0 {
			availW = 0
		}
		// Distribute extra width proportionally among columns.
		colsNatural := naturalW - sepWidth
		colsAvail := availW - sepWidth
		if colsAvail > colsNatural && colsNatural > 0 {
			extra := colsAvail - colsNatural
			distributed := 0
			for i := range colWidths {
				if i < len(colWidths)-1 {
					add := extra * colWidths[i] / colsNatural
					colWidths[i] += add
					distributed += add
				} else {
					colWidths[i] += extra - distributed
				}
			}
		}
		w = availW
	}

	n.CalculatedColWidths = colWidths

	h := len(n.Rows)
	if len(n.Headers) > 0 {
		h += 2
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

// renderCell draws a single table cell with alignment, padding, and truncation.
func renderCell(grid *Grid, x, y int, text string, colWidth int, align string, style tcell.Style) {
	runes := []rune(text)
	if len(runes) > colWidth {
		if colWidth > 1 {
			runes = append(runes[:colWidth-1], '…')
		} else {
			runes = runes[:colWidth]
		}
	}

	textWidth := len(runes)
	extra := colWidth - textWidth
	if extra < 0 {
		extra = 0
	}

	var leftPad int
	switch align {
	case "right":
		leftPad = extra
	case "center":
		leftPad = extra / 2
	default: // "left"
		leftPad = 0
	}
	rightPad := extra - leftPad

	for i := 0; i < leftPad; i++ {
		grid.SetContent(x+i, y, ' ', style)
	}
	for i, r := range runes {
		grid.SetContent(x+leftPad+i, y, r, style)
	}
	for i := 0; i < rightPad; i++ {
		grid.SetContent(x+leftPad+textWidth+i, y, ' ', style)
	}
}

// Render draws the Table to the grid.
func (n *Table) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)

	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, n.Style.Title, borderStyle)
	}

	pad := n.Style.Padding
	contentX := layout.X + pad.Left + borderOffset
	contentEndX := layout.X + layout.W - pad.Right - borderOffset
	curY := layout.Y + pad.Top + borderOffset

	colAlign := func(i int) string {
		if i < len(n.ColAligns) && n.ColAligns[i] != "" {
			return n.ColAligns[i]
		}
		return "left"
	}

	divChar := ' '
	if n.Dividers {
		divChar = '│'
	}

	// Draw headers.
	if len(n.Headers) > 0 {
		headerStyle := style.Bold(true)
		curX := contentX
		for i, h := range n.Headers {
			renderCell(grid, curX, curY, h, n.CalculatedColWidths[i], colAlign(i), headerStyle)
			curX += n.CalculatedColWidths[i]
			if i < len(n.CalculatedColWidths)-1 {
				grid.SetContent(curX, curY, divChar, headerStyle)
				curX++
			}
		}
		// Fill remainder of header row.
		for ; curX < contentEndX; curX++ {
			grid.SetContent(curX, curY, ' ', headerStyle)
		}
		curY++

		// Draw separator line.
		curX = contentX
		for i, cw := range n.CalculatedColWidths {
			for j := 0; j < cw; j++ {
				grid.SetContent(curX+j, curY, '─', style)
			}
			curX += cw
			if i < len(n.CalculatedColWidths)-1 {
				sep := '─'
				if n.Dividers {
					sep = '┼'
				}
				grid.SetContent(curX, curY, sep, style)
				curX++
			}
		}
		// Fill remainder of separator.
		for ; curX < contentEndX; curX++ {
			grid.SetContent(curX, curY, '─', style)
		}
		curY++
	}

	// Draw rows.
	for rowIdx, row := range n.Rows {
		rowStyle := style
		if n.StripeBackground != 0 && rowIdx%2 == 1 {
			rowStyle = rowStyle.Background(n.StripeBackground)
		}

		curX := contentX
		for i, cell := range row {
			if i >= len(n.CalculatedColWidths) {
				break
			}
			renderCell(grid, curX, curY, cell, n.CalculatedColWidths[i], colAlign(i), rowStyle)
			curX += n.CalculatedColWidths[i]
			if i < len(n.CalculatedColWidths)-1 {
				grid.SetContent(curX, curY, divChar, rowStyle)
				curX++
			}
		}
		// Fill remainder of row (needed for striped backgrounds and FillWidth).
		for ; curX < contentEndX; curX++ {
			grid.SetContent(curX, curY, ' ', rowStyle)
		}
		curY++
	}
}

// GetStyle returns the style of the Table node.
func (n *Table) GetStyle() Style {
	return n.Style
}
