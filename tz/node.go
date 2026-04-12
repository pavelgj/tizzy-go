package tz

import "github.com/gdamore/tcell/v2"

// Node represents a component in the UI tree.
type Node interface {
	// GetStyle returns the style of the node.
	GetStyle() Style
}

// ParentNode indicates that a node has children.
type ParentNode interface {
	Node
	GetChildren() []Node
}

// Layoutable indicates that a node can calculate its own layout.
type Layoutable interface {
	Layout(x, y int, c Constraints) LayoutResult
}

// Renderable indicates that a node can render itself.
type Renderable interface {
	Render(grid *Grid, res LayoutResult, focusedID string, componentStates map[string]any)
}

// EventContext provides context for event handling.
type EventContext struct {
	Layout    LayoutResult
	// PopupOpen reports whether any Popup overlay is currently open. Components
	// can use this to suppress keys that should be handled by the popup instead
	// (e.g. TextInput suppresses Enter/Up/Down when a popup is open above it).
	PopupOpen bool
}

// EventHandler indicates that a node can handle events.
type EventHandler interface {
	HandleEvent(ev tcell.Event, state any, ctx EventContext) bool
	DefaultState() any
}

// CursorProvider allows a component to manage the terminal hardware cursor.
// The framework calls UpdateScrollOffset after layout and GetCursorPosition
// after rendering, then calls screen.ShowCursor with the result.
type CursorProvider interface {
	// UpdateScrollOffset adjusts internal scroll offsets so the cursor stays
	// visible. Called after layout, before rendering.
	UpdateScrollOffset(layout LayoutResult, state any)

	// GetCursorPosition returns the screen coordinates where the terminal
	// cursor should appear, and whether it should be shown at all.
	GetCursorPosition(layout LayoutResult, state any) (x, y int, show bool)
}

// FocusGainHandler allows a component to react when it receives focus.
// The framework calls this after updating the focused ID.
type FocusGainHandler interface {
	OnFocusGained(state any)
}

// Dismissable allows a component to be closed/reset when another component
// gains focus. The framework calls Dismiss on every Dismissable node (except
// the one that just received focus) whenever focus changes.
type Dismissable interface {
	Dismiss(state any)
}

// FocusScope allows a component to control which of its children participate
// in focus traversal. Components that don't implement this fall back to
// traversing all children via ParentNode.GetChildren().
type FocusScope interface {
	FocusableChildren(componentStates map[string]any) []Node
}

// CustomHitTester allows components to override default hit testing for children
type CustomHitTester interface {
	FindNodePathAt(x, y int, res LayoutResult, componentStates map[string]any) []Node
}

// Focusable indicates that a node can receive focus.
type Focusable interface {
	IsFocusable() bool
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
	ID              string
	Title           string
	Focusable       bool
	Multiline       bool
	Width           int // 0 means auto or stretch? Let's assume absolute or percentage later. For now, absolute cells.
	Height          int
	MaxHeight       int
	FlexDirection   string // "row" or "column"
	JustifyContent  string // "flex-start", "center", "flex-end"
	Border          bool
	Padding         Padding
	Margin          Margin
	Color           tcell.Color
	Background      tcell.Color
	FocusColor      tcell.Color
	FocusBackground tcell.Color
	FillWidth       bool
	FillHeight      bool
	GridRow         int
	GridCol         int
	GridRowSpan     int
	GridColSpan     int
}
