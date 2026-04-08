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
	OnSubmit func(string)
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

	finalW := w + pad.Left + pad.Right + borderSize
	if n.Style.FillWidth && c.MaxW > finalW {
		finalW = c.MaxW
	}

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    finalW,
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
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	if n.Style.Multiline {
		vScrollOffset := 0
		scrollOffset := 0
		if n.Style.ID != "" && componentStates != nil {
			if stateObj, ok := componentStates[n.Style.ID]; ok {
				state := stateObj.(*TextInputState)
				if state.cursorOffset > len(n.Value) {
					state.cursorOffset = len(n.Value)
				}
				vScrollOffset = state.vScrollOffset
				scrollOffset = state.scrollOffset
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
			if scrollOffset < len(val) {
				val = val[scrollOffset:]
			} else {
				val = ""
			}
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
				if state.cursorOffset > len(n.Value) {
					state.cursorOffset = len(n.Value)
				}
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

func (n *TextInput) DefaultState() any {
	return &TextInputState{cursorOffset: len(n.Value)}
}

func (n *TextInput) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	s, ok := state.(*TextInputState)
	if !ok {
		return false
	}

	key, ok := ev.(*tcell.EventKey)
	if !ok {
		return false
	}

	if ctx.PopupOpen {
		if key.Key() == tcell.KeyEnter || key.Key() == tcell.KeyUp || key.Key() == tcell.KeyDown {
			return false
		}
	}

	dirty := false

	if key.Key() == tcell.KeyLeft {
		s.cursorOffset--
		if s.cursorOffset < 0 {
			s.cursorOffset = 0
		}
		dirty = true
	} else if key.Key() == tcell.KeyRight {
		s.cursorOffset++
		if s.cursorOffset > len(n.Value) {
			s.cursorOffset = len(n.Value)
		}
		dirty = true
	} else if key.Key() == tcell.KeyBackspace || key.Key() == tcell.KeyBackspace2 {
		if s.cursorOffset > 0 {
			newVal := n.Value[:s.cursorOffset-1] + n.Value[s.cursorOffset:]
			n.Value = newVal
			s.cursorOffset--
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		}
	} else if key.Key() == tcell.KeyDelete {
		if s.cursorOffset < len(n.Value) {
			newVal := n.Value[:s.cursorOffset] + n.Value[s.cursorOffset+1:]
			n.Value = newVal
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		}
	} else if key.Key() == tcell.KeyEnter {
		if n.Style.Multiline {
			newVal := n.Value[:s.cursorOffset] + "\n" + n.Value[s.cursorOffset:]
			n.Value = newVal
			s.cursorOffset++
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		} else {
			if n.OnSubmit != nil {
				n.OnSubmit(n.Value)
			}
		}
	} else if key.Key() == tcell.KeyUp {
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			if line > 0 {
				s.cursorOffset = lineColToOffset(n.Value, line-1, col)
				dirty = true
			}
		}
	} else if key.Key() == tcell.KeyDown {
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			s.cursorOffset = lineColToOffset(n.Value, line+1, col)
			dirty = true
		}
	} else if key.Key() == tcell.KeyPgUp {
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			borderOffset := 0
			if n.Style.Border {
				borderOffset = 1
			}
			h := ctx.Layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
			if h > 0 {
				newLine := line - h
				if newLine < 0 {
					newLine = 0
				}
				s.cursorOffset = lineColToOffset(n.Value, newLine, col)
				dirty = true
			}
		}
	} else if key.Key() == tcell.KeyPgDn {
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			borderOffset := 0
			if n.Style.Border {
				borderOffset = 1
			}
			h := ctx.Layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
			if h > 0 {
				lines := strings.Split(n.Value, "\n")
				newLine := line + h
				if newLine >= len(lines) {
					newLine = len(lines) - 1
				}
				s.cursorOffset = lineColToOffset(n.Value, newLine, col)
				dirty = true
			}
		}
	} else if key.Key() == tcell.KeyHome {
		s.cursorOffset = 0
		dirty = true
	} else if key.Key() == tcell.KeyEnd {
		s.cursorOffset = len(n.Value)
		dirty = true
	} else if key.Key() == tcell.KeyRune {
		newVal := n.Value[:s.cursorOffset] + string(key.Rune()) + n.Value[s.cursorOffset:]
		n.Value = newVal
		s.cursorOffset++
		dirty = true
		if n.OnChange != nil {
			n.OnChange(newVal)
		}
	}

	return dirty
}
