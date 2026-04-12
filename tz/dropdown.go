package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Dropdown is a component that allows selecting an option from a list.
type Dropdown struct {
	Style         Style
	Options       []string
	SelectedIndex int
	OnChange      func(int)
	MaxListHeight int
	Placeholder   string
	state         *DropdownState // live pointer; set by NewDropdown
}

// NewDropdown creates a new Dropdown component.
func NewDropdown(ctx *RenderContext, style Style, options []string, selectedIndex int, onChange func(int), maxListHeight ...int) *Dropdown {
	// Consume a hook slot so that sibling components keep stable hook indices.
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
	}
	ctx.hookIndex++

	// Store state under the component's ID (not the hook index) so that
	// n.state and componentStates[id] always refer to the same object.
	// Layout() reads n.state to decide whether to embed the portal child,
	// while event handlers look up componentStates[id] — they must agree.
	stateObj, ok := ctx.app.componentStates[style.ID]
	if !ok || stateObj == nil {
		stateObj = &DropdownState{Open: false, FocusedIndex: selectedIndex}
		ctx.app.componentStates[style.ID] = stateObj
	}
	state := stateObj.(*DropdownState)

	mlh := 0
	if len(maxListHeight) > 0 {
		mlh = maxListHeight[0]
	}
	return &Dropdown{
		Style:         style,
		Options:       options,
		SelectedIndex: selectedIndex,
		OnChange:      onChange,
		MaxListHeight: mlh,
		state:         state,
	}
}

// DropdownState stores the interactive state of a Dropdown.
type DropdownState struct {
	Open         bool
	FocusedIndex int
	ScrollOffset int
	OpenAbove    bool
}

// GetStyle returns the style of the Dropdown node.
func (d *Dropdown) GetStyle() Style {
	return d.Style
}

// Layout calculates the layout for the Dropdown component.
// When the dropdown is open, the portal is embedded as a zero-size child so
// that collectPortals can find it during the layout-tree walk.
func (n *Dropdown) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	maxW := len(n.Placeholder)
	for _, opt := range n.Options {
		if len(opt) > maxW {
			maxW = len(opt)
		}
	}

	w := maxW + 6
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

	result := LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}

	// Embed the portal as a zero-size child so collectPortals picks it up.
	if n.state != nil && n.state.Open {
		portalLayout := Layout(n.buildListPortal(), boxX, boxY, c)
		result.Children = append(result.Children, portalLayout)
	}

	return result
}

// buildListPortal creates the Portal that renders the open dropdown list.
func (d *Dropdown) buildListPortal() *Portal {
	myID := d.Style.ID
	state := d.state

	return &Portal{
		Child: &dropdownListNode{Dropdown: d, State: state},
		PositionFn: func(screenW, screenH int, mainLayout LayoutResult) (x, y, maxW, maxH int) {
			res := findLayoutResultByID(mainLayout, myID)
			if res == nil {
				return 0, 0, screenW, screenH
			}
			popupH := d.visibleItemCount() + 2 // +2 for border
			spaceBelow := screenH - (res.Y + res.H)
			if spaceBelow < popupH && res.Y >= popupH {
				state.OpenAbove = true
			} else {
				state.OpenAbove = false
			}
			listY := res.Y + res.H
			if state.OpenAbove {
				listY = res.Y - popupH
			}
			remaining := screenH - listY
			if remaining < 0 {
				remaining = 0
			}
			return res.X, listY, res.W, remaining
		},
		OnOutsideClick: func() {
			state.Open = false
		},
	}
}

func (d *Dropdown) visibleItemCount() int {
	maxH := d.MaxListHeight
	if maxH <= 0 {
		maxH = 5
	}
	if maxH > len(d.Options) {
		maxH = len(d.Options)
	}
	return maxH
}

