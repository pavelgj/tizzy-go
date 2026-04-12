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

// TabsState tracks the active tab index and scroll offset.
type TabsState struct {
	ActiveTab    int
	ScrollOffset int // index of the first visible tab when overflowing
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

// tabsLastVisible returns the index of the last tab that fits within availW,
// starting from scrollOffset. It accounts for scroll arrows (</>).
func tabsLastVisible(tabs []Tab, scrollOffset, availW int) int {
	leftW := 0
	if scrollOffset > 0 {
		leftW = 1 // space for '<' arrow
	}

	// First pass: try without right arrow
	usedW := leftW
	last := scrollOffset - 1
	for i := scrollOffset; i < len(tabs); i++ {
		slotW := len(tabs[i].Label) + 4
		if usedW+slotW > availW {
			break
		}
		last = i
		usedW += slotW
	}
	if last >= len(tabs)-1 {
		return last // all remaining tabs fit, no right arrow needed
	}

	// More tabs remain: recompute reserving 1 char for the '>' arrow
	usedW = leftW
	last = scrollOffset - 1
	for i := scrollOffset; i < len(tabs); i++ {
		slotW := len(tabs[i].Label) + 4
		if usedW+slotW+1 > availW {
			break
		}
		last = i
		usedW += slotW
	}
	return last
}

// ensureActiveVisible adjusts ScrollOffset so that ActiveTab is within the
// visible range given the available header width.
func (n *Tabs) ensureActiveVisible(s *TabsState, availW int) {
	// Scroll left: active tab is before the visible window
	if s.ActiveTab < s.ScrollOffset {
		s.ScrollOffset = s.ActiveTab
		return
	}
	// Scroll right: keep advancing ScrollOffset until active tab is visible
	for s.ScrollOffset < len(n.Tabs)-1 {
		last := tabsLastVisible(n.Tabs, s.ScrollOffset, availW)
		if last >= s.ActiveTab {
			break
		}
		s.ScrollOffset++
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

	var state *TabsState
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			state, _ = stateObj.(*TabsState)
		}
	}
	activeTabIndex := 0
	scrollOffset := 0
	if state != nil {
		activeTabIndex = state.ActiveTab
		scrollOffset = state.ScrollOffset
	}

	isFocused := n.Style.ID != "" && n.Style.ID == focusedID

	var activeStyle tcell.Style
	if isFocused {
		activeStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	} else {
		activeStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
	}

	topY := layout.Y + n.Style.Padding.Top
	sepY := topY + 1
	startX := layout.X + n.Style.Padding.Left
	endX := layout.X + layout.W - n.Style.Padding.Right
	availW := endX - startX

	// Compute total header width to decide if overflow scrolling is needed
	totalHeadersW := 0
	for _, tab := range n.Tabs {
		totalHeadersW += len(tab.Label) + 4
	}

	// Determine the visible tab range
	firstVisible := 0
	lastVisible := len(n.Tabs) - 1
	if totalHeadersW > availW {
		firstVisible = scrollOffset
		lastVisible = tabsLastVisible(n.Tabs, scrollOffset, availW)
	}

	leftArrow := firstVisible > 0
	rightArrow := lastVisible >= 0 && lastVisible < len(n.Tabs)-1

	// Build slot positions for visible tabs
	type tabSlot struct {
		tabIdx int
		x      int
		width  int
	}
	var slots []tabSlot
	curX := startX
	if leftArrow {
		curX++ // reserve 1 char for '<'
	}
	for i := firstVisible; i <= lastVisible; i++ {
		slotW := len(n.Tabs[i].Label) + 4
		slots = append(slots, tabSlot{i, curX, slotW})
		curX += slotW
	}

	// Draw scroll arrows
	if leftArrow {
		grid.SetContent(startX, topY, '<', baseStyle)
		grid.SetContent(startX, sepY, '─', baseStyle)
	}
	if rightArrow {
		grid.SetContent(endX-1, topY, '>', baseStyle)
		grid.SetContent(endX-1, sepY, '─', baseStyle)
	}

	// Row 0: label row (visible tabs only)
	for _, slot := range slots {
		tab := n.Tabs[slot.tabIdx]
		if slot.tabIdx == activeTabIndex {
			// Draw: ╭ space label space ╮
			grid.SetContent(slot.x, topY, '╭', activeStyle)
			grid.SetContent(slot.x+1, topY, ' ', activeStyle)
			drawText(grid, slot.x+2, topY, tab.Label, activeStyle)
			grid.SetContent(slot.x+2+len(tab.Label), topY, ' ', activeStyle)
			grid.SetContent(slot.x+slot.width-1, topY, '╮', activeStyle)
		} else {
			// Draw: space space label space space
			drawText(grid, slot.x, topY, "  "+tab.Label+"  ", baseStyle)
		}
	}

	// Row 1: separator line with a break (╯...╰) under the active tab
	for _, slot := range slots {
		if slot.tabIdx == activeTabIndex {
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

	// Extend separator to fill remaining space between last slot and right arrow (or endX)
	lastSlotEnd := startX
	if leftArrow {
		lastSlotEnd = startX + 1
	}
	if len(slots) > 0 {
		last := slots[len(slots)-1]
		lastSlotEnd = last.x + last.width
	}
	sepEnd := endX
	if rightArrow {
		sepEnd = endX - 1
	}
	for x := lastSlotEnd; x < sepEnd; x++ {
		grid.SetContent(x, sepY, '─', baseStyle)
	}

	// Render ONLY the active child content
	if activeTabIndex >= 0 && activeTabIndex < len(layout.Children) {
		Render(grid, layout.Children[activeTabIndex], focusedID, componentStates)
	}
}

// GetChildren returns all tab content nodes.
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
	availW := ctx.Layout.W - n.Style.Padding.Left - n.Style.Padding.Right

	if key, ok := ev.(*tcell.EventKey); ok {
		if key.Key() == tcell.KeyRight {
			s.ActiveTab++
			if s.ActiveTab >= len(n.Tabs) {
				s.ActiveTab = 0
				s.ScrollOffset = 0
			}
			n.ensureActiveVisible(s, availW)
			return true
		} else if key.Key() == tcell.KeyLeft {
			s.ActiveTab--
			if s.ActiveTab < 0 {
				s.ActiveTab = len(n.Tabs) - 1
				s.ScrollOffset = 0 // ensureActiveVisible will scroll right as needed
			}
			n.ensureActiveVisible(s, availW)
			return true
		}
	} else if mouse, ok := ev.(*tcell.EventMouse); ok {
		mx, my := mouse.Position()
		topY := ctx.Layout.Y + n.Style.Padding.Top
		startX := ctx.Layout.X + n.Style.Padding.Left
		endX := ctx.Layout.X + ctx.Layout.W - n.Style.Padding.Right

		if my == topY {
			totalHeadersW := 0
			for _, tab := range n.Tabs {
				totalHeadersW += len(tab.Label) + 4
			}

			if totalHeadersW > availW {
				// Overflow mode: handle arrow clicks and visible tab clicks
				leftArrow := s.ScrollOffset > 0
				lastVisible := tabsLastVisible(n.Tabs, s.ScrollOffset, availW)
				rightArrow := lastVisible < len(n.Tabs)-1

				if leftArrow && mx == startX {
					s.ScrollOffset--
					return true
				}
				if rightArrow && mx == endX-1 {
					s.ScrollOffset++
					return true
				}

				curX := startX
				if leftArrow {
					curX++
				}
				for i := s.ScrollOffset; i <= lastVisible; i++ {
					slotW := len(n.Tabs[i].Label) + 4
					if mx >= curX && mx < curX+slotW {
						if s.ActiveTab != i {
							s.ActiveTab = i
							return true
						}
						break
					}
					curX += slotW
				}
			} else {
				// Normal mode: click any tab label
				curX := ctx.Layout.X + n.Style.Padding.Left
				for i, tab := range n.Tabs {
					slotW := len(tab.Label) + 4
					if mx >= curX && mx < curX+slotW {
						if s.ActiveTab != i {
							s.ActiveTab = i
							return true
						}
						break
					}
					curX += slotW
				}
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
