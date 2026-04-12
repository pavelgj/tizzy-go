package tz

// Portal renders its child at an absolute screen position, in a Z-layer above
// the main tree. The framework collects Portal nodes during tree traversal and
// renders them in Z-order after the main render pass.
//
// X, Y are the absolute screen coordinates of the portal's top-left corner.
// Set X to -1 to auto-center the portal on the screen.
// If PositionFn is set, it overrides X and Y.
//
// TrapFocus restricts Tab/Shift-Tab focus traversal to nodes inside the portal.
// OnOutsideClick is called when the user clicks outside the portal's bounds.
// PopupMode signals to underlying input components (e.g. TextInput) that a
// popup is open and they should suppress navigation key handling.
type Portal struct {
	Style          Style
	Child          Node
	X, Y           int
	PositionFn     func(screenW, screenH int, mainLayout LayoutResult) (x, y, maxW, maxH int)
	OnOutsideClick func()
	TrapFocus      bool
	PopupMode      bool
}

// GetStyle returns the portal's style.
func (p *Portal) GetStyle() Style {
	return p.Style
}

// GetChildren returns the portal's child, allowing tree traversal into the
// portal's subtree for node-lookup purposes. Portal children are NOT included
// in normal focus traversal (see TrapFocus and app.go).
func (p *Portal) GetChildren() []Node {
	if p.Child != nil {
		return []Node{p.Child}
	}
	return nil
}

// Layout returns a zero-size result. Portals do not consume space in the
// main layout — they are positioned and sized independently.
func (p *Portal) Layout(x, y int, c Constraints) LayoutResult {
	return LayoutResult{Node: p, X: x, Y: y, W: 0, H: 0}
}

// Render is a no-op for Portal. The framework renders portal content via a
// separate pass in RenderFrame.
func (p *Portal) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
}

// collectedPortal holds a Portal node together with its computed content layout,
// resolved after the main layout pass.
type collectedPortal struct {
	portal  *Portal
	content LayoutResult
}

// computePortalLayout resolves the portal's position and lays out its child.
func computePortalLayout(p *Portal, screenW, screenH int, mainLayout LayoutResult) LayoutResult {
	var x, y, maxW, maxH int

	if p.PositionFn != nil {
		x, y, maxW, maxH = p.PositionFn(screenW, screenH, mainLayout)
	} else if p.X == -1 {
		// Auto-center: measure child first, then center.
		maxW = screenW - 4
		maxH = screenH - 4
		if maxW < 0 {
			maxW = 0
		}
		if maxH < 0 {
			maxH = 0
		}
		measured := Layout(p.Child, 0, 0, Constraints{MaxW: maxW, MaxH: maxH})
		x = (screenW - measured.W) / 2
		y = (screenH - measured.H) / 2
		maxW = screenW - x
		maxH = screenH - y
	} else {
		x, y = p.X, p.Y
		maxW = screenW - x
		maxH = screenH - y
	}

	if maxW < 0 {
		maxW = 0
	}
	if maxH < 0 {
		maxH = 0
	}

	return Layout(p.Child, x, y, Constraints{MaxW: maxW, MaxH: maxH})
}

// collectPortals walks the main layout result tree and collects all Portal
// nodes, computing their real screen layouts via computePortalLayout.
// Portals are appended in tree order (last = topmost for Z-ordering).
func collectPortals(res LayoutResult, screenW, screenH int, mainLayout LayoutResult, portals *[]collectedPortal) {
	if p, ok := res.Node.(*Portal); ok {
		content := computePortalLayout(p, screenW, screenH, mainLayout)
		*portals = append(*portals, collectedPortal{portal: p, content: content})
		// Do not recurse into Portal's zero-size layout children.
		return
	}
	for _, child := range res.Children {
		collectPortals(child, screenW, screenH, mainLayout, portals)
	}
}
