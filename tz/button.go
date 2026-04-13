package tz

import (
	"github.com/gdamore/tcell/v2"
)

// Button is a node that allows user interaction.
type Button struct {
	Style    Style
	Label    string
	OnClick  func()
	Disabled bool
}

// NewButton creates a new Button node.
func NewButton(style Style, label string, onClick func()) *Button {
	return &Button{
		Style:   style,
		Label:   label,
		OnClick: onClick,
	}
}

// Layout calculates the layout for the Button node.
func (n *Button) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	w := len(n.Label) + 4 // "[ " and " ]"
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

// Render draws the Button node to the grid.
func (n *Button) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := n.Style.ID != "" && n.Style.ID == focusedID

	var style tcell.Style
	var borderColor tcell.Color
	if n.Disabled {
		style = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(n.Style.Background).Attributes(tcell.AttrDim)
		borderColor = tcell.ColorGray
	} else if focused {
		focusFg := n.Style.FocusColor
		if focusFg == 0 {
			focusFg = tcell.ColorBlack
		}
		focusBg := n.Style.FocusBackground
		if focusBg == 0 {
			focusBg = tcell.ColorYellow
		}
		style = tcell.StyleDefault.Foreground(focusFg).Background(focusBg).Attributes(n.Style.TextAttrs)
		borderColor = focusBg
	} else {
		style = tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background).Attributes(n.Style.TextAttrs)
		borderColor = tcell.ColorYellow
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", tcell.StyleDefault.Foreground(borderColor))
	}

	label := "[ " + n.Label + " ]"
	innerW := layout.W - borderOffset*2 - n.Style.Padding.Left - n.Style.Padding.Right
	labelOffset := 0
	if innerW > len(label) {
		labelOffset = (innerW - len(label)) / 2
	}
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset+labelOffset, layout.Y+n.Style.Padding.Top+borderOffset, label, style)
}

// WithDisabled sets the Disabled field and returns the button for chaining.
func (n *Button) WithDisabled(disabled bool) *Button {
	n.Disabled = disabled
	return n
}

// GetStyle returns the style of the Button node.
func (n *Button) GetStyle() Style {
	return n.Style
}

// IsFocusable indicates that a node can receive focus.
func (n *Button) IsFocusable() bool {
	return n.Style.Focusable && !n.Disabled
}

// HandleEvent handles mouse and key events for the button.
func (n *Button) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	if n.Disabled {
		return false
	}
	if mev, ok := ev.(MouseEvent); ok {
		if mev.Buttons()&tcell.Button1 != 0 {
			if n.OnClick != nil {
				n.OnClick()
				return true
			}
		}
	}
	if key, ok := ev.(*tcell.EventKey); ok {
		if key.Key() == tcell.KeyEnter || key.Rune() == ' ' {
			if n.OnClick != nil {
				n.OnClick()
				return true
			}
		}
	}
	return false
}

// DefaultState returns the default state for the button.
func (n *Button) DefaultState() any {
	return nil
}
