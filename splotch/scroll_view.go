package splotch

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
func NewScrollView(style Style, child Node) *ScrollView {
	return &ScrollView{
		Style: style,
		Child: child,
	}
}

// node implements the Node interface.
func (s *ScrollView) node() {}
