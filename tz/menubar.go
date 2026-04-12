package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// MenuBar is a component that displays a list of menus at the top of the screen.
type MenuBar struct {
	Style Style
	Menus []Menu
	state *MenuBarState // live pointer; same object as componentStates[id]
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
	// Consume a hook slot for stable sibling indices.
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
	}
	ctx.hookIndex++
	style.Focusable = true // MenuBar must be focusable for keyboard access

	// Store state under the component's ID so that n.state and
	// componentStates[id] always refer to the same object (same fix as Dropdown).
	stateObj, ok := ctx.app.componentStates[style.ID]
	if !ok || stateObj == nil {
		stateObj = &MenuBarState{OpenMenuIndex: -1, FocusedItemIndex: -1}
		ctx.app.componentStates[style.ID] = stateObj
	}
	state := stateObj.(*MenuBarState)

	return &MenuBar{
		Style: style,
		Menus: menus,
		state: state,
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

func (m *MenuBar) DefaultState() any {
	return &MenuBarState{OpenMenuIndex: -1, FocusedItemIndex: -1}
}

// Dismiss closes the open menu when another component gains focus.
func (m *MenuBar) Dismiss(state any) {
	if s, ok := state.(*MenuBarState); ok {
		s.OpenMenuIndex = -1
	}
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
			if key.Key() == tcell.KeyRune {
				r := key.Rune()
				for i, menu := range m.Menus {
					if menu.AltRune == r && r != 0 {
						s.OpenMenuIndex = i
						s.FocusedItemIndex = -1
						return true
					}
				}
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

// IsFocusable indicates that a node can receive focus.
func (n *MenuBar) IsFocusable() bool {
	return n.Style.Focusable
}

// Layout calculates the layout for the MenuBar component.
// When a menu is open, the list portal is embedded as a zero-size child so
// that collectPortals can find it during the layout-tree walk.
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

	result := LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}

	// Embed the portal as a zero-size child so collectPortals picks it up.
	if n.state != nil && n.state.OpenMenuIndex >= 0 {
		portalLayout := Layout(n.buildListPortal(), boxX, boxY, c)
		result.Children = append(result.Children, portalLayout)
	}

	return result
}

// buildListPortal creates the Portal that renders the open menu list.
func (m *MenuBar) buildListPortal() *Portal {
	myID := m.Style.ID
	state := m.state

	return &Portal{
		Child: &menuBarListNode{MenuBar: m, State: state},
		PositionFn: func(screenW, screenH int, mainLayout LayoutResult) (x, y, maxW, maxH int) {
			res := findLayoutResultByID(mainLayout, myID)
			if res == nil {
				return 0, 0, screenW, screenH
			}
			borderOffset := 0
			if m.Style.Border {
				borderOffset = 1
			}
			curX := res.X + borderOffset + m.Style.Padding.Left
			menuX := curX
			for i := 0; i < state.OpenMenuIndex; i++ {
				menuX += len(m.Menus[i].Title) + 4
			}
			listW := m.menuListWidth()
			listY := res.Y + res.H
			remaining := screenH - listY
			if remaining < 0 {
				remaining = 0
			}
			return menuX, listY, listW, remaining
		},
		OnOutsideClick: func() {
			state.OpenMenuIndex = -1
		},
	}
}

// menuListWidth returns the width of the open menu's list box (with border).
func (m *MenuBar) menuListWidth() int {
	if m.state == nil || m.state.OpenMenuIndex < 0 || m.state.OpenMenuIndex >= len(m.Menus) {
		return 0
	}
	openMenu := m.Menus[m.state.OpenMenuIndex]
	listW := 0
	for _, item := range openMenu.Items {
		if len(item.Label) > listW {
			listW = len(item.Label)
		}
	}
	return listW + 4 // +2 for padding, +2 for borders
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

	state := n.state

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

// ---------------------------------------------------------------------------
// menuBarListNode — renders the open menu list and handles mouse events on it.
// ---------------------------------------------------------------------------

type menuBarListNode struct {
	Style   Style
	MenuBar *MenuBar
	State   *MenuBarState
}

func (n *menuBarListNode) GetStyle() Style { return n.Style }

func (n *menuBarListNode) Layout(x, y int, c Constraints) LayoutResult {
	w := n.MenuBar.menuListWidth()
	if w <= 0 {
		w = 20
	}
	openMenu := n.MenuBar.Menus[n.State.OpenMenuIndex]
	h := len(openMenu.Items) + 2 // +2 for border
	return LayoutResult{Node: n, X: x, Y: y, W: w, H: h}
}

func (n *menuBarListNode) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	if n.State.OpenMenuIndex < 0 || n.State.OpenMenuIndex >= len(n.MenuBar.Menus) {
		return
	}
	openMenu := n.MenuBar.Menus[n.State.OpenMenuIndex]
	listW := layout.W
	listH := len(openMenu.Items) + 2 // +2 for border
	listX := layout.X
	listY := layout.Y

	style := tcell.StyleDefault.Foreground(n.MenuBar.Style.Color).Background(tcell.ColorBlack)

	// Shadow (right and bottom edges)
	for i := 1; i <= listH; i++ {
		if listY+i < grid.H && listX+listW < grid.W {
			cell := grid.Cells[listY+i][listX+listW]
			grid.SetContent(listX+listW, listY+i, cell.Rune, cell.Style.Background(tcell.ColorDarkGray))
		}
	}
	for j := 1; j <= listW; j++ {
		if listY+listH < grid.H && listX+j < grid.W {
			cell := grid.Cells[listY+listH][listX+j]
			grid.SetContent(listX+j, listY+listH, cell.Rune, cell.Style.Background(tcell.ColorDarkGray))
		}
	}

	// Background fill
	for row := 0; row < listH; row++ {
		for col := 0; col < listW; col++ {
			if listY+row < grid.H && listX+col < grid.W {
				grid.SetContent(listX+col, listY+row, ' ', style)
			}
		}
	}

	drawBorder(grid, listX, listY, listW, listH, "", style)

	for i, item := range openMenu.Items {
		itemStyle := style
		if n.State.FocusedItemIndex == i {
			itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}
		if item.Disabled {
			itemStyle = itemStyle.Foreground(tcell.ColorGray)
		}
		label := " " + item.Label
		for len(label) < listW-2 {
			label += " "
		}
		curX := listX + 1
		for _, r := range label {
			if listY+i+1 < grid.H && curX < grid.W && curX < listX+listW-1 {
				grid.SetContent(curX, listY+i+1, r, itemStyle)
				curX++
			}
		}
	}
}

func (n *menuBarListNode) DefaultState() any { return nil }

func (n *menuBarListNode) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	mouse, ok := ev.(MouseEvent)
	if !ok {
		return false
	}
	if mouse.Buttons()&tcell.Button1 == 0 {
		return false
	}

	if n.State.OpenMenuIndex < 0 || n.State.OpenMenuIndex >= len(n.MenuBar.Menus) {
		return false
	}
	openMenu := n.MenuBar.Menus[n.State.OpenMenuIndex]

	mx, my := mouse.Position()
	listX := ctx.Layout.X
	listY := ctx.Layout.Y
	listW := ctx.Layout.W
	listH := len(openMenu.Items) + 2

	if mx >= listX && mx < listX+listW && my >= listY+1 && my < listY+listH-1 {
		clickedIndex := my - listY - 1
		if clickedIndex >= 0 && clickedIndex < len(openMenu.Items) {
			item := openMenu.Items[clickedIndex]
			if !item.Disabled && item.Action != nil {
				item.Action()
			}
			n.State.OpenMenuIndex = -1
			return true
		}
	}
	return false
}
