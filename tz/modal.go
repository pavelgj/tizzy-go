package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

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

func (s *ModalState) IsOpen() bool {
	return s.Open
}

// GetStyle returns the style of the Modal node.
func (m *Modal) GetStyle() Style {
	return m.Style
}

// GetChildren returns the children of the Modal node.
func (m *Modal) GetChildren() []Node {
	if m.Child != nil {
		return []Node{m.Child}
	}
	return nil
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

func (m *Modal) HandleOverlayEvent(ev tcell.Event, state any, ctx EventContext) (bool, *LayoutResult) {
	s, ok := state.(*ModalState)
	if !ok || !s.Open {
		return false, nil
	}

	mouse, ok := ev.(*tcell.EventMouse)
	if !ok {
		return false, nil
	}

	mx, my := mouse.Position()
	res := ctx.OverlayLayout

	if mouse.Buttons()&tcell.Button1 != 0 {
		// Check if click is inside modal content
		if mx >= res.X && mx < res.X+res.W && my >= res.Y && my < res.Y+res.H {
			// Click is inside modal content!
			// Tell app.go to search this layout!
			return false, &ctx.OverlayLayout
		}

		// Click is OUTSIDE modal content!
		// Trap it!
		return true, nil
	}

	return false, nil
}
