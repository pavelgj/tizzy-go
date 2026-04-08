package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// ProgressBar is a node that displays a visual representation of completion percentage.
type ProgressBar struct {
	Style      Style
	Percent    float64 // 0.0 to 1.0
	FilledChar string
	EmptyChar  string
}

// NewProgressBar creates a new ProgressBar node.
func NewProgressBar(style Style, percent float64) *ProgressBar {
	return &ProgressBar{
		Style:      style,
		Percent:    percent,
		FilledChar: "█",
		EmptyChar:  "░",
	}
}

// node implements the Node interface.
func (p *ProgressBar) node() {}

// Layout calculates the layout for the ProgressBar node.
func (n *ProgressBar) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	w := 20
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

// Render draws the ProgressBar node to the grid.
func (n *ProgressBar) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	
	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
	}
	
	w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
	if w < 0 {
		w = 0
	}
	
	filledW := int(float64(w) * n.Percent)
	if filledW > w {
		filledW = w
	}
	
	str := ""
	for i := 0; i < filledW; i++ {
		str += n.FilledChar
	}
	for i := filledW; i < w; i++ {
		str += n.EmptyChar
	}
	
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, str, style)
}

// GetStyle returns the style of the ProgressBar node.
func (n *ProgressBar) GetStyle() Style {
	return n.Style
}
