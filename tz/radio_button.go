package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// RadioButton is a node that allows user to select a single option from a set.
type RadioButton struct {
	Style          Style
	Label          string
	Value          string
	Selected       bool
	SelectedChar   string // Default "*"
	UnselectedChar string // Default " "
	OnChange       func(string)
}

// NewRadioButton creates a new RadioButton node.
func NewRadioButton(ctx *RenderContext, style Style, label string, value string, selected bool, onChange func(string)) *RadioButton {
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
		ctx.hookIndex++
	}

	return &RadioButton{
		Style:          style,
		Label:          label,
		Value:          value,
		Selected:       selected,
		SelectedChar:   "*",
		UnselectedChar: " ",
		OnChange:       onChange,
	}
}

// Layout calculates the layout for the RadioButton node.
func (n *RadioButton) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	indicatorLen := len(n.SelectedChar)
	if len(n.UnselectedChar) > indicatorLen {
		indicatorLen = len(n.UnselectedChar)
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

// Render draws the RadioButton node to the grid.
func (n *RadioButton) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := n.Style.ID != "" && n.Style.ID == focusedID

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

	indicator := n.UnselectedChar
	if n.Selected {
		indicator = n.SelectedChar
	}

	text := "(" + indicator + ") " + n.Label
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, text, style)
}

// GetStyle returns the style of the RadioButton node.
func (n *RadioButton) GetStyle() Style {
	return n.Style
}

// IsFocusable indicates that a node can receive focus.
func (n *RadioButton) IsFocusable() bool {
	return n.Style.Focusable
}

// HandleEvent handles mouse and key events for the radio button.
func (n *RadioButton) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	if mev, ok := ev.(MouseEvent); ok {
		if mev.Buttons()&tcell.Button1 != 0 {
			if n.OnChange != nil {
				n.OnChange(n.Value)
			}
			return true
		}
	}
	if key, ok := ev.(*tcell.EventKey); ok {
		if key.Key() == tcell.KeyEnter || key.Rune() == ' ' {
			if n.OnChange != nil {
				n.OnChange(n.Value)
			}
			return true
		}
	}
	return false
}

// DefaultState returns the default state for the radio button.
func (n *RadioButton) DefaultState() any {
	return nil
}
