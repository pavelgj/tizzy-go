package splotch

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
	_, _ = UseState[*TabsState](ctx, &TabsState{ActiveTab: 0})

	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	}

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

	headerH := 1
	headersW := 0
	for _, tab := range n.Tabs {
		headersW += len(tab.Label) + 4 // "[ " + label + " ]"
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
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)

	activeTabIndex := 0
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			if state, ok := stateObj.(*TabsState); ok {
				activeTabIndex = state.ActiveTab
			}
		}
	}

	isFocused := n.Style.ID != "" && n.Style.ID == focusedID

	// Draw headers
	curX := layout.X + n.Style.Padding.Left
	curY := layout.Y + n.Style.Padding.Top

	for i, tab := range n.Tabs {
		label := "[ " + tab.Label + " ]"
		itemStyle := style
		if i == activeTabIndex {
			if isFocused {
				itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
			} else {
				itemStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
			}
		}
		drawText(grid, curX, curY, label, itemStyle)
		curX += len(label)
	}

	// Render ONLY the active child content
	if activeTabIndex >= 0 && activeTabIndex < len(layout.Children) {
		Render(grid, layout.Children[activeTabIndex], focusedID, componentStates)
	}
}

// GetStyle returns the style of the Tabs component.
func (n *Tabs) GetStyle() Style {
	return n.Style
}
