package splotch

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// App manages the terminal screen and the event loop.
type App struct {
	screen          tcell.Screen
	focusedID       string
	componentStates map[string]any
	previousGrid    *Grid
}

// NewApp creates a new App instance.
func NewApp() (*App, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	return &App{
		screen:          s,
		componentStates: make(map[string]any),
	}, nil
}

// Run starts the application loop.
// It takes a function that returns the root Node of the UI,
// and a function to handle events and update state.
func (a *App) Run(renderFn func() Node, updateFn func(tcell.Event)) error {
	if err := a.screen.Init(); err != nil {
		return err
	}
	defer a.screen.Fini()

	a.screen.EnableMouse()

	// Set default style
	a.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))

	// Start animation ticker
	go func() {
		for {
			time.Sleep(100 * time.Millisecond) // 10 FPS
			a.screen.PostEvent(&EventTick{t: time.Now()})
		}
	}()

	for {
		// 1. Get the current UI tree
		root := renderFn()

		focusableIDs := findFocusableIDs(root)
		if a.focusedID == "" && len(focusableIDs) > 0 {
			a.focusedID = focusableIDs[0]
		}

		// 2. Compute layout
		w, h := a.screen.Size()
		layout := Layout(root, 0, 0, Constraints{MaxW: w, MaxH: h})

		// 3. Render to grid
		grid := NewGrid(w, h)
		Render(grid, layout, a.focusedID, a.componentStates)

		// 4. Diff and update screen
		if a.previousGrid == nil || a.previousGrid.W != w || a.previousGrid.H != h {
			a.screen.Clear()
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					cell := grid.Cells[y][x]
					a.screen.SetContent(x, y, cell.Rune, nil, cell.Style)
				}
			}
		} else {
			for y := 0; y < h; y++ {
				for x := 0; x < w; x++ {
					newCell := grid.Cells[y][x]
					oldCell := a.previousGrid.Cells[y][x]
					if newCell.Rune != oldCell.Rune || newCell.Style != oldCell.Style {
						a.screen.SetContent(x, y, newCell.Rune, nil, newCell.Style)
					}
				}
			}
		}

		a.previousGrid = grid

		// Handle cursor for focused TextInput
		if a.focusedID != "" {
			focusedNode := findNodeByID(root, a.focusedID)
			if input, ok := focusedNode.(*TextInput); ok {
				res := findLayoutResultByID(layout, a.focusedID)
				if res != nil {
					stateObj, ok := a.componentStates[a.focusedID]
					if ok {
						state := stateObj.(*TextInputState)
						borderOffset := 0
						if input.Style.Border {
							borderOffset = 1
						}
						if input.Style.Multiline {
							line, col := offsetToLineCol(input.Value, state.cursorOffset)
							a.screen.ShowCursor(res.X+input.Style.Padding.Left+col+borderOffset, res.Y+input.Style.Padding.Top+line+borderOffset-state.vScrollOffset)
						} else {
							visualOffset := state.cursorOffset - state.scrollOffset
							a.screen.ShowCursor(res.X+input.Style.Padding.Left+visualOffset+borderOffset, res.Y+input.Style.Padding.Top+borderOffset)
						}
					} else {
						borderOffset := 0
						if input.Style.Border {
							borderOffset = 1
						}
						a.screen.ShowCursor(res.X+len(input.Value)+borderOffset, res.Y+borderOffset)
					}
				}
			} else {
				a.screen.HideCursor()
			}
		} else {
			a.screen.HideCursor()
		}

		a.screen.Show()

		// Event loop
		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
				return nil
			}
			if ev.Key() == tcell.KeyTab {
				a.focusedID = nextFocus(a.focusedID, focusableIDs)
			}
			if ev.Key() == tcell.KeyBacktab {
				a.focusedID = prevFocus(a.focusedID, focusableIDs)
			}
			if a.focusedID != "" {
				focusedNode := findNodeByID(root, a.focusedID)
				if input, ok := focusedNode.(*TextInput); ok {
					stateObj, ok := a.componentStates[a.focusedID]
					var state *TextInputState
					if !ok {
						state = &TextInputState{cursorOffset: len(input.Value)}
						a.componentStates[a.focusedID] = state
					} else {
						state = stateObj.(*TextInputState)
					}

					if state.cursorOffset > len(input.Value) {
						state.cursorOffset = len(input.Value)
					}

					if ev.Key() == tcell.KeyLeft {
						if state.cursorOffset > 0 {
							state.cursorOffset--
						}
					} else if ev.Key() == tcell.KeyRight {
						if state.cursorOffset < len(input.Value) {
							state.cursorOffset++
						}
					} else if ev.Key() == tcell.KeyUp && input.Style.Multiline {
						line, col := offsetToLineCol(input.Value, state.cursorOffset)
						if line > 0 {
							state.cursorOffset = lineColToOffset(input.Value, line-1, col)
						}
					} else if ev.Key() == tcell.KeyDown && input.Style.Multiline {
						line, col := offsetToLineCol(input.Value, state.cursorOffset)
						lines := strings.Split(input.Value, "\n")
						if line < len(lines)-1 {
							state.cursorOffset = lineColToOffset(input.Value, line+1, col)
						}
					} else if ev.Key() == tcell.KeyEnter && input.Style.Multiline {
						newVal := input.Value[:state.cursorOffset] + "\n" + input.Value[state.cursorOffset:]
						state.cursorOffset++
						if input.OnChange != nil {
							input.OnChange(newVal)
						}
					} else if ev.Key() == tcell.KeyRune {
						newVal := input.Value[:state.cursorOffset] + string(ev.Rune()) + input.Value[state.cursorOffset:]
						state.cursorOffset++
						if input.OnChange != nil {
							input.OnChange(newVal)
						}
					} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
						if state.cursorOffset > 0 {
							newVal := input.Value[:state.cursorOffset-1] + input.Value[state.cursorOffset:]
							state.cursorOffset--
							if input.OnChange != nil {
								input.OnChange(newVal)
							}
						}
					} else if ev.Key() == tcell.KeyDelete {
						if state.cursorOffset < len(input.Value) {
							newVal := input.Value[:state.cursorOffset] + input.Value[state.cursorOffset+1:]
							if input.OnChange != nil {
								input.OnChange(newVal)
							}
						}
					}

					// Update scroll offset
					res := findLayoutResultByID(layout, a.focusedID)
					if res != nil {
						borderOffset := 0
						if input.Style.Border {
							borderOffset = 1
						}

						w := res.W - input.Style.Padding.Left - input.Style.Padding.Right - borderOffset*2
						if w > 0 && !input.Style.Multiline {
							if state.cursorOffset < state.scrollOffset {
								state.scrollOffset = state.cursorOffset
							}
							if state.cursorOffset > state.scrollOffset+w {
								state.scrollOffset = state.cursorOffset - w
							}
						}

						if input.Style.Multiline {
							line, _ := offsetToLineCol(input.Value, state.cursorOffset)
							h := res.H - input.Style.Padding.Top - input.Style.Padding.Bottom - borderOffset*2
							if h > 0 {
								if line < state.vScrollOffset {
									state.vScrollOffset = line
								}
								if line >= state.vScrollOffset+h {
									state.vScrollOffset = line - h + 1
								}
							}
						}
					}
				} else if btn, ok := focusedNode.(*Button); ok {
					if ev.Key() == tcell.KeyEnter {
						if btn.OnClick != nil {
							btn.OnClick()
						}
					}
				} else if cb, ok := focusedNode.(*Checkbox); ok {
					if ev.Key() == tcell.KeyEnter || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
						cb.Checked = !cb.Checked
						if cb.OnChange != nil {
							cb.OnChange(cb.Checked)
						}
					}
				} else if rb, ok := focusedNode.(*RadioButton); ok {
					if ev.Key() == tcell.KeyEnter || (ev.Key() == tcell.KeyRune && ev.Rune() == ' ') {
						if rb.OnChange != nil {
							rb.OnChange(rb.Value)
						}
					}
				}
			}
		case *EventTick:
			// Continuous rendering for animations
		case *tcell.EventMouse:
			mx, my := ev.Position()
			if ev.Buttons()&tcell.Button1 != 0 {
				clickedNode := findNodeAt(layout, mx, my)
				if clickedNode != nil {
					var nodeStyle Style
					switch n := clickedNode.(type) {
					case *Text:
						nodeStyle = n.Style
					case *TextInput:
						nodeStyle = n.Style
					case *Button:
						nodeStyle = n.Style
					case *Checkbox:
						nodeStyle = n.Style
					case *RadioButton:
						nodeStyle = n.Style
					case *Spinner:
						nodeStyle = n.Style
					case *ProgressBar:
						nodeStyle = n.Style
					}
					if nodeStyle.Focusable && nodeStyle.ID != "" {
						a.focusedID = nodeStyle.ID
					}

					if btn, ok := clickedNode.(*Button); ok {
						if btn.OnClick != nil {
							btn.OnClick()
						}
					}
					if cb, ok := clickedNode.(*Checkbox); ok {
						cb.Checked = !cb.Checked
						if cb.OnChange != nil {
							cb.OnChange(cb.Checked)
						}
					}
					if rb, ok := clickedNode.(*RadioButton); ok {
						if rb.OnChange != nil {
							rb.OnChange(rb.Value)
						}
					}
				}
			}
		case *tcell.EventResize:
			a.screen.Sync()
		}

		// Call user update function to handle state changes
		if updateFn != nil {
			updateFn(ev)
		}
	}
}

