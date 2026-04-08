package splotch

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// TextInput is a node that allows text input.
type TextInput struct {
	Style    Style
	Value    string
	OnChange func(string)
	Cursor   *int // Optional controlled cursor position
}

// TextInputState tracks the state for TextInput.
type TextInputState struct {
	cursorOffset  int
	scrollOffset  int
	vScrollOffset int
}

// NewTextInput creates a new TextInput node.
func NewTextInput(ctx *RenderContext, style Style, value string, onChange func(string)) *TextInput {
	_, _ = UseState[*TextInputState](ctx, &TextInputState{cursorOffset: len(value)})

	if style.ID == "" {
		style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
	}

	return &TextInput{
		Style:    style,
		Value:    value,
		OnChange: onChange,
	}
}

// node implements the Node interface.
func (t *TextInput) node() {}

// Layout calculates the layout for the TextInput node.
func (n *TextInput) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin

	boxX := x + margin.Left
	boxY := y + margin.Top

	w := 0
	h := 1
	if n.Style.Multiline {
		lines := strings.Split(n.Value, "\n")
		for _, line := range lines {
			if len(line) > w {
				w = len(line)
			}
		}
		h = len(lines)
	} else {
		w = len(n.Value)
	}
	
	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	layoutH := h + pad.Top + pad.Bottom + borderSize
	if n.Style.MaxHeight > 0 && layoutH > n.Style.MaxHeight {
		layoutH = n.Style.MaxHeight
	}

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}
}

// Render draws the TextInput node to the grid.
func (n *TextInput) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	focused := false
	if n.Style.ID != "" && n.Style.ID == focusedID {
		focused = true
	}
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
	borderStyle := style
	if focused {
		borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
		style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
	}
	
	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, borderStyle)
	}

	if n.Style.Multiline {
		vScrollOffset := 0
		if n.Style.ID != "" && componentStates != nil {
			if stateObj, ok := componentStates[n.Style.ID]; ok {
				state := stateObj.(*TextInputState)
				vScrollOffset = state.vScrollOffset
			}
		}

		lines := strings.Split(n.Value, "\n")
		w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
		visibleHeight := layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
		
		for i := 0; i < visibleHeight; i++ {
			lineIdx := i + vScrollOffset
			if lineIdx >= len(lines) {
				break
			}
			val := lines[lineIdx]
			if len(val) > w && w > 0 {
				val = val[:w]
			}
			drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset+i, val, style)
		}
	} else {
		val := n.Value
		scrollOffset := 0
		if n.Style.ID != "" && componentStates != nil {
			if stateObj, ok := componentStates[n.Style.ID]; ok {
				state := stateObj.(*TextInputState)
				scrollOffset = state.scrollOffset
			}
		}
		
		w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
		if scrollOffset < len(val) {
			val = val[scrollOffset:]
		} else {
			val = ""
		}
		if len(val) > w && w > 0 {
			val = val[:w]
		}
		
		drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, val, style)
	}
}

// GetStyle returns the style of the TextInput node.
func (n *TextInput) GetStyle() Style {
	return n.Style
}

// GetCursorCoords returns the cursor coordinates relative to the input content area.
func (n *TextInput) GetCursorCoords(ctx *RenderContext) (int, int) {
	stateObj, ok := ctx.app.componentStates[n.Style.ID]
	if !ok {
		return 0, 0
	}
	state := stateObj.(*TextInputState)

	if n.Style.Multiline {
		line, col := offsetToLineCol(n.Value, state.cursorOffset)
		return col, line
	}

	visualOffset := state.cursorOffset - state.scrollOffset
	return visualOffset, 0
}
