package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Popup represents a floating overlay component positioned absolutely.
type Popup struct {
	Style Style
	Child Node
	X     int
	Y     int
}

// NewPopup creates a new Popup component.
func NewPopup(ctx *RenderContext, style Style, child Node, x, y int, isOpen bool) *Popup {
	stateObj, _ := ctx.UseState(&PopupState{Open: isOpen})
	state := stateObj.(*PopupState)
	state.Open = isOpen // Sync with passed prop

	// Derive hook ID and set it on style
	id := fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	style.ID = id

	return &Popup{
		Style: style,
		Child: child,
		X:     x,
		Y:     y,
	}
}

// PopupState stores the interactive state of a Popup.
type PopupState struct {
	Open bool
}

func (s *PopupState) IsOpen() bool {
	return s.Open
}

// GetStyle returns the style of the Popup node.
func (p *Popup) GetStyle() Style {
	return p.Style
}

// Layout calculates the layout for the Popup component.
func (p *Popup) Layout(x, y int, c Constraints) LayoutResult {
	// Popups do not take space in the parent layout
	return LayoutResult{
		Node: p,
		X:    x,
		Y:    y,
		W:    0,
		H:    0,
	}
}

// Render draws the Popup component.
func (p *Popup) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	// Do nothing, rendered as overlay via RenderOverlay.
}

// RenderOverlay renders the Popup at its absolute position on top of the main grid.
func (p *Popup) RenderOverlay(grid *Grid, screenW, screenH int, mainLayout LayoutResult, focusedID string, componentStates map[string]any) {
	state, ok := componentStates[p.Style.ID].(*PopupState)
	if !ok || !state.Open {
		return
	}

	maxPopupW := screenW - p.X
	maxPopupH := screenH - p.Y
	if maxPopupW < 0 {
		maxPopupW = 0
	}
	if maxPopupH < 0 {
		maxPopupH = 0
	}

	popupConstraints := Constraints{MaxW: maxPopupW, MaxH: maxPopupH}
	popupLayout := Layout(p.Child, p.X, p.Y, popupConstraints)

	// Fill background to prevent see-through
	style := tcell.StyleDefault.Foreground(p.Style.Color).Background(p.Style.Background)
	for y := p.Y; y < p.Y+popupLayout.H; y++ {
		for x := p.X; x < p.X+popupLayout.W; x++ {
			if x < screenW && y < screenH {
				grid.SetContent(x, y, ' ', style)
			}
		}
	}

	Render(grid, popupLayout, focusedID, componentStates)
}
