package splotch

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

// Checkbox is a node that allows user to toggle a boolean value.
type Checkbox struct {
	Style         Style
	Label         string
	Checked       bool
	CheckedChar   string // Default "x"
	UncheckedChar string // Default " "
	OnChange      func(bool)
}

// NewCheckbox creates a new Checkbox node.
func NewCheckbox(ctx *RenderContext, style Style, label string, checked bool, onChange func(bool)) *Checkbox {
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
		ctx.hookIndex++
	}

	return &Checkbox{
		Style:         style,
		Label:         label,
		Checked:       checked,
		CheckedChar:   "x",
		UncheckedChar: " ",
		OnChange:      onChange,
	}
}

// node implements the Node interface.
func (c *Checkbox) node() {}

// Layout calculates the layout for the Checkbox node.
func (n *Checkbox) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	indicatorLen := len(n.CheckedChar)
	if len(n.UncheckedChar) > indicatorLen {
		indicatorLen = len(n.UncheckedChar)
	}
	w := len(n.Label) + indicatorLen + 3
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

// Render draws the Checkbox node to the grid.
func (n *Checkbox) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := false
	if n.Style.ID != "" && n.Style.ID == focusedID {
		focused = true
	}
	
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	if focused {
		style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	}
	
	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}
	
	indicator := n.UncheckedChar
	if n.Checked {
		indicator = n.CheckedChar
	}
	
	text := "[" + indicator + "] " + n.Label
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, text, style)
}

// GetStyle returns the style of the Checkbox node.
func (n *Checkbox) GetStyle() Style {
	return n.Style
}
