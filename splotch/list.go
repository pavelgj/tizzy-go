package splotch

import (
	"fmt"
	"github.com/gdamore/tcell/v2"
)

// List is a component that displays a list of selectable items.
type List struct {
	Style             Style
	Items             []any
	RenderItem        func(item any, index int, selected bool, cursor bool) Node
	OnSelect          func(int)
	OnSelectionChange func(int)
	OnFocus           func(state *ListState)
}

func (l *List) node() {}

// NewList creates a new List component.
func NewList(ctx *RenderContext, style Style, key string, items []any, initialSelectedIndex int, renderItem func(item any, index int, selected bool, cursor bool) Node, onSelect func(int)) *List {
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
		ctx.hookIndex++
	}

	stateObj, ok := ctx.app.componentStates[style.ID]
	var state *ListState
	if !ok {
		state = &ListState{SelectedIndex: initialSelectedIndex, CursorIndex: 0, Key: key}
		if initialSelectedIndex >= 0 {
			state.CursorIndex = initialSelectedIndex
		}
		ctx.app.componentStates[style.ID] = state
	} else {
		state = stateObj.(*ListState)
		if state.Key != key {
			state.SelectedIndex = initialSelectedIndex
			state.CursorIndex = 0
			if initialSelectedIndex >= 0 {
				state.CursorIndex = initialSelectedIndex
			}
			state.ScrollOffset = 0
			state.Key = key
		}
	}

	return &List{
		Style:      style,
		Items:      items,
		RenderItem: renderItem,
		OnSelect:   onSelect,
	}
}

// NewListItem creates a standard list item with selection and cursor highlighting.
func NewListItem(label string, selected bool, cursor bool) Node {
	style := Style{FillWidth: true}
	textColor := tcell.ColorWhite
	if selected {
		style.Background = tcell.ColorBlue
		textColor = tcell.ColorWhite
	}
	if cursor {
		style.Background = tcell.ColorGray
		textColor = tcell.ColorWhite
	}
	return NewBox(style, NewText(Style{Color: textColor, Background: style.Background}, label))
}

// ListState stores the state of a List component.
type ListState struct {
	SelectedIndex int
	CursorIndex   int
	ScrollOffset  int
	Key           string
}

// GetStyle returns the style of the List node.
func (l *List) GetStyle() Style {
	return l.Style
}

// Layout calculates the layout for the List component.
func (l *List) Layout(x, y int, c Constraints) LayoutResult {
	w := c.MaxW
	if l.Style.Width > 0 {
		w = l.Style.Width
	}
	h := c.MaxH
	if l.Style.Height > 0 {
		h = l.Style.Height
	}

	return LayoutResult{
		Node: l,
		X:    x,
		Y:    y,
		W:    w,
		H:    h,
	}
}

// Render draws the List component to the grid.
func (l *List) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	stateObj, ok := componentStates[l.Style.ID]
	var state *ListState
	if !ok {
		state = &ListState{}
	} else {
		state = stateObj.(*ListState)
	}

	borderOffset := 0
	if l.Style.Border {
		borderOffset = 1
		borderStyle := tcell.StyleDefault.Foreground(l.Style.Color).Background(l.Style.Background)
		if l.Style.ID == focusedID {
			focusColor := tcell.ColorYellow
			if l.Style.FocusColor != tcell.ColorReset {
				focusColor = l.Style.FocusColor
			}
			focusBg := l.Style.Background
			if l.Style.FocusBackground != tcell.ColorReset {
				focusBg = l.Style.FocusBackground
			}
			borderStyle = tcell.StyleDefault.Foreground(focusColor).Background(focusBg)
		}
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, l.Style.Title, borderStyle)
	}

	viewportH := layout.H - borderOffset*2
	viewportW := layout.W - borderOffset*2

	curY := layout.Y + borderOffset
	curX := layout.X + borderOffset

	for i := 0; i < viewportH; i++ {
		idx := state.ScrollOffset + i
		if idx >= len(l.Items) {
			break
		}

		item := l.Items[idx]
		selected := idx == state.SelectedIndex
		cursor := idx == state.CursorIndex

		itemNode := l.RenderItem(item, idx, selected, cursor)

		// Layout item with full width and 1 line height
		itemLayout := Layout(itemNode, curX, curY, Constraints{MaxW: viewportW, MaxH: 1})

		// Ensure background fills the width if selected
		if selected {
			// We could modify the layout to fit the full width or let the child fill it
			// Let's assume RenderItem returns a Box or Text that respects layout width
			itemLayout.W = viewportW
		}

		Render(grid, itemLayout, focusedID, componentStates)

		curY++
	}
}

func (l *List) DefaultState() any {
	return &ListState{SelectedIndex: -1, CursorIndex: 0}
}

func (l *List) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s, ok := state.(*ListState)
	if !ok {
		return false
	}

	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return false
	}

	viewportH := 20
	borderOffset := 0
	if l.Style.Border {
		borderOffset = 1
	}
	if ctx.Layout.H > 0 {
		viewportH = ctx.Layout.H - borderOffset*2
	}

	dirty := false

	if key.Key() == tcell.KeyUp {
		if s.CursorIndex > 0 {
			s.CursorIndex--
			if s.CursorIndex < s.ScrollOffset {
				s.ScrollOffset = s.CursorIndex
			}
			dirty = true
			if l.OnSelectionChange != nil {
				l.OnSelectionChange(s.CursorIndex)
			}
		}
	} else if key.Key() == tcell.KeyDown {
		if s.CursorIndex < len(l.Items)-1 {
			s.CursorIndex++
			if s.CursorIndex >= s.ScrollOffset+viewportH {
				s.ScrollOffset = s.CursorIndex - viewportH + 1
			}
			dirty = true
			if l.OnSelectionChange != nil {
				l.OnSelectionChange(s.CursorIndex)
			}
		}
	} else if key.Key() == tcell.KeyPgUp {
		if len(l.Items) > 0 {
			s.CursorIndex -= viewportH
			if s.CursorIndex < 0 {
				s.CursorIndex = 0
			}
			if s.CursorIndex < s.ScrollOffset {
				s.ScrollOffset = s.CursorIndex
			}
			dirty = true
			if l.OnSelectionChange != nil {
				l.OnSelectionChange(s.CursorIndex)
			}
		}
	} else if key.Key() == tcell.KeyPgDn {
		if len(l.Items) > 0 {
			s.CursorIndex += viewportH
			if s.CursorIndex >= len(l.Items) {
				s.CursorIndex = len(l.Items) - 1
			}
			if s.CursorIndex >= s.ScrollOffset+viewportH {
				s.ScrollOffset = s.CursorIndex - viewportH + 1
				if s.ScrollOffset < 0 {
					s.ScrollOffset = 0
				}
			}
			dirty = true
			if l.OnSelectionChange != nil {
				l.OnSelectionChange(s.CursorIndex)
			}
		}
	} else if key.Key() == tcell.KeyEnter {
		s.SelectedIndex = s.CursorIndex
		dirty = true
		if l.OnSelect != nil {
			l.OnSelect(s.SelectedIndex)
		}
	}

	return dirty
}