// Render draws the Dropdown component to the grid.
func (n *Dropdown) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	borderOffset := 0
	focused := n.Style.ID != "" && n.Style.ID == focusedID

	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	borderStyle := style
	focusColor := tcell.ColorYellow
	if n.Style.FocusColor != tcell.ColorDefault {
		focusColor = n.Style.FocusColor
	}
	if focused {
		borderStyle = tcell.StyleDefault.Foreground(focusColor).Background(n.Style.Background)
	}
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	pad := n.Style.Padding
	curX := layout.X + pad.Left + borderOffset
	curY := layout.Y + pad.Top + borderOffset

	hasSelection := n.SelectedIndex >= 0 && n.SelectedIndex < len(n.Options)
	selectedText := n.Placeholder
	if hasSelection {
		selectedText = n.Options[n.SelectedIndex]
	}

	// Brackets and chevron match the border color when focused.
	decorStyle := style
	if focused {
		decorStyle = borderStyle
	}

	// Placeholder is dimmed to distinguish it from an actual selection.
	contentStyle := style
	if !hasSelection {
		contentStyle = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(n.Style.Background)
	}

	// Chevron flips to reflect open/closed state.
	chevron := "▼"
	if n.state != nil && n.state.Open {
		chevron = "▲"
	}

	drawText(grid, curX, curY, "[ ", decorStyle)
	curX += 2
	drawText(grid, curX, curY, selectedText, contentStyle)
	curX += len(selectedText)

	remaining := layout.W - borderOffset*2 - pad.Left - pad.Right - 2 - len(selectedText) - 4
	if remaining > 0 {
		for i := 0; i < remaining; i++ {
			grid.SetContent(curX+i, curY, ' ', style)
		}
		curX += remaining
	}

	drawText(grid, curX, curY, " "+chevron+" ]", decorStyle)
}

func (d *Dropdown) DefaultState() any {
	return &DropdownState{}
}

// Dismiss closes the dropdown when another component gains focus.
func (d *Dropdown) Dismiss(state any) {
	if s, ok := state.(*DropdownState); ok {
		s.Open = false
	}
}

// IsFocusable indicates that a node can receive focus.
func (d *Dropdown) IsFocusable() bool {
	return d.Style.Focusable
}

