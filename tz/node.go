package tz

import (
	"github.com/gdamore/tcell/v2"
)

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
	Layout        LayoutResult
	PopupOpen     bool
	OverlayLayout LayoutResult
}

// EventHandler indicates that a node can handle events.
type EventHandler interface {
	HandleEvent(ev tcell.Event, state any, ctx EventContext) bool
	DefaultState() any
}

// OpenableState allows app.go to check if a component has an open overlay
type OpenableState interface {
	IsOpen() bool
}

// OverlayHandler allows app.go to delegate events to the overlay
type OverlayHandler interface {
	HandleOverlayEvent(ev tcell.Event, state any, ctx EventContext) (bool, *LayoutResult)
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

// Box is a container node that can hold children.
type Box struct {
	Style    Style
	Children []Node
}

// NewBox creates a new Box node.
func NewBox(style Style, children ...Node) *Box {
	var validChildren []Node
	for _, c := range children {
		if c != nil {
			validChildren = append(validChildren, c)
		}
	}
	return &Box{Style: style, Children: validChildren}
}

// GetStyle returns the style of the Box node.
func (b *Box) GetChildren() []Node {
	return b.Children
}

func (b *Box) GetStyle() Style {
	return b.Style
}

// IsFocusable indicates that a node can receive focus.
func (b *Box) IsFocusable() bool {
	return b.Style.Focusable
}

// Layout calculates the layout for the Box node.
func (n *Box) Layout(x, y int, c Constraints) LayoutResult {
	borderSize := 0
	if n.Style.Border {
		borderSize = 1
	}

	pad := n.Style.Padding
	margin := n.Style.Margin

	boxX := x + margin.Left
	boxY := y + margin.Top

	res := LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    0,
		H:    0,
	}

	// Children start after border and padding
	curX := boxX + borderSize + pad.Left
	curY := boxY + borderSize + pad.Top

	contentW := 0
	contentH := 0

	childConstraints := Constraints{
		MaxW: c.MaxW - (borderSize * 2) - pad.Left - pad.Right,
		MaxH: c.MaxH - (borderSize * 2) - pad.Top - pad.Bottom,
	}
	if childConstraints.MaxW < 0 {
		childConstraints.MaxW = 0
	}
	if childConstraints.MaxH < 0 {
		childConstraints.MaxH = 0
	}

	for _, child := range n.Children {
		childMargin := child.GetStyle().Margin

		cRes := Layout(child, curX, curY, childConstraints)
		res.Children = append(res.Children, cRes)

		if n.Style.FlexDirection == "row" {
			curX += childMargin.Left + cRes.W + childMargin.Right
			contentW += childMargin.Left + cRes.W + childMargin.Right
			if cRes.H+childMargin.Top+childMargin.Bottom > contentH {
				contentH = cRes.H + childMargin.Top + childMargin.Bottom
			}
			childConstraints.MaxW -= (childMargin.Left + cRes.W + childMargin.Right)
			if childConstraints.MaxW < 0 {
				childConstraints.MaxW = 0
			}
		} else { // default to column
			curY += childMargin.Top + cRes.H + childMargin.Bottom
			contentH += childMargin.Top + cRes.H + childMargin.Bottom
			if cRes.W+childMargin.Left+childMargin.Right > contentW {
				contentW = childMargin.Left + cRes.W + childMargin.Right
			}
			childConstraints.MaxH -= (childMargin.Top + cRes.H + childMargin.Bottom)
			if childConstraints.MaxH < 0 {
				childConstraints.MaxH = 0
			}
		}
	}

	// Total size of this box (excluding its own margin)
	res.W = contentW + (borderSize * 2) + pad.Left + pad.Right
	res.H = contentH + (borderSize * 2) + pad.Top + pad.Bottom

	if n.Style.Width > 0 {
		res.W = n.Style.Width
	}
	if n.Style.Height > 0 {
		res.H = n.Style.Height
	}

	// If we are centering, we take up all available space!
	if n.Style.JustifyContent == "center" {
		if n.Style.FlexDirection == "row" {
			if c.MaxW > res.W {
				res.W = c.MaxW
			}
		} else {
			if c.MaxH > res.H {
				res.H = c.MaxH
			}
		}
	}

	if n.Style.FillWidth {
		if c.MaxW > res.W {
			res.W = c.MaxW
		}
	}
	if n.Style.FillHeight {
		if c.MaxH > res.H {
			res.H = c.MaxH
		}
	}

	// Flexbox alignment
	if n.Style.JustifyContent == "center" {
		if n.Style.FlexDirection == "row" {
			remainingW := res.W - (borderSize * 2) - pad.Left - pad.Right - contentW
			if remainingW > 0 {
				shift := remainingW / 2
				for i := range res.Children {
					res.Children[i].X += shift
				}
			}
		} else { // column
			remainingH := res.H - (borderSize * 2) - pad.Top - pad.Bottom - contentH
			if remainingH > 0 {
				shift := remainingH / 2
				for i := range res.Children {
					res.Children[i].Y += shift
				}
			}
		}
	}

	return res
}

// Render draws the Box node to the grid.
func (n *Box) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := n.Style.ID != "" && n.Style.ID == focusedID
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	borderStyle := style
	if focused {
		focusColor := tcell.ColorYellow
		if n.Style.FocusColor != tcell.ColorReset {
			focusColor = n.Style.FocusColor
		}
		focusBg := n.Style.Background
		if n.Style.FocusBackground != tcell.ColorReset {
			focusBg = n.Style.FocusBackground
		}
		borderStyle = tcell.StyleDefault.Foreground(focusColor).Background(focusBg)
	}
	if n.Style.Background != tcell.ColorReset {
		for r := 0; r < layout.H; r++ {
			for c := 0; c < layout.W; c++ {
				grid.SetContent(layout.X+c, layout.Y+r, ' ', style)
			}
		}
	}
	if n.Style.Border {
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, n.Style.Title, borderStyle)
	}
	for _, child := range layout.Children {
		Render(grid, child, focusedID, componentStates)
	}
}
