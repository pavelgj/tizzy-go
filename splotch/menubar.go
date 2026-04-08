package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// MenuBar is a component that displays a list of menus at the top of the screen.
type MenuBar struct {
	Style Style
	Menus []Menu
}

// Menu represents a single menu in the MenuBar.
type Menu struct {
	Title   string
	Items   []MenuItem
	AltRune rune
}

// MenuItem represents an item in a Menu.
type MenuItem struct {
	Label    string
	Action   func()
	Disabled bool
	Shortcut string
}

func (m *MenuBar) node() {}

// NewMenuBar creates a new MenuBar component.
func NewMenuBar(style Style, menus []Menu) *MenuBar {
	return &MenuBar{
		Style: style,
		Menus: menus,
	}
}

// MenuBarState stores the interactive state of a MenuBar.
type MenuBarState struct {
	OpenMenuIndex    int // -1 if no menu is open
	FocusedItemIndex int
	HoverItemIndex   int
}

// GetStyle returns the style of the MenuBar node.
func (m *MenuBar) GetStyle() Style {
	return m.Style
}

// Layout calculates the layout for the MenuBar component.
func (n *MenuBar) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	totalW := 0
	for _, menu := range n.Menus {
		totalW += len(menu.Title) + 4
	}

	w := totalW
	if n.Style.Width > 0 {
		w = n.Style.Width
	}
	if n.Style.FillWidth && c.MaxW > 0 {
		w = c.MaxW - margin.Left - margin.Right
	}

	h := 1
	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	layoutH := h + pad.Top + pad.Bottom + borderSize

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}
}

// Render draws the MenuBar component to the grid.
func (n *MenuBar) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	borderStyle := style
	if n.Style.ID != "" && n.Style.ID == focusedID {
		borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
	}

	curX := layout.X + borderOffset + n.Style.Padding.Left
	curY := layout.Y + borderOffset + n.Style.Padding.Top

	var state *MenuBarState
	if n.Style.ID != "" {
		if s, ok := componentStates[n.Style.ID].(*MenuBarState); ok {
			state = s
		}
	}

	for i, menu := range n.Menus {
		title := " " + menu.Title + " "
		titleStyle := style

		if state != nil && state.OpenMenuIndex == i {
			titleStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}

		drawText(grid, curX, curY, title, titleStyle)
		curX += len(title) + 2
	}

	for x := curX; x < layout.X+layout.W-borderOffset; x++ {
		grid.SetContent(x, curY, ' ', style)
	}
}
