package tz

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// ScrollView is a node that acts as a viewport for a single child.
type ScrollView struct {
	Style Style
	Child Node
}

// ScrollViewState stores the mutable state of a ScrollView.
type ScrollViewState struct {
	ScrollOffset int
}

// NewScrollView creates a new ScrollView node.
func NewScrollView(ctx *RenderContext, style Style, child Node) *ScrollView {
	if style.ID != "" {
		// Use custom ID for state
		if _, ok := ctx.app.componentStates[style.ID]; !ok {
			ctx.app.componentStates[style.ID] = &ScrollViewState{ScrollOffset: 0}
		}
	} else {
		_, _ = UseState[*ScrollViewState](ctx, &ScrollViewState{ScrollOffset: 0})
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	}

	return &ScrollView{
		Style: style,
		Child: child,
	}
}

// GetStyle returns the style of the ScrollView node.
func (s *ScrollView) GetChildren() []Node {
	if s.Child != nil {
		return []Node{s.Child}
	}
	return nil
}

func (s *ScrollView) GetStyle() Style {
	return s.Style
}

// Layout calculates the layout for the ScrollView component.
func (n *ScrollView) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	w := 20
	h := 10

	if n.Style.Width > 0 {
		w = n.Style.Width
	}
	if n.Style.FillWidth {
		w = c.MaxW - pad.Left - pad.Right - margin.Left - margin.Right - borderSize
		if w < 0 {
			w = 0
		}
	}

	if n.Style.Height > 0 {
		h = n.Style.Height
	}
	if n.Style.FillHeight {
		h = c.MaxH - pad.Top - pad.Bottom - margin.Top - margin.Bottom - borderSize
		if h < 0 {
			h = 0
		}
	}

	viewportW := w
	viewportH := h

	childConstraints := Constraints{
		MaxW: viewportW,
		MaxH: 10000,
	}

	childX := boxX + borderSize + pad.Left
	childY := boxY + borderSize + pad.Top

	var childRes LayoutResult
	if n.Child != nil {
		childRes = Layout(n.Child, childX, childY, childConstraints)
	}

	return LayoutResult{
		Node:     n,
		X:        boxX,
		Y:        boxY,
		W:        viewportW + pad.Left + pad.Right + borderSize,
		H:        viewportH + pad.Top + pad.Bottom + borderSize,
		Children: []LayoutResult{childRes},
	}
}

// Render draws the ScrollView component to the grid.
func (n *ScrollView) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	borderOffset := 0
	focused := n.Style.ID != "" && n.Style.ID == focusedID
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	if focused {
		focusColor := tcell.ColorYellow
		if n.Style.FocusColor != tcell.ColorDefault {
			focusColor = n.Style.FocusColor
		}
		focusBg := n.Style.Background
		if n.Style.FocusBackground != tcell.ColorDefault {
			focusBg = n.Style.FocusBackground
		}
		borderStyle = tcell.StyleDefault.Foreground(focusColor).Background(focusBg)
	}
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, n.Style.Title, borderStyle)
	}

	pad := n.Style.Padding
	viewportW := layout.W - pad.Left - pad.Right - borderOffset*2
	viewportH := layout.H - pad.Top - pad.Bottom - borderOffset*2

	var state *ScrollViewState
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			state = stateObj.(*ScrollViewState)
		}
	}
	scrollOffset := 0
	if state != nil {
		scrollOffset = state.ScrollOffset
	}

	if len(layout.Children) > 0 {
		childLayout := layout.Children[0]
		contentH := childLayout.H

		// Clamp scroll offset so we never scroll past the content.
		if state != nil && contentH > viewportH {
			maxOffset := contentH - viewportH
			if state.ScrollOffset > maxOffset {
				state.ScrollOffset = maxOffset
				scrollOffset = maxOffset
			}
		} else if state != nil {
			state.ScrollOffset = 0
			scrollOffset = 0
		}

		tempGrid := NewGrid(viewportW, viewportH)
		shiftedLayout := shiftLayout(childLayout, -childLayout.X, -childLayout.Y-scrollOffset)
		Render(tempGrid, shiftedLayout, focusedID, componentStates)

		for y := 0; y < viewportH; y++ {
			for x := 0; x < viewportW; x++ {
				cell := tempGrid.Cells[y][x]
				grid.SetContent(layout.X+pad.Left+borderOffset+x, layout.Y+pad.Top+borderOffset+y, cell.Rune, cell.Style)
			}
		}

		// Scrollbar in the right border column (only when border is on and content overflows).
		if n.Style.Border && contentH > viewportH {
			trackH := layout.H - 2 // rows between the two corner characters
			scrollBarX := layout.X + layout.W - 1

			// Proportional thumb.
			thumbSize := trackH * viewportH / contentH
			if thumbSize < 1 {
				thumbSize = 1
			}
			thumbStart := 0
			if contentH > viewportH {
				thumbStart = (trackH - thumbSize) * scrollOffset / (contentH - viewportH)
			}

			for i := 0; i < trackH; i++ {
				ry := layout.Y + 1 + i
				if i >= thumbStart && i < thumbStart+thumbSize {
					grid.SetContent(scrollBarX, ry, '█', borderStyle)
				}
				// else: leave the existing │ border character
			}

			// ▲/▼ in the corner characters when there is hidden content.
			if scrollOffset > 0 {
				grid.SetContent(scrollBarX, layout.Y, '▲', borderStyle)
			}
			if scrollOffset+viewportH < contentH {
				grid.SetContent(scrollBarX, layout.Y+layout.H-1, '▼', borderStyle)
			}
		}
	}
}

// IsFocusable indicates that a node can receive focus.
func (n *ScrollView) IsFocusable() bool {
	return n.Style.Focusable
}

// DefaultState returns the default state for the scroll view.
func (n *ScrollView) DefaultState() any {
	return &ScrollViewState{ScrollOffset: 0}
}

// HandleEvent handles key events for scrolling.
func (n *ScrollView) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s, ok := state.(*ScrollViewState)
	if !ok {
		return false
	}

	if mouse, ok := ev.(MouseEvent); ok {
		if mouse.Buttons()&tcell.WheelUp != 0 {
			s.ScrollOffset--
			if s.ScrollOffset < 0 {
				s.ScrollOffset = 0
			}
			return true
		} else if mouse.Buttons()&tcell.WheelDown != 0 {
			s.ScrollOffset++
			return true
		}
	}

	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return false
	}

	dirty := false

	if key.Key() == tcell.KeyUp {
		s.ScrollOffset--
		if s.ScrollOffset < 0 {
			s.ScrollOffset = 0
		}
		dirty = true
	} else if key.Key() == tcell.KeyDown {
		s.ScrollOffset++
		dirty = true
	}

	return dirty
}

// FindNodePathAt overrides default hit testing to account for scroll offset.
func (sv *ScrollView) FindNodePathAt(x, y int, res LayoutResult, componentStates map[string]any) []Node {
	scrollOffset := 0
	if sv.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[sv.Style.ID]; ok {
			state := stateObj.(*ScrollViewState)
			scrollOffset = state.ScrollOffset
		}
	}

	for _, child := range res.Children {
		if path := findNodePathAt(child, x, y+scrollOffset, componentStates); path != nil {
			return append([]Node{sv}, path...)
		}
	}
	return []Node{sv}
}
