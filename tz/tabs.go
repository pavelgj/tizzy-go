package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Tab represents a single tab in a Tabs component.
type Tab struct {
	Label   string
	Content Node
}

// Tabs represents a component that allows switching between different content views.
type Tabs struct {
	Style Style
	Tabs  []Tab
}

// TabsState tracks the active tab index.
type TabsState struct {
	ActiveTab int
}

// NewTabs creates a new Tabs component.
func NewTabs(ctx *RenderContext, style Style, tabs []Tab) *Tabs {
	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
	}
	ctx.hookIndex++ // keep the hookIndex advancing as if UseState was called

	// Store state under style.ID — the same key used by Render and HandleEvent.
	ctx.app.mu.Lock()
	if ctx.app.componentStates[style.ID] == nil {
		ctx.app.componentStates[style.ID] = &TabsState{}
	}
	ctx.app.mu.Unlock()

	return &Tabs{
		Style: style,
		Tabs:  tabs,
	}
}

// Layout calculates the layout for the Tabs component.
func (n *Tabs) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	headerH := 2 // row 0: labels/top-borders, row 1: separator line
	headersW := 0
	for _, tab := range n.Tabs {
		headersW += len(tab.Label) + 4 // ╭ + space + label + space + ╮
	}

	childConstraints := Constraints{
		MaxW: c.MaxW - pad.Left - pad.Right,
		MaxH: c.MaxH - headerH - pad.Top - pad.Bottom,
	}
	if childConstraints.MaxW < 0 {
		childConstraints.MaxW = 0
	}
	if childConstraints.MaxH < 0 {
		childConstraints.MaxH = 0
	}

	var childrenLayouts []LayoutResult
	contentW := 0
	contentH := 0

	contentX := boxX + pad.Left
	contentY := boxY + headerH + pad.Top

	for _, tab := range n.Tabs {
		res := Layout(tab.Content, contentX, contentY, childConstraints)
		childrenLayouts = append(childrenLayouts, res)
		if res.W > contentW {
			contentW = res.W
		}
		if res.H > contentH {
			contentH = res.H
		}
	}

	w := headersW
	if contentW > w {
		w = contentW
	}
	w += pad.Left + pad.Right

	h := headerH + contentH + pad.Top + pad.Bottom

	if n.Style.Width > 0 {
		w = n.Style.Width
	}
	if n.Style.Height > 0 {
		h = n.Style.Height
	}

	return LayoutResult{
		Node:     n,
		X:        boxX,
		Y:        boxY,
		W:        w,
		H:        h,
		Children: childrenLayouts,
	}
}

// Render draws the Tabs component to the grid.
func (n *Tabs) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	baseStyle := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)

	activeTabIndex := 0
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			if state, ok := stateObj.(*TabsState); ok {
				activeTabIndex = state.ActiveTab
			}
		}
	}

	isFocused := n.Style.ID != "" && n.Style.ID == focusedID

	var activeStyle tcell.Style
	if isFocused {
		activeStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	} else {
		activeStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
	}

	// Row positions
	topY := layout.Y + n.Style.Padding.Top
	sepY := topY + 1
	startX := layout.X + n.Style.Padding.Left
	endX := layout.X + layout.W - n.Style.Padding.Right

	// Compute slot positions (each slot = ╭ + space + label + space + ╮)
	type tabSlot struct {
		x     int
		width int
	}
	slots := make([]tabSlot, len(n.Tabs))
	curX := startX
	for i, tab := range n.Tabs {
		slotW := len(tab.Label) + 4
		slots[i] = tabSlot{x: curX, width: slotW}
		curX += slotW
	}

	// Row 0: label row
	for i, tab := range n.Tabs {
		slot := slots[i]
		if i == activeTabIndex {
			// Draw: ╭ space label space ╮
			grid.SetContent(slot.x, topY, '╭', activeStyle)
			grid.SetContent(slot.x+1, topY, ' ', activeStyle)
			drawText(grid, slot.x+2, topY, tab.Label, activeStyle)
			grid.SetContent(slot.x+2+len(tab.Label), topY, ' ', activeStyle)
			grid.SetContent(slot.x+slot.width-1, topY, '╮', activeStyle)
		} else {
			// Draw: space space label space space (same width as active slot)
			drawText(grid, slot.x, topY, "  "+tab.Label+"  ", baseStyle)
		}
	}

	// Row 1: separator line, with a break (╯...╰) under the active tab
	for i, slot := range slots {
		if i == activeTabIndex {
			grid.SetContent(slot.x, sepY, '╯', activeStyle)
			for x := slot.x + 1; x < slot.x+slot.width-1; x++ {
				grid.SetContent(x, sepY, ' ', activeStyle)
			}
			grid.SetContent(slot.x+slot.width-1, sepY, '╰', activeStyle)
		} else {
			for x := slot.x; x < slot.x+slot.width; x++ {
				grid.SetContent(x, sepY, '─', baseStyle)
			}
		}
	}
	// Extend separator to fill remaining component width
	lastSlotEnd := slots[len(slots)-1].x + slots[len(slots)-1].width
	for x := lastSlotEnd; x < endX; x++ {
		grid.SetContent(x, sepY, '─', baseStyle)
	}

	// Render ONLY the active child content
	if activeTabIndex >= 0 && activeTabIndex < len(layout.Children) {
		Render(grid, layout.Children[activeTabIndex], focusedID, componentStates)
	}
}

