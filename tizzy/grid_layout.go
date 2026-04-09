package tizzy

import "github.com/gdamore/tcell/v2"

// GridTrack defines a row or column in the grid.
type GridTrack struct {
	Size   int  // Absolute size in cells, or flex factor
	IsFlex bool // True if Size is a flex factor (like fr)
}

// Fixed creates a fixed-size grid track.
func Fixed(size int) GridTrack {
	return GridTrack{Size: size, IsFlex: false}
}

// Flex creates a flexible grid track.
func Flex(factor int) GridTrack {
	return GridTrack{Size: factor, IsFlex: true}
}

// GridBox is a container that lays out children in a 2D grid.
type GridBox struct {
	Style    Style
	Columns  []GridTrack
	Rows     []GridTrack
	Children []Node
}

// NewGridBox creates a new GridBox node.
func NewGridBox(style Style, columns []GridTrack, rows []GridTrack, children ...Node) *GridBox {
	var validChildren []Node
	for _, c := range children {
		if c != nil {
			validChildren = append(validChildren, c)
		}
	}
	return &GridBox{
		Style:    style,
		Columns:  columns,
		Rows:     rows,
		Children: validChildren,
	}
}

// GetStyle returns the style of the GridBox node.
func (g *GridBox) GetStyle() Style {
	return g.Style
}

// Layout calculates the layout for the GridBox node.
func (g *GridBox) Layout(x, y int, c Constraints) LayoutResult {
	borderSize := 0
	if g.Style.Border {
		borderSize = 1
	}

	pad := g.Style.Padding
	margin := g.Style.Margin

	boxX := x + margin.Left
	boxY := y + margin.Top

	res := LayoutResult{
		Node: g,
		X:    boxX,
		Y:    boxY,
		W:    0,
		H:    0,
	}

	// Available space for tracks
	availW := c.MaxW - (borderSize * 2) - pad.Left - pad.Right
	availH := c.MaxH - (borderSize * 2) - pad.Top - pad.Bottom
	if availW < 0 {
		availW = 0
	}
	if availH < 0 {
		availH = 0
	}

	// Calculate column widths
	colWidths := make([]int, len(g.Columns))
	totalFixedW := 0
	totalFlexW := 0
	for _, col := range g.Columns {
		if col.IsFlex {
			totalFlexW += col.Size
		} else {
			totalFixedW += col.Size
		}
	}

	remainingW := availW - totalFixedW
	if remainingW < 0 {
		remainingW = 0
	}

	for i, col := range g.Columns {
		if col.IsFlex {
			if totalFlexW > 0 {
				colWidths[i] = (remainingW * col.Size) / totalFlexW
			} else {
				colWidths[i] = 0
			}
		} else {
			colWidths[i] = col.Size
		}
	}

	// Calculate row heights
	rowHeights := make([]int, len(g.Rows))
	totalFixedH := 0
	totalFlexH := 0
	for _, row := range g.Rows {
		if row.IsFlex {
			totalFlexH += row.Size
		} else {
			totalFixedH += row.Size
		}
	}

	remainingH := availH - totalFixedH
	if remainingH < 0 {
		remainingH = 0
	}

	for i, row := range g.Rows {
		if row.IsFlex {
			if totalFlexH > 0 {
				rowHeights[i] = (remainingH * row.Size) / totalFlexH
			} else {
				rowHeights[i] = 0
			}
		} else {
			rowHeights[i] = row.Size
		}
	}

	// Calculate track positions
	colPositions := make([]int, len(g.Columns))
	curX := boxX + borderSize + pad.Left
	for i, w := range colWidths {
		colPositions[i] = curX
		curX += w
	}

	rowPositions := make([]int, len(g.Rows))
	curY := boxY + borderSize + pad.Top
	for i, h := range rowHeights {
		rowPositions[i] = curY
		curY += h
	}

	// Layout children
	for _, child := range g.Children {
		style := child.GetStyle()
		row := style.GridRow
		col := style.GridCol
		rowSpan := style.GridRowSpan
		colSpan := style.GridColSpan

		if rowSpan <= 0 {
			rowSpan = 1
		}
		if colSpan <= 0 {
			colSpan = 1
		}

		// Calculate child bounds
		childX := boxX + borderSize + pad.Left
		if col < len(colPositions) {
			childX = colPositions[col]
		}

		childY := boxY + borderSize + pad.Top
		if row < len(rowPositions) {
			childY = rowPositions[row]
		}

		childW := 0
		for i := 0; i < colSpan && col+i < len(colWidths); i++ {
			childW += colWidths[col+i]
		}

		childH := 0
		for i := 0; i < rowSpan && row+i < len(rowHeights); i++ {
			childH += rowHeights[row+i]
		}

		// Apply child margins
		childMargin := style.Margin
		childX += childMargin.Left
		childY += childMargin.Top
		childW -= (childMargin.Left + childMargin.Right)
		childH -= (childMargin.Top + childMargin.Bottom)

		if childW < 0 {
			childW = 0
		}
		if childH < 0 {
			childH = 0
		}

		childConstraints := Constraints{
			MaxW: childW,
			MaxH: childH,
		}

		cRes := Layout(child, childX, childY, childConstraints)
		res.Children = append(res.Children, cRes)
	}

	// Grid size is the sum of tracks or stretched if FillWidth/FillHeight are true!
	totalW := 0
	for _, w := range colWidths {
		totalW += w
	}
	totalH := 0
	for _, h := range rowHeights {
		totalH += h
	}

	res.W = totalW + (borderSize * 2) + pad.Left + pad.Right
	res.H = totalH + (borderSize * 2) + pad.Top + pad.Bottom

	if g.Style.FillWidth && c.MaxW > res.W {
		res.W = c.MaxW
	}
	if g.Style.FillHeight && c.MaxH > res.H {
		res.H = c.MaxH
	}

	return res
}

// Render draws the GridBox node to the grid.
func (g *GridBox) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := g.Style.ID != "" && g.Style.ID == focusedID
	style := tcell.StyleDefault.Foreground(g.Style.Color).Background(g.Style.Background)
	borderStyle := style
	if focused {
		focusColor := tcell.ColorYellow
		if g.Style.FocusColor != tcell.ColorReset {
			focusColor = g.Style.FocusColor
		}
		focusBg := g.Style.Background
		if g.Style.FocusBackground != tcell.ColorReset {
			focusBg = g.Style.FocusBackground
		}
		borderStyle = tcell.StyleDefault.Foreground(focusColor).Background(focusBg)
	}
	if g.Style.Border {
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, g.Style.Title, borderStyle)
	}
	for _, child := range layout.Children {
		Render(grid, child, focusedID, componentStates)
	}
}

// IsFocusable indicates that a node can receive focus.
func (g *GridBox) IsFocusable() bool {
	return g.Style.Focusable
}
