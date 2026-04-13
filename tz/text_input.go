package tz

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// TextInput is a node that allows text input.
type TextInput struct {
	Style       Style
	Value       string
	Placeholder string   // hint text shown when Value is empty
	Mask        rune     // if non-zero, render this rune instead of actual characters (e.g. '*' for passwords)
	Disabled    bool     // if true, input is read-only and not focusable
	MaxLen      int      // if > 0, caps the number of runes that can be entered
	OnChange    func(string)
	OnSubmit    func(string)
	Cursor      *int // Optional controlled cursor position
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
	hookKey := fmt.Sprintf("hook-%d", ctx.hookIndex-1)

	if style.ID == "" {
		// No explicit ID: use the hook key so state lookups match.
		style.ID = hookKey
	} else if state, exists := ctx.app.componentStates[hookKey]; exists {
		// Explicit ID provided: alias the state under it so that the app's
		// focus/cursor management (which keys on Style.ID) can find it.
		ctx.app.componentStates[style.ID] = state
	}

	return &TextInput{
		Style:    style,
		Value:    value,
		OnChange: onChange,
	}
}

// node implements the Node interface.

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
	focused := !n.Disabled && n.Style.ID != "" && n.Style.ID == focusedID

	var style tcell.Style
	var borderStyle tcell.Style
	if n.Disabled {
		style = tcell.StyleDefault.Foreground(tcell.ColorGray).Background(n.Style.Background).Attributes(tcell.AttrDim)
		borderStyle = style
	} else if focused {
		if n.Style.Border {
			// Border mode: border glows with focus color, text stays normal
			borderColor := n.Style.FocusColor
			if borderColor == 0 {
				borderColor = tcell.ColorYellow
			}
			borderStyle = tcell.StyleDefault.Foreground(borderColor).Background(n.Style.Background)
			style = tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background).Attributes(n.Style.TextAttrs)
		} else {
			// No-border mode: entire content area is highlighted
			focusFg := n.Style.FocusColor
			if focusFg == 0 {
				focusFg = tcell.ColorBlack
			}
			focusBg := n.Style.FocusBackground
			if focusBg == 0 {
				focusBg = tcell.ColorYellow
			}
			style = tcell.StyleDefault.Foreground(focusFg).Background(focusBg).Attributes(n.Style.TextAttrs)
			borderStyle = style
		}
	} else {
		style = tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background).Attributes(n.Style.TextAttrs)
		borderStyle = style
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, n.Style.Title, borderStyle)
	}

	contentX := layout.X + n.Style.Padding.Left + borderOffset
	contentY := layout.Y + n.Style.Padding.Top + borderOffset
	contentW := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
	contentH := layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
	if contentW <= 0 || contentH <= 0 {
		return
	}

	if n.Style.Multiline {
		n.renderMultiline(grid, componentStates, contentX, contentY, contentW, contentH, borderOffset, style)
	} else {
		n.renderSingleLine(grid, componentStates, contentX, contentY, contentW, borderOffset, style)
	}
}

func (n *TextInput) renderSingleLine(grid *Grid, componentStates map[string]any, contentX, contentY, contentW, borderOffset int, style tcell.Style) {
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

	if n.Value == "" && n.Placeholder != "" {
		ph := n.Placeholder
		if len(ph) > contentW {
			ph = ph[:contentW]
		}
		_, bg, _ := style.Decompose()
		placeholderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(bg).Attributes(tcell.AttrDim)
		drawText(grid, contentX, contentY, ph, placeholderStyle)
		return
	}

	val := n.Value
	hasOverflowLeft := scrollOffset > 0
	if scrollOffset < len(val) {
		val = val[scrollOffset:]
	} else {
		val = ""
	}
	hasOverflowRight := len(val) > contentW
	if len(val) > contentW {
		val = val[:contentW]
	}
	if n.Mask != 0 {
		val = strings.Repeat(string(n.Mask), len([]rune(val)))
	}

	drawText(grid, contentX, contentY, val, style)

	// Scroll indicators: rendered on the border characters (not content cells).
	// Only shown when a border is present to avoid corrupting visible text.
	if borderOffset > 0 {
		_, bg, _ := style.Decompose()
		indicatorStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(bg).Attributes(tcell.AttrBold)
		if hasOverflowLeft {
			drawText(grid, contentX-1, contentY, "<", indicatorStyle)
		}
		if hasOverflowRight {
			drawText(grid, contentX+contentW, contentY, ">", indicatorStyle)
		}
	}
}

