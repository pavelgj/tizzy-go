package splotch

import (
	"time"
	"github.com/gdamore/tcell/v2"
)

// Spinner is a node that displays a loading animation.
type Spinner struct {
	Style    Style
	Frames   []string
	Interval time.Duration
}

// NewSpinner creates a new Spinner node.
func NewSpinner(style Style) *Spinner {
	return &Spinner{
		Style:    style,
		Frames:   []string{"|", "/", "-", "\\"},
		Interval: 100 * time.Millisecond,
	}
}

// node implements the Node interface.
func (s *Spinner) node() {}

// Layout calculates the layout for the Spinner node.
func (n *Spinner) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	w := 1
	h := 1

	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
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

// Render draws the Spinner node to the grid.
func (n *Spinner) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	
	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
	}
	
	now := time.Now()
	frameIdx := int((now.UnixNano() / int64(n.Interval)) % int64(len(n.Frames)))
	val := n.Frames[frameIdx]
	
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, val, style)
}

// GetStyle returns the style of the Spinner node.
func (n *Spinner) GetStyle() Style {
	return n.Style
}