func (d *Dropdown) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s, ok := state.(*DropdownState)
	if !ok {
		return false
	}

	if mev, ok := ev.(MouseEvent); ok {
		if mev.Buttons()&tcell.Button1 != 0 {
			s.Open = !s.Open
			return true
		}
	}

	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return false
	}

	dirty := false
	n := d.visibleItemCount()

	if key.Key() == tcell.KeyEnter {
		if s.Open {
			d.SelectedIndex = s.FocusedIndex
			if d.OnChange != nil {
				d.OnChange(s.FocusedIndex)
			}
			s.Open = false
		} else {
			s.Open = true
		}
		dirty = true
	} else if key.Key() == tcell.KeyEscape {
		if s.Open {
			s.Open = false
			dirty = true
		}
	} else if key.Key() == tcell.KeyUp {
		if s.Open {
			if s.FocusedIndex > 0 {
				s.FocusedIndex--
			}
			if s.FocusedIndex < s.ScrollOffset {
				s.ScrollOffset = s.FocusedIndex
			}
			if s.FocusedIndex >= s.ScrollOffset+n {
				s.ScrollOffset = s.FocusedIndex - n + 1
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyDown {
		if s.Open {
			if s.FocusedIndex < len(d.Options)-1 {
				s.FocusedIndex++
			}
			if s.FocusedIndex >= s.ScrollOffset+n {
				s.ScrollOffset = s.FocusedIndex - n + 1
			}
			if s.FocusedIndex < s.ScrollOffset {
				s.ScrollOffset = 0
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyPgUp {
		if s.Open {
			s.FocusedIndex -= n
			if s.FocusedIndex < 0 {
				s.FocusedIndex = 0
			}
			s.ScrollOffset -= n
			if s.ScrollOffset < 0 {
				s.ScrollOffset = 0
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyPgDn {
		if s.Open {
			s.FocusedIndex += n
			if s.FocusedIndex >= len(d.Options) {
				s.FocusedIndex = len(d.Options) - 1
			}
			s.ScrollOffset += n
			if s.ScrollOffset+n > len(d.Options) {
				s.ScrollOffset = len(d.Options) - n
				if s.ScrollOffset < 0 {
					s.ScrollOffset = 0
				}
			}
			dirty = true
		}
	}

	return dirty
}

// ---------------------------------------------------------------------------
// dropdownListNode — renders the open list and handles all mouse events on it.
// ---------------------------------------------------------------------------

type dropdownListNode struct {
	Style    Style
	Dropdown *Dropdown
	State    *DropdownState
}

func (n *dropdownListNode) GetStyle() Style { return n.Style }

func (n *dropdownListNode) Layout(x, y int, c Constraints) LayoutResult {
	w := c.MaxW
	if w <= 0 {
		w = 20
	}
	h := n.Dropdown.visibleItemCount() + 2 // +2 for border
	return LayoutResult{Node: n, X: x, Y: y, W: w, H: h}
}

func (n *dropdownListNode) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	listH := n.Dropdown.visibleItemCount()
	listW := layout.W
	listX := layout.X
	listY := layout.Y
	popupH := listH + 2

	style := tcell.StyleDefault.Foreground(n.Dropdown.Style.Color).Background(tcell.ColorBlack)

	// Shadow (right and bottom edges)
	for i := 1; i <= popupH; i++ {
		if listY+i < grid.H && listX+listW < grid.W {
			cell := grid.Cells[listY+i][listX+listW]
			grid.SetContent(listX+listW, listY+i, cell.Rune, cell.Style.Background(tcell.ColorDarkGray))
		}
	}
	for j := 1; j <= listW; j++ {
		if listY+popupH < grid.H && listX+j < grid.W {
			cell := grid.Cells[listY+popupH][listX+j]
			grid.SetContent(listX+j, listY+popupH, cell.Rune, cell.Style.Background(tcell.ColorDarkGray))
		}
	}

	// Background fill
	for row := 0; row < popupH; row++ {
		for col := 0; col < listW; col++ {
			if listY+row < grid.H && listX+col < grid.W {
				grid.SetContent(listX+col, listY+row, ' ', style)
			}
		}
	}

	drawBorder(grid, listX, listY, listW, popupH, "", style)

	// Scroll indicators: show ▲/▼ in the border when items are hidden above/below.
	if n.State.ScrollOffset > 0 && listW >= 3 {
		grid.SetContent(listX+listW-2, listY, '▲', style)
	}
	if n.State.ScrollOffset+listH < len(n.Dropdown.Options) && listW >= 3 {
		grid.SetContent(listX+listW-2, listY+popupH-1, '▼', style)
	}

	for i := 0; i < listH; i++ {
		optIdx := i + n.State.ScrollOffset
		if optIdx >= len(n.Dropdown.Options) {
			break
		}
		opt := n.Dropdown.Options[optIdx]
		optStyle := style
		if optIdx == n.State.FocusedIndex {
			optStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}
		// Mark the currently selected item with a checkmark.
		prefix := ' '
		if optIdx == n.Dropdown.SelectedIndex {
			prefix = '✓'
		}
		labelRunes := append([]rune{prefix}, []rune(opt)...)
		for len(labelRunes) < listW-2 {
			labelRunes = append(labelRunes, ' ')
		}
		curX := listX + 1
		for _, r := range labelRunes {
			if listY+1+i < grid.H && curX < grid.W && curX < listX+listW-1 {
				grid.SetContent(curX, listY+1+i, r, optStyle)
				curX++
			}
		}
	}
}

func (n *dropdownListNode) DefaultState() any { return nil }

func (n *dropdownListNode) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	mouse, ok := ev.(MouseEvent)
	if !ok {
		return false
	}

	mx, my := mouse.Position()
	listH := n.Dropdown.visibleItemCount()
	listX := ctx.Layout.X
	listY := ctx.Layout.Y
	listW := ctx.Layout.W

	if mouse.Buttons()&tcell.Button1 != 0 {
		if mx >= listX && mx < listX+listW && my >= listY+1 && my < listY+1+listH {
			clickedIndex := my - listY - 1 + n.State.ScrollOffset
			if clickedIndex >= 0 && clickedIndex < len(n.Dropdown.Options) {
				n.Dropdown.SelectedIndex = clickedIndex
				if n.Dropdown.OnChange != nil {
					n.Dropdown.OnChange(clickedIndex)
				}
				n.State.Open = false
				return true
			}
		}
		return false
	}

	if mouse.Buttons()&tcell.WheelUp != 0 {
		if mx >= listX && mx < listX+listW && my >= listY+1 && my < listY+1+listH {
			n.State.ScrollOffset--
			if n.State.ScrollOffset < 0 {
				n.State.ScrollOffset = 0
			}
			return true
		}
	} else if mouse.Buttons()&tcell.WheelDown != 0 {
		if mx >= listX && mx < listX+listW && my >= listY+1 && my < listY+1+listH {
			maxH := n.Dropdown.visibleItemCount()
			n.State.ScrollOffset++
			if n.State.ScrollOffset+maxH > len(n.Dropdown.Options) {
				n.State.ScrollOffset = len(n.Dropdown.Options) - maxH
				if n.State.ScrollOffset < 0 {
					n.State.ScrollOffset = 0
				}
			}
			return true
		}
	}

	return false
}
