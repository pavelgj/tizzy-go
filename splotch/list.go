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
}

func (l *List) node() {}

// NewList creates a new List component.
func NewList(ctx *RenderContext, style Style, key string, items []any, renderItem func(item any, index int, selected bool, cursor bool) Node, onSelect func(int)) *List {
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
		ctx.hookIndex++
	}

	stateObj, ok := ctx.app.componentStates[style.ID]
	var state *ListState
	if !ok {
		state = &ListState{SelectedIndex: -1, CursorIndex: 0, Key: key}
		ctx.app.componentStates[style.ID] = state
	} else {
		state = stateObj.(*ListState)
		if state.Key != key {
			state.SelectedIndex = -1
			state.CursorIndex = 0
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
			borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
		}
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
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
