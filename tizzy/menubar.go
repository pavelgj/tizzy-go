package tizzy

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// MenuBar is a component that displays a list of menus at the top of the screen.
type MenuBar struct {
	Style Style
	Menus []Menu
	State *MenuBarState // Added
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

// NewMenuBar creates a new MenuBar component.
func NewMenuBar(ctx *RenderContext, style Style, menus []Menu) *MenuBar {
	stateObj, _ := ctx.UseState(&MenuBarState{OpenMenuIndex: -1})

	// Derive hook ID and set it on style
	id := fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	style.ID = id
	style.Focusable = true // MenuBar must be focusable for keyboard access

	return &MenuBar{
		Style: style,
		Menus: menus,
		State: stateObj.(*MenuBarState),
	}
}

// MenuBarState stores the interactive state of a MenuBar.
type MenuBarState struct {
	OpenMenuIndex    int // -1 if no menu is open
	FocusedItemIndex int
	HoverItemIndex   int
}

// GetStyle returns the style of the MenuBar node.
func (m *MenuBar) DefaultState() any {
	return &MenuBarState{OpenMenuIndex: -1, FocusedItemIndex: -1}
}

func (m *MenuBar) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s := state.(*MenuBarState)
	if key, ok := ev.(*tcell.EventKey); ok {
		if key.Key() == tcell.KeyTab || key.Key() == tcell.KeyBacktab {
			s.OpenMenuIndex = -1
			return false // Let app.go handle focus change
		}
		if s.OpenMenuIndex == -1 {
			if key.Key() == tcell.KeyEnter || key.Key() == tcell.KeyDown {
				s.OpenMenuIndex = 0
				s.FocusedItemIndex = -1
				return true
			}
		} else {
			openMenu := m.Menus[s.OpenMenuIndex]
			if key.Key() == tcell.KeyDown {
				s.FocusedItemIndex++
				if s.FocusedItemIndex >= len(openMenu.Items) {
					s.FocusedItemIndex = 0
				}
				return true
			} else if key.Key() == tcell.KeyUp {
				s.FocusedItemIndex--
				if s.FocusedItemIndex < 0 {
					s.FocusedItemIndex = len(openMenu.Items) - 1
				}
				return true
			} else if key.Key() == tcell.KeyRight {
				s.OpenMenuIndex++
				if s.OpenMenuIndex >= len(m.Menus) {
					s.OpenMenuIndex = 0
				}
				s.FocusedItemIndex = -1
				return true
			} else if key.Key() == tcell.KeyLeft {
				s.OpenMenuIndex--
				if s.OpenMenuIndex < 0 {
					s.OpenMenuIndex = len(m.Menus) - 1
				}
				s.FocusedItemIndex = -1
				return true
			} else if key.Key() == tcell.KeyEnter {
				if s.FocusedItemIndex >= 0 && s.FocusedItemIndex < len(openMenu.Items) {
					item := openMenu.Items[s.FocusedItemIndex]
					if !item.Disabled && item.Action != nil {
						item.Action()
					}
					s.OpenMenuIndex = -1
					return true
				}
			} else if key.Key() == tcell.KeyEscape {
				s.OpenMenuIndex = -1
				return true
			}
		}
	} else if mouse, ok := ev.(*tcell.EventMouse); ok {
		mx, my := mouse.Position()
		borderOffset := 0
		if m.Style.Border {
			borderOffset = 1
		}
		curX := ctx.Layout.X + borderOffset + m.Style.Padding.Left
		curY := ctx.Layout.Y + borderOffset + m.Style.Padding.Top

		if my == curY {
			for i, menu := range m.Menus {
				titleLen := len(menu.Title) + 2 // " " + title + " "
				if mx >= curX && mx < curX+titleLen {
					s.OpenMenuIndex = i
					s.FocusedItemIndex = -1
					return true
				}
				curX += titleLen + 2 // +2 for spacing
			}
		}
	}
	return false
}

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
	if n.Style.ID != "" && n.Style.ID == focusedID {
		style = style.Background(tcell.ColorNavy)
	}
	borderStyle := style
	if n.Style.ID != "" && n.Style.ID == focusedID {
		borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	curX := layout.X + borderOffset + n.Style.Padding.Left
	curY := layout.Y + borderOffset + n.Style.Padding.Top

	state := n.State

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

// IsFocusable indicates that a node can receive focus.
func (n *MenuBar) IsFocusable() bool {
	return n.Style.Focusable
}
