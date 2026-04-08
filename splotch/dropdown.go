package splotch

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
}

func (d *Dropdown) node() {}

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

	maxW := 0
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

	selectedText := ""
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

func (d *Dropdown) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s, ok := state.(*DropdownState)
	if !ok {
		return false
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
