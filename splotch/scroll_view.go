package splotch

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
	_, _ = UseState[*ScrollViewState](ctx, &ScrollViewState{ScrollOffset: 0})

	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	}

	return &ScrollView{
		Style: style,
		Child: child,
	}
}

// node implements the Node interface.
func (s *ScrollView) node() {}

// GetStyle returns the style of the ScrollView node.
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
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray)
	if n.Style.ID != "" && n.Style.ID == focusedID {
		borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow)
	}
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
	}
	
	pad := n.Style.Padding
	viewportW := layout.W - pad.Left - pad.Right - borderOffset*2
	viewportH := layout.H - pad.Top - pad.Bottom - borderOffset*2
	
	scrollOffset := 0
	if n.Style.ID != "" && componentStates != nil {
		if stateObj, ok := componentStates[n.Style.ID]; ok {
			state := stateObj.(*ScrollViewState)
			scrollOffset = state.ScrollOffset
		}
	}
	
	if len(layout.Children) > 0 {
		childLayout := layout.Children[0]
		
		tempGrid := NewGrid(viewportW, viewportH)
		
		shiftedLayout := shiftLayout(childLayout, -childLayout.X, -childLayout.Y - scrollOffset)
		
		Render(tempGrid, shiftedLayout, focusedID, componentStates)
		
		for y := 0; y < viewportH; y++ {
			for x := 0; x < viewportW; x++ {
				cell := tempGrid.Cells[y][x]
				grid.SetContent(layout.X+pad.Left+borderOffset+x, layout.Y+pad.Top+borderOffset+y, cell.Rune, cell.Style)
			}
		}
	}
}
