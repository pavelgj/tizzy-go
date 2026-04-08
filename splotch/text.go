package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// Text is a leaf node that displays text.
type Text struct {
	Style   Style
	Content string
}

// NewText creates a new Text node.
func NewText(style Style, content string) *Text {
	return &Text{Style: style, Content: content}
}

// Layout calculates the layout for the Text node.
func (n *Text) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin

	// Position of the border box
	boxX := x + margin.Left
	boxY := y + margin.Top

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    len(n.Content) + pad.Left + pad.Right,
		H:    1 + pad.Top + pad.Bottom,
	}
}

// Render draws the Text node to the grid.
func (n *Text) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := false
	if n.Style.ID != "" && n.Style.ID == focusedID {
		focused = true
	}
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	if focused {
		style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	}
	drawText(grid, layout.X+n.Style.Padding.Left, layout.Y+n.Style.Padding.Top, n.Content, style)
}

// GetStyle returns the style of the Text node.
func (n *Text) GetStyle() Style {
	return n.Style
}
