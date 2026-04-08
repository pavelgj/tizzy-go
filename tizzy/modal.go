package tizzy

import "fmt"

// Modal represents a dialog overlay component.
type Modal struct {
	Style Style
	Child Node
}


// NewModal creates a new Modal component.
func NewModal(ctx *RenderContext, style Style, child Node, isOpen bool) *Modal {
	stateObj, _ := ctx.UseState(&ModalState{Open: isOpen})
	state := stateObj.(*ModalState)
	state.Open = isOpen // Sync with passed prop

	// Derive hook ID and set it on style
	id := fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	style.ID = id

	return &Modal{
		Style: style,
		Child: child,
	}
}

// ModalState stores the interactive state of a Modal.
type ModalState struct {
	Open bool
}

// GetStyle returns the style of the Modal node.
func (m *Modal) GetStyle() Style {
	return m.Style
}

// Layout calculates the layout for the Modal component.
func (n *Modal) Layout(x, y int, c Constraints) LayoutResult {
	return LayoutResult{
		Node: n,
		X:    x,
		Y:    y,
		W:    0,
		H:    0,
	}
}

// Render draws the Modal component to the grid.
func (n *Modal) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	// Do nothing, rendered as overlay in App.Run
}
