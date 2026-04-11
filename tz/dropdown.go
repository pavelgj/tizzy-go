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
}

// NewDropdown creates a new Dropdown component.
func NewDropdown(ctx *RenderContext, style Style, options []string, selectedIndex int, onChange func(int), maxListHeight ...int) *Dropdown {
	_, _ = UseState[*DropdownState](ctx, &DropdownState{Open: false, FocusedIndex: selectedIndex})

	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	}

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
	}
}

// DropdownState stores the interactive state of a Dropdown.
type DropdownState struct {
	Open         bool
	FocusedIndex int
	ScrollOffset int
	OpenAbove    bool
}

func (s *DropdownState) IsOpen() bool {
	return s.Open
}

// GetStyle returns the style of the Dropdown node.
func (d *Dropdown) GetStyle() Style {
	return d.Style
}

// Layout calculates the layout for the Dropdown component.
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

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}
}

// Render draws the Dropdown component to the grid.
func (n *Dropdown) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	borderOffset := 0
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	borderStyle := style
	if n.Style.ID != "" && n.Style.ID == focusedID {
		borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	pad := n.Style.Padding
	curX := layout.X + pad.Left + borderOffset
	curY := layout.Y + pad.Top + borderOffset

	selectedText := n.Placeholder
	if n.SelectedIndex >= 0 && n.SelectedIndex < len(n.Options) {
		selectedText = n.Options[n.SelectedIndex]
	}

	drawText(grid, curX, curY, "[ ", style)
	curX += 2
	drawText(grid, curX, curY, selectedText, style)
	curX += len(selectedText)

	remaining := layout.W - borderOffset*2 - pad.Left - pad.Right - 2 - len(selectedText) - 4
	if remaining > 0 {
		for i := 0; i < remaining; i++ {
			grid.SetContent(curX+i, curY, ' ', style)
		}
		curX += remaining
	}

	drawText(grid, curX, curY, " v ]", style)
}

