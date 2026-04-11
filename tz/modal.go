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
	Open         bool
	OverlayLayout LayoutResult // computed during RenderOverlay, used by event handlers
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

// RenderOverlay renders the Modal as a centered overlay on top of the main grid.
// It also stores the computed content LayoutResult in ModalState.OverlayLayout so
// that event handlers can use it without recomputing layout.
func (m *Modal) RenderOverlay(grid *Grid, screenW, screenH int, mainLayout LayoutResult, focusedID string, componentStates map[string]any) {
	state, ok := componentStates[m.Style.ID].(*ModalState)
	if !ok || !state.Open {
		return
	}

	maxModalW := screenW - 4
	maxModalH := screenH - 4
	if maxModalW < 0 {
		maxModalW = 0
	}
	if maxModalH < 0 {
		maxModalH = 0
	}

	modalConstraints := Constraints{MaxW: maxModalW, MaxH: maxModalH}

	// Layout at 0,0 first to measure size
	measured := Layout(m.Child, 0, 0, modalConstraints)

	modalW := measured.W + 2
	modalH := measured.H + 2
	if modalW > screenW {
		modalW = screenW
	}
	if modalH > screenH {
		modalH = screenH
	}

	modalX := (screenW - modalW) / 2
	modalY := (screenH - modalH) / 2

	// Final layout at correct position
	contentLayout := Layout(m.Child, modalX+1, modalY+1, modalConstraints)

	// Store for use by event handlers
	state.OverlayLayout = contentLayout

	style := tcell.StyleDefault.Foreground(m.Style.Color).Background(m.Style.Background)
	drawBorder(grid, modalX, modalY, modalW, modalH, "", style)

	for y := modalY + 1; y < modalY+modalH-1; y++ {
		for x := modalX + 1; x < modalX+modalW-1; x++ {
			grid.SetContent(x, y, ' ', style)
		}
	}

	Render(grid, contentLayout, focusedID, componentStates)
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