// GetStyle returns the style of the Tabs component.
func (n *Tabs) GetChildren() []Node {
	var children []Node
	for _, tab := range n.Tabs {
		children = append(children, tab.Content)
	}
	return children
}

func (n *Tabs) DefaultState() any {
	return &TabsState{ActiveTab: 0}
}

func (n *Tabs) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s := state.(*TabsState)
	if key, ok := ev.(*tcell.EventKey); ok {
		if key.Key() == tcell.KeyRight {
			s.ActiveTab++
			if s.ActiveTab >= len(n.Tabs) {
				s.ActiveTab = 0
			}
			return true
		} else if key.Key() == tcell.KeyLeft {
			s.ActiveTab--
			if s.ActiveTab < 0 {
				s.ActiveTab = len(n.Tabs) - 1
			}
			return true
		}
	} else if mouse, ok := ev.(*tcell.EventMouse); ok {
		mx, my := mouse.Position()
		curX := ctx.Layout.X + n.Style.Padding.Left
		curY := ctx.Layout.Y + n.Style.Padding.Top

		if my == curY {
			for i, tab := range n.Tabs {
				labelLen := len(tab.Label) + 4 // "[ " + label + " ]"
				if mx >= curX && mx < curX+labelLen {
					if s.ActiveTab != i {
						s.ActiveTab = i
						return true
					}
					break
				}
				curX += labelLen
			}
		}
	}
	return false
}

func (n *Tabs) GetStyle() Style {
	return n.Style
}

// IsFocusable indicates that a node can receive focus.
func (n *Tabs) IsFocusable() bool {
	return n.Style.Focusable
}

// FocusableChildren returns only the active tab's content for focus traversal.
func (n *Tabs) FocusableChildren(componentStates map[string]any) []Node {
	activeIdx := 0
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			activeIdx = stateObj.(*TabsState).ActiveTab
		}
	}
	if activeIdx >= 0 && activeIdx < len(n.Tabs) {
		return []Node{n.Tabs[activeIdx].Content}
	}
	return nil
}

// FindNodePathAt overrides default hit testing to only search the active tab.
func (n *Tabs) FindNodePathAt(x, y int, res LayoutResult, componentStates map[string]any) []Node {
	activeIdx := 0
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			activeIdx = stateObj.(*TabsState).ActiveTab
		}
	}
	if activeIdx >= 0 && activeIdx < len(res.Children) {
		if path := findNodePathAt(res.Children[activeIdx], x, y, componentStates); path != nil {
			return append([]Node{res.Node}, path...)
		}
	}
	return []Node{res.Node}
}
