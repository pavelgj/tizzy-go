package tizzy

import (
	"github.com/gdamore/tcell/v2"
)

// Button is a node that allows user interaction.
type Button struct {
	Style   Style
	Label   string
	OnClick func()
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

	label := "[ " + n.Label + " ]"
	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, label, style)
}

// GetStyle returns the style of the Button node.
func (n *Button) GetStyle() Style {
	return n.Style
}

// IsFocusable indicates that a node can receive focus.
func (n *Button) IsFocusable() bool {
	return n.Style.Focusable
}

// HandleEvent handles mouse and key events for the button.
func (n *Button) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
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