func findFocusableIDs(node Node) []string {
	var ids []string
	switch n := node.(type) {
	case *Text:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *TextInput:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *Button:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *Checkbox:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *RadioButton:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *Box:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
		for _, child := range n.Children {
			ids = append(ids, findFocusableIDs(child)...)
		}
	}
	return ids
}

func nextFocus(current string, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	if current == "" {
		return ids[0]
	}
	for i, id := range ids {
		if id == current {
			return ids[(i+1)%len(ids)]
		}
	}
	return ids[0]
}

func prevFocus(current string, ids []string) string {
	if len(ids) == 0 {
		return ""
	}
	if current == "" {
		return ids[len(ids)-1]
	}
	for i, id := range ids {
		if id == current {
			return ids[(i-1+len(ids))%len(ids)]
		}
	}
	return ids[len(ids)-1]
}

func findNodeByID(node Node, id string) Node {
	switch n := node.(type) {
	case *Text:
		if n.Style.ID == id {
			return n
		}
	case *Checkbox:
		if n.Style.ID == id {
			return n
		}
	case *RadioButton:
		if n.Style.ID == id {
			return n
		}
	case *Box:
		if n.Style.ID == id {
			return n
		}
		for _, child := range n.Children {
			if found := findNodeByID(child, id); found != nil {
				return found
			}
		}
	case *TextInput:
		if n.Style.ID == id {
			return n
		}
	case *Button:
		if n.Style.ID == id {
			return n
		}
	}
	return nil
}

func findLayoutResultByID(res LayoutResult, id string) *LayoutResult {
	switch n := res.Node.(type) {
	case *Text:
		if n.Style.ID == id {
			return &res
		}
	case *Box:
		if n.Style.ID == id {
			return &res
		}
		for _, child := range res.Children {
			if found := findLayoutResultByID(child, id); found != nil {
				return found
			}
		}
	case *TextInput:
		if n.Style.ID == id {
			return &res
		}
	case *Button:
		if n.Style.ID == id {
			return &res
		}
	}
	return nil
}

type TextInputState struct {
	cursorOffset  int
	scrollOffset  int
	vScrollOffset int
}

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

type EventTick struct {
	t time.Time
}

func (e *EventTick) When() time.Time {
	return e.t
}

func findNodeAt(res LayoutResult, x, y int) Node {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		// Check children first for nested components
		for _, child := range res.Children {
			if n := findNodeAt(child, x, y); n != nil {
				return n
			}
		}
		return res.Node
	}
	return nil
}
