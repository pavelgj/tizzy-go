package splotch

// LayoutResult holds the calculated position and size of a node.
type LayoutResult struct {
	Node     Node
	X, Y     int
	W, H     int
	Children []LayoutResult
}

// Constraints defines the maximum size allowed for a node.
type Constraints struct {
	MaxW, MaxH int
}

// Layout calculates the layout of a node and its children.
func Layout(node Node, x, y int, c Constraints) LayoutResult {
	if l, ok := node.(Layoutable); ok {
		return l.Layout(x, y, c)
	}

	return LayoutResult{Node: node, X: x, Y: y, W: 0, H: 0}
}
