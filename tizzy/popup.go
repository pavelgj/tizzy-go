package tizzy

import "fmt"

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
	// Do nothing, rendered as overlay in App.Run
}