func (n *TextInput) renderMultiline(grid *Grid, componentStates map[string]any, contentX, contentY, contentW, contentH, borderOffset int, style tcell.Style) {
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

	if n.Value == "" && n.Placeholder != "" {
		ph := n.Placeholder
		if len(ph) > contentW {
			ph = ph[:contentW]
		}
		_, bg, _ := style.Decompose()
		placeholderStyle := tcell.StyleDefault.Foreground(tcell.ColorGray).Background(bg).Attributes(tcell.AttrDim)
		drawText(grid, contentX, contentY, ph, placeholderStyle)
		return
	}

	lines := strings.Split(n.Value, "\n")
	for i := 0; i < contentH; i++ {
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
		if len(val) > contentW {
			val = val[:contentW]
		}
		if n.Mask != 0 {
			val = strings.Repeat(string(n.Mask), len([]rune(val)))
		}
		drawText(grid, contentX, contentY+i, val, style)
	}

	// Vertical scroll indicators rendered on the top/bottom border characters.
	// Only shown when a border is present to avoid corrupting visible text.
	if borderOffset > 0 {
		hasOverflowUp := vScrollOffset > 0
		hasOverflowDown := vScrollOffset+contentH < len(lines)
		_, bg, _ := style.Decompose()
		indicatorStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(bg).Attributes(tcell.AttrBold)
		if hasOverflowUp {
			drawText(grid, contentX+contentW-1, contentY-1, "↑", indicatorStyle)
		}
		if hasOverflowDown {
			drawText(grid, contentX+contentW-1, contentY+contentH, "↓", indicatorStyle)
		}
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

// UpdateScrollOffset adjusts scroll offsets so the cursor stays within the
// visible area. Called by the framework after layout, before rendering.
func (n *TextInput) UpdateScrollOffset(layout LayoutResult, state any) {
	s, ok := state.(*TextInputState)
	if !ok {
		return
	}

	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
	}

	w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
	if w > 0 && !n.Style.Multiline {
		if s.cursorOffset < s.scrollOffset {
			s.scrollOffset = s.cursorOffset
		}
		if s.cursorOffset > s.scrollOffset+w {
			s.scrollOffset = s.cursorOffset - w
		}
	}

	if n.Style.Multiline {
		line, col := offsetToLineCol(n.Value, s.cursorOffset)
		h := layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
		if h > 0 {
			if line < s.vScrollOffset {
				s.vScrollOffset = line
			}
			if line >= s.vScrollOffset+h {
				s.vScrollOffset = line - h + 1
			}
		}
		if w > 0 {
			if col < s.scrollOffset {
				s.scrollOffset = col
			}
			if col >= s.scrollOffset+w {
				s.scrollOffset = col - w + 1
			}
		}
	}
}

// GetCursorPosition returns the screen coordinates for the terminal cursor.
// Called by the framework after rendering.
func (n *TextInput) GetCursorPosition(layout LayoutResult, state any) (x, y int, show bool) {
	borderOffset := 0
	if n.Style.Border {
		borderOffset = 1
	}

	s, ok := state.(*TextInputState)
	if !ok {
		return layout.X + len(n.Value) + borderOffset, layout.Y + borderOffset, true
	}

	if n.Cursor != nil {
		s.cursorOffset = *n.Cursor
	}

	if n.Style.Multiline {
		line, col := offsetToLineCol(n.Value, s.cursorOffset)
		return layout.X + n.Style.Padding.Left + col - s.scrollOffset + borderOffset,
			layout.Y + n.Style.Padding.Top + line + borderOffset - s.vScrollOffset,
			true
	}

	visualOffset := s.cursorOffset - s.scrollOffset
	return layout.X + n.Style.Padding.Left + visualOffset + borderOffset,
		layout.Y + n.Style.Padding.Top + borderOffset,
		true
}

func (n *TextInput) DefaultState() any {
	return &TextInputState{cursorOffset: len(n.Value)}
}

// IsFocusable indicates that a node can receive focus.
func (n *TextInput) IsFocusable() bool {
	return n.Style.Focusable && !n.Disabled
}

func (n *TextInput) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
	if n.Disabled {
		return false
	}

	s, ok := state.(*TextInputState)
	if !ok {
		return false
	}

	// Mouse click: position cursor at the clicked cell.
	if mev, ok := ev.(MouseEvent); ok {
		if mev.Buttons()&tcell.Button1 != 0 {
			mx, my := mev.Position()
			borderOffset := 0
			if n.Style.Border {
				borderOffset = 1
			}
			clickX := mx - ctx.Layout.X - borderOffset - n.Style.Padding.Left
			clickY := my - ctx.Layout.Y - borderOffset - n.Style.Padding.Top
			if n.Style.Multiline {
				targetLine := s.vScrollOffset + clickY
				targetCol := s.scrollOffset + clickX
				s.cursorOffset = lineColToOffset(n.Value, targetLine, targetCol)
			} else {
				target := s.scrollOffset + clickX
				if target < 0 {
					target = 0
				}
				if target > len(n.Value) {
					target = len(n.Value)
				}
				s.cursorOffset = target
			}
			return true
		}
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
	ctrl := key.Modifiers()&tcell.ModCtrl != 0

	switch key.Key() {
	case tcell.KeyLeft:
		if ctrl {
			s.cursorOffset = prevWordStart(n.Value, s.cursorOffset)
		} else {
			s.cursorOffset--
			if s.cursorOffset < 0 {
				s.cursorOffset = 0
			}
		}
		dirty = true

	case tcell.KeyRight:
		if ctrl {
			s.cursorOffset = nextWordEnd(n.Value, s.cursorOffset)
		} else {
			s.cursorOffset++
			if s.cursorOffset > len(n.Value) {
				s.cursorOffset = len(n.Value)
			}
		}
		dirty = true

	case tcell.KeyBackspace, tcell.KeyBackspace2:
		if s.cursorOffset > 0 {
			newVal := n.Value[:s.cursorOffset-1] + n.Value[s.cursorOffset:]
			n.Value = newVal
			s.cursorOffset--
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		}

	case tcell.KeyDelete:
		if s.cursorOffset < len(n.Value) {
			newVal := n.Value[:s.cursorOffset] + n.Value[s.cursorOffset+1:]
			n.Value = newVal
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		}

	case tcell.KeyCtrlW:
		// Delete from cursor back to the start of the previous word (Unix readline convention).
		if s.cursorOffset > 0 {
			wordStart := prevWordStart(n.Value, s.cursorOffset)
			newVal := n.Value[:wordStart] + n.Value[s.cursorOffset:]
			s.cursorOffset = wordStart
			n.Value = newVal
			dirty = true
			if n.OnChange != nil {
				n.OnChange(newVal)
			}
		}

	case tcell.KeyEnter:
		if n.Style.Multiline {
			if n.MaxLen > 0 && len([]rune(n.Value)) >= n.MaxLen {
				return false
			}
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

	case tcell.KeyUp:
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			if line > 0 {
				s.cursorOffset = lineColToOffset(n.Value, line-1, col)
				dirty = true
			}
		}

	case tcell.KeyDown:
		if n.Style.Multiline {
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			s.cursorOffset = lineColToOffset(n.Value, line+1, col)
			dirty = true
		}

	case tcell.KeyPgUp:
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

	case tcell.KeyPgDn:
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

	case tcell.KeyHome:
		if n.Style.Multiline {
			// First press: go to start of current line. If already there, go to absolute start.
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			if col > 0 {
				s.cursorOffset = lineColToOffset(n.Value, line, 0)
			} else {
				s.cursorOffset = 0
			}
		} else {
			s.cursorOffset = 0
		}
		dirty = true

	case tcell.KeyEnd:
		if n.Style.Multiline {
			// First press: go to end of current line. If already there, go to absolute end.
			line, col := offsetToLineCol(n.Value, s.cursorOffset)
			lines := strings.Split(n.Value, "\n")
			lineLen := len(lines[line])
			if col < lineLen {
				s.cursorOffset = lineColToOffset(n.Value, line, lineLen)
			} else {
				s.cursorOffset = len(n.Value)
			}
		} else {
			s.cursorOffset = len(n.Value)
		}
		dirty = true

	case tcell.KeyRune:
		if n.MaxLen > 0 && len([]rune(n.Value)) >= n.MaxLen {
			return false
		}
		newVal := n.Value[:s.cursorOffset] + string(key.Rune()) + n.Value[s.cursorOffset:]
		n.Value = newVal
		s.cursorOffset++
		dirty = true
		if n.OnChange != nil {
			n.OnChange(newVal)
		}
	}

	// Adjust horizontal scroll offset to keep cursor visible (single-line only).
	// For multiline, UpdateScrollOffset handles both axes using the correct
	// column position rather than the raw byte offset.
	if !n.Style.Multiline {
		borderOffset := 0
		if n.Style.Border {
			borderOffset = 1
		}
		w := ctx.Layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
		if w > 0 {
			if s.cursorOffset-s.scrollOffset >= w {
				s.scrollOffset = s.cursorOffset - w + 1
				dirty = true
			}
			if s.cursorOffset < s.scrollOffset {
				s.scrollOffset = s.cursorOffset
				dirty = true
			}
		}
	}

	return dirty
}

// prevWordStart returns the offset of the start of the word before the cursor.
func prevWordStart(s string, offset int) int {
	if offset > len(s) {
		offset = len(s)
	}
	for offset > 0 && isWordSep(rune(s[offset-1])) {
		offset--
	}
	for offset > 0 && !isWordSep(rune(s[offset-1])) {
		offset--
	}
	return offset
}

// nextWordEnd returns the offset just past the end of the next word.
func nextWordEnd(s string, offset int) int {
	for offset < len(s) && isWordSep(rune(s[offset])) {
		offset++
	}
	for offset < len(s) && !isWordSep(rune(s[offset])) {
		offset++
	}
	return offset
}

// isWordSep reports whether r is a word separator for navigation purposes.
func isWordSep(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '.', ',', '(', ')', '[', ']', '{', '}', '/', '\\', ':', ';', '\'', '"', '!', '?':
		return true
	}
	return false
}

// offsetToLineCol converts a flat byte offset into a (line, col) pair within
// a multiline string.
func offsetToLineCol(text string, offset int) (int, int) {
	lines := strings.Split(text, "\n")
	currentOffset := 0
	for lineIdx, line := range lines {
		if offset <= currentOffset+len(line) {
			return lineIdx, offset - currentOffset
		}
		currentOffset += len(line) + 1 // +1 for \n
	}
	if len(lines) == 0 {
		return 0, 0
	}
	return len(lines) - 1, len(lines[len(lines)-1])
}

// lineColToOffset converts a (line, col) pair into a flat byte offset within
// a multiline string.
func lineColToOffset(text string, line, col int) int {
	lines := strings.Split(text, "\n")
	if line < 0 {
		return 0
	}
	if line >= len(lines) {
		return len(text)
	}
	offset := 0
	for i := 0; i < line; i++ {
		offset += len(lines[i]) + 1
	}
	c := col
	if c > len(lines[line]) {
		c = len(lines[line])
	}
	return offset + c
}
