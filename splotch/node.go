package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// Node represents a component in the UI tree.
type Node interface {
	// For now, this is a marker interface.
	// We will add methods as needed for layout and rendering.
}

// Padding defines the spacing inside a node.
type Padding struct {
	Top, Bottom, Left, Right int
}

// Margin defines the spacing outside a node.
type Margin struct {
	Top, Bottom, Left, Right int
}

// Style defines the layout and appearance of a node.
type Style struct {
	ID             string
	Focusable      bool
	Width          int // 0 means auto or stretch? Let's assume absolute or percentage later. For now, absolute cells.
	Height         int
	FlexDirection  string // "row" or "column"
	JustifyContent string // "flex-start", "center", "flex-end"
	Border         bool
	Padding        Padding
	Margin         Margin
	Color          tcell.Color
	Background     tcell.Color
}

// Box is a container node that can hold children.
type Box struct {
	Style    Style
	Children []Node
}

// Text is a leaf node that displays text.
type Text struct {
	Style   Style
	Content string
}

// NewBox creates a new Box node.
func NewBox(style Style, children ...Node) *Box {
	return &Box{Style: style, Children: children}
}

// NewText creates a new Text node.
func NewText(style Style, content string) *Text {
	return &Text{Style: style, Content: content}
}