func (d *Dropdown) DefaultState() any {
	return &DropdownState{}
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

			maxH := d.MaxListHeight
			if maxH <= 0 {
				maxH = 5
			}
			if maxH > len(d.Options) {
				maxH = len(d.Options)
			}

			if s.FocusedIndex < s.ScrollOffset {
				s.ScrollOffset = s.FocusedIndex
			}
			if s.FocusedIndex >= s.ScrollOffset+maxH {
				s.ScrollOffset = s.FocusedIndex - maxH + 1
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyDown {
		if s.Open {
			if s.FocusedIndex < len(d.Options)-1 {
				s.FocusedIndex++
			}

			maxH := d.MaxListHeight
			if maxH <= 0 {
				maxH = 5
			}
			if maxH > len(d.Options) {
				maxH = len(d.Options)
			}

			if s.FocusedIndex >= s.ScrollOffset+maxH {
				s.ScrollOffset = s.FocusedIndex - maxH + 1
			}
			if s.FocusedIndex < s.ScrollOffset {
				s.ScrollOffset = 0
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyPgUp {
		if s.Open {
			maxH := d.MaxListHeight
			if maxH <= 0 {
				maxH = 5
			}
			if maxH > len(d.Options) {
				maxH = len(d.Options)
			}

			s.FocusedIndex -= maxH
			if s.FocusedIndex < 0 {
				s.FocusedIndex = 0
			}

			s.ScrollOffset -= maxH
			if s.ScrollOffset < 0 {
				s.ScrollOffset = 0
			}
			dirty = true
		}
	} else if key.Key() == tcell.KeyPgDn {
		if s.Open {
			maxH := d.MaxListHeight
			if maxH <= 0 {
				maxH = 5
			}
			if maxH > len(d.Options) {
				maxH = len(d.Options)
			}

			s.FocusedIndex += maxH
			if s.FocusedIndex >= len(d.Options) {
				s.FocusedIndex = len(d.Options) - 1
			}

			s.ScrollOffset += maxH
			if s.ScrollOffset+maxH > len(d.Options) {
				s.ScrollOffset = len(d.Options) - maxH
				if s.ScrollOffset < 0 {
					s.ScrollOffset = 0
				}
			}
			dirty = true
		}
	}

	return dirty
}

// RenderOverlay renders the Dropdown's open list on top of the main grid.
func (d *Dropdown) RenderOverlay(grid *Grid, screenW, screenH int, mainLayout LayoutResult, focusedID string, componentStates map[string]any) {
	state, ok := componentStates[d.Style.ID].(*DropdownState)
	if !ok || !state.Open {
		return
	}

	res := findLayoutResultByID(mainLayout, d.Style.ID)
	if res == nil {
		return
	}

	listY := res.Y + res.H
	listW := res.W

	maxH := d.MaxListHeight
	if maxH <= 0 {
		maxH = 5
	}
	if maxH > len(d.Options) {
		maxH = len(d.Options)
	}
	listH := maxH

	style := tcell.StyleDefault.Foreground(d.Style.Color).Background(tcell.ColorBlack)
	popupH := listH + 2

	spaceBelow := screenH - (res.Y + res.H)
	if spaceBelow < popupH && res.Y >= popupH {
		state.OpenAbove = true
	} else {
		state.OpenAbove = false
	}

	if state.OpenAbove {
		listY = res.Y - popupH
	}

	// Draw shadow (right and bottom edges only)
	for i := 1; i <= popupH; i++ {
		if listY+i < screenH && res.X+listW < screenW {
			currentCell := grid.Cells[listY+i][res.X+listW]
			grid.SetContent(res.X+listW, listY+i, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
		}
	}
	for j := 1; j <= listW; j++ {
		if listY+popupH < screenH && res.X+j < screenW {
			currentCell := grid.Cells[listY+popupH][res.X+j]
			grid.SetContent(res.X+j, listY+popupH, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
		}
	}

	// Fill background
	for y := 0; y < popupH; y++ {
		for x := 0; x < listW; x++ {
			if listY+y < screenH && res.X+x < screenW {
				grid.SetContent(res.X+x, listY+y, ' ', style)
			}
		}
	}

	// Draw border
	drawBorder(grid, res.X, listY, listW, popupH, "", style)

	// Draw items
	for i := 0; i < listH; i++ {
		optIdx := i + state.ScrollOffset
		if optIdx >= len(d.Options) {
			break
		}
		opt := d.Options[optIdx]
		optStyle := style
		if optIdx == state.FocusedIndex {
			optStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}

		label := " " + opt
		for len(label) < listW-2 {
			label += " "
		}

		curX := res.X + 1
		for _, r := range label {
			if listY+1+i < screenH && curX < screenW && curX < res.X+listW-1 {
				grid.SetContent(curX, listY+1+i, r, optStyle)
				curX++
			}
		}
	}
}

func (d *Dropdown) HandleOverlayEvent(ev tcell.Event, state any, ctx EventContext) (bool, *LayoutResult) {
	s, ok := state.(*DropdownState)
	if !ok || !s.Open {
		return false, nil
	}

	mouse, ok := ev.(*tcell.EventMouse)
	if !ok {
		return false, nil
	}

	mx, my := mouse.Position()
	res := ctx.Layout
	listY := res.Y + res.H
	listW := res.W

	maxH := d.MaxListHeight
	if maxH <= 0 {
		maxH = 5
	}
	if maxH > len(d.Options) {
		maxH = len(d.Options)
	}
	listH := maxH
	if s.OpenAbove {
		listY = res.Y - (listH + 2)
	}

	if mouse.Buttons()&tcell.Button1 != 0 {
		if mx >= res.X && mx < res.X+listW && my >= listY+1 && my < listY+1+listH {
			clickedIndex := my - listY - 1 + s.ScrollOffset
			if clickedIndex >= 0 && clickedIndex < len(d.Options) {
				d.SelectedIndex = clickedIndex
				if d.OnChange != nil {
					d.OnChange(clickedIndex)
				}
				s.Open = false
				return true, nil
			}
		} else {
			if !(mx >= res.X && mx < res.X+res.W && my >= res.Y && my < res.Y+res.H) {
				s.Open = false
				return true, nil
			}
		}
	} else if mouse.Buttons()&tcell.WheelUp != 0 {
		if mx >= res.X && mx < res.X+listW && my >= listY+1 && my < listY+1+listH {
			s.ScrollOffset--
			if s.ScrollOffset < 0 {
				s.ScrollOffset = 0
			}
			return true, nil
		}
	} else if mouse.Buttons()&tcell.WheelDown != 0 {
		if mx >= res.X && mx < res.X+listW && my >= listY+1 && my < listY+1+listH {
			s.ScrollOffset++
			if s.ScrollOffset+maxH > len(d.Options) {
				s.ScrollOffset = len(d.Options) - maxH
				if s.ScrollOffset < 0 {
					s.ScrollOffset = 0
				}
			}
			return true, nil
		}
	}

	return false, nil
}
