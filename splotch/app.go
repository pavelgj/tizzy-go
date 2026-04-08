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

		// Render dropdown overlays
		for id, stateObj := range a.componentStates {
			if state, ok := stateObj.(*DropdownState); ok && state.Open {
				res := findLayoutResultByID(layout, id)
				dropdownNode := findNodeByID(root, id)
				if res != nil && dropdownNode != nil {
					if drp, ok := dropdownNode.(*Dropdown); ok {
						listY := res.Y + res.H
						listW := res.W
						maxH := drp.MaxListHeight
						if maxH <= 0 {
							maxH = 5 // Default limit
						}
						if maxH > len(drp.Options) {
							maxH = len(drp.Options)
						}
						listH := maxH
						
						style := tcell.StyleDefault.Foreground(drp.Style.Color).Background(drp.Style.Background)
						
						for y := 0; y < listH; y++ {
							for x := 0; x < listW; x++ {
								if listY+y < h && res.X+x < w {
									grid.SetContent(res.X+x, listY+y, ' ', style)
								}
							}
						}
						
						for i := 0; i < listH; i++ {
							optIdx := i + state.ScrollOffset
							if optIdx >= len(drp.Options) {
								break
							}
							opt := drp.Options[optIdx]
							optStyle := style
							if optIdx == state.FocusedIndex {
								optStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
							}
							
							curX := res.X + 1
							for _, r := range opt {
								if listY+i < h && curX < w {
									grid.SetContent(curX, listY+i, r, optStyle)
									curX++
								}
							}
							
							for j := len(opt) + 1; j < listW; j++ {
								if listY+i < h && res.X+j < w {
									grid.SetContent(res.X+j, listY+i, ' ', optStyle)
								}
							}
						}
					}
				}
			}
		}

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
			if a.handleKeyEvent(ev, root, layout, focusableIDs) {
				return nil
			}
		case *EventTick:
			// Continuous rendering for animations
		case *tcell.EventMouse:
			mx, my := ev.Position()
			if ev.Buttons()&tcell.Button1 != 0 {
				handled := false
				for id, stateObj := range a.componentStates {
					if state, ok := stateObj.(*DropdownState); ok && state.Open {
						res := findLayoutResultByID(layout, id)
						dropdownNode := findNodeByID(root, id)
						if res != nil && dropdownNode != nil {
							if drp, ok := dropdownNode.(*Dropdown); ok {
								listY := res.Y + res.H
								listW := res.W
								listH := len(drp.Options)
								
								if mx >= res.X && mx < res.X+listW && my >= listY && my < listY+listH {
									clickedIndex := my - listY
									drp.SelectedIndex = clickedIndex
									if drp.OnChange != nil {
										drp.OnChange(clickedIndex)
									}
									state.Open = false
									handled = true
									break
								}
							}
						}
					}
				}

				if !handled {
					path := findNodePathAt(layout, mx, my, a.componentStates)
					if len(path) > 0 {
						clickedNode := path[len(path)-1]
						
						var nodeStyle Style
						var focusableNode Node
						for i := len(path) - 1; i >= 0; i-- {
							n := path[i]
							switch node := n.(type) {
							case *Text:
								nodeStyle = node.Style
							case *TextInput:
								nodeStyle = node.Style
							case *Button:
								nodeStyle = node.Style
							case *Checkbox:
								nodeStyle = node.Style
							case *RadioButton:
								nodeStyle = node.Style
							case *Spinner:
								nodeStyle = node.Style
							case *ProgressBar:
								nodeStyle = node.Style
							case *ScrollView:
								nodeStyle = node.Style
							case *Dropdown:
								nodeStyle = node.Style
							case *Box:
								nodeStyle = node.Style
							}
							if nodeStyle.Focusable && nodeStyle.ID != "" {
								focusableNode = n
								break
							}
						}
						
						if focusableNode != nil {
							a.focusedID = nodeStyle.ID
							a.closeOtherDropdowns(a.focusedID)
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
						if drp, ok := clickedNode.(*Dropdown); ok {
							stateObj, ok := a.componentStates[drp.Style.ID]
							var state *DropdownState
							if !ok {
								state = &DropdownState{}
								a.componentStates[drp.Style.ID] = state
							} else {
								state = stateObj.(*DropdownState)
							}
							state.Open = !state.Open
						}
					}
				}
			} else if ev.Buttons()&tcell.Button4 != 0 { // Wheel Up
				sv := findScrollViewAt(layout, mx, my, a.componentStates)
				if sv != nil && sv.Style.ID != "" {
					stateObj, ok := a.componentStates[sv.Style.ID]
					var state *ScrollViewState
					if !ok {
						state = &ScrollViewState{}
						a.componentStates[sv.Style.ID] = state
					} else {
						state = stateObj.(*ScrollViewState)
					}
					state.ScrollOffset--
					if state.ScrollOffset < 0 {
						state.ScrollOffset = 0
					}
				}
			} else if ev.Buttons()&tcell.Button5 != 0 { // Wheel Down
				sv := findScrollViewAt(layout, mx, my, a.componentStates)
				if sv != nil && sv.Style.ID != "" {
					stateObj, ok := a.componentStates[sv.Style.ID]
					var state *ScrollViewState
					if !ok {
						state = &ScrollViewState{}
						a.componentStates[sv.Style.ID] = state
					} else {
						state = stateObj.(*ScrollViewState)
					}
					state.ScrollOffset++
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
	case *Dropdown:
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
	case *ScrollView:
		if n.Style.ID == id {
			return n
		}
		return findNodeByID(n.Child, id)
	case *TextInput:
		if n.Style.ID == id {
			return n
		}
	case *Button:
		if n.Style.ID == id {
			return n
		}
	case *Dropdown:
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
	case *Dropdown:
		if n.Style.ID == id {
			return &res
		}
	}
	return nil
}

func (a *App) closeOtherDropdowns(keepID string) {
	for id, stateObj := range a.componentStates {
		if id == keepID {
			continue
		}
		if state, ok := stateObj.(*DropdownState); ok {
			state.Open = false
		}
	}
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

func findNodeAt(res LayoutResult, x, y int, componentStates map[string]any) Node {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		scrollOffset := 0
		if sv, ok := res.Node.(*ScrollView); ok {
			if sv.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[sv.Style.ID]; ok {
					state := stateObj.(*ScrollViewState)
					scrollOffset = state.ScrollOffset
				}
			}
		}

		for _, child := range res.Children {
			if n := findNodeAt(child, x, y+scrollOffset, componentStates); n != nil {
				return n
			}
		}
		return res.Node
	}
	return nil
}

func findScrollViewAt(res LayoutResult, x, y int, componentStates map[string]any) *ScrollView {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		scrollOffset := 0
		if sv, ok := res.Node.(*ScrollView); ok {
			if sv.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[sv.Style.ID]; ok {
					state := stateObj.(*ScrollViewState)
					scrollOffset = state.ScrollOffset
				}
			}
			for _, child := range res.Children {
				if svChild := findScrollViewAt(child, x, y+scrollOffset, componentStates); svChild != nil {
					return svChild
				}
			}
			return sv
		}

		for _, child := range res.Children {
			if svChild := findScrollViewAt(child, x, y, componentStates); svChild != nil {
				return svChild
			}
		}
	}
	return nil
}

func findNodePathAt(res LayoutResult, x, y int, componentStates map[string]any) []Node {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		scrollOffset := 0
		if sv, ok := res.Node.(*ScrollView); ok {
			if sv.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[sv.Style.ID]; ok {
					state := stateObj.(*ScrollViewState)
					scrollOffset = state.ScrollOffset
				}
			}
		}

		for _, child := range res.Children {
			if path := findNodePathAt(child, x, y+scrollOffset, componentStates); path != nil {
				return append([]Node{res.Node}, path...)
			}
		}
		return []Node{res.Node}
	}
	return nil
}

func (a *App) handleKeyEvent(ev *tcell.EventKey, root Node, layout LayoutResult, focusableIDs []string) bool {
	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		return true // Request exit
	}
	if ev.Key() == tcell.KeyTab {
		a.focusedID = nextFocus(a.focusedID, focusableIDs)
	}
	if ev.Key() == tcell.KeyBacktab {
		a.focusedID = prevFocus(a.focusedID, focusableIDs)
	}
	if ev.Key() == tcell.KeyTab || ev.Key() == tcell.KeyBacktab {
		a.closeOtherDropdowns(a.focusedID)
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

			if ev.Key() == tcell.KeyLeft {
				state.cursorOffset--
				if state.cursorOffset < 0 {
					state.cursorOffset = 0
				}
			} else if ev.Key() == tcell.KeyRight {
				state.cursorOffset++
				if state.cursorOffset > len(input.Value) {
					state.cursorOffset = len(input.Value)
				}
			} else if ev.Key() == tcell.KeyBackspace || ev.Key() == tcell.KeyBackspace2 {
				if state.cursorOffset > 0 {
					newVal := input.Value[:state.cursorOffset-1] + input.Value[state.cursorOffset:]
					input.Value = newVal
					state.cursorOffset--
					if input.OnChange != nil {
						input.OnChange(newVal)
					}
				}
			} else if ev.Key() == tcell.KeyDelete {
				if state.cursorOffset < len(input.Value) {
					newVal := input.Value[:state.cursorOffset] + input.Value[state.cursorOffset+1:]
					input.Value = newVal
					if input.OnChange != nil {
						input.OnChange(newVal)
					}
				}
			} else if ev.Key() == tcell.KeyEnter {
				if input.Style.Multiline {
					newVal := input.Value[:state.cursorOffset] + "\n" + input.Value[state.cursorOffset:]
					input.Value = newVal
					state.cursorOffset++
					if input.OnChange != nil {
						input.OnChange(newVal)
					}
				}
			} else if ev.Key() == tcell.KeyUp {
				if input.Style.Multiline {
					line, col := offsetToLineCol(input.Value, state.cursorOffset)
					if line > 0 {
						state.cursorOffset = lineColToOffset(input.Value, line-1, col)
					}
				}
			} else if ev.Key() == tcell.KeyDown {
				if input.Style.Multiline {
					line, col := offsetToLineCol(input.Value, state.cursorOffset)
					state.cursorOffset = lineColToOffset(input.Value, line+1, col)
				}
			} else if ev.Key() == tcell.KeyRune {
				newVal := input.Value[:state.cursorOffset] + string(ev.Rune()) + input.Value[state.cursorOffset:]
				input.Value = newVal
				state.cursorOffset++
				if input.OnChange != nil {
					input.OnChange(newVal)
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
		} else if drp, ok := focusedNode.(*Dropdown); ok {
			stateObj, ok := a.componentStates[a.focusedID]
			var state *DropdownState
			if !ok {
				state = &DropdownState{}
				a.componentStates[a.focusedID] = state
			} else {
				state = stateObj.(*DropdownState)
			}

			if ev.Key() == tcell.KeyEnter {
				if state.Open {
					drp.SelectedIndex = state.FocusedIndex
					if drp.OnChange != nil {
						drp.OnChange(state.FocusedIndex)
					}
					state.Open = false
				} else {
					state.Open = true
				}
			} else if ev.Key() == tcell.KeyEscape {
				state.Open = false
			} else if ev.Key() == tcell.KeyUp {
				if state.Open {
					if state.FocusedIndex > 0 {
						state.FocusedIndex--
					}
					
					maxH := drp.MaxListHeight
					if maxH <= 0 {
						maxH = 5
					}
					if maxH > len(drp.Options) {
						maxH = len(drp.Options)
					}
					
					if state.FocusedIndex < state.ScrollOffset {
						state.ScrollOffset = state.FocusedIndex
					}
					if state.FocusedIndex >= state.ScrollOffset + maxH {
						state.ScrollOffset = state.FocusedIndex - maxH + 1
					}
				}
			} else if ev.Key() == tcell.KeyDown {
				if state.Open {
					if state.FocusedIndex < len(drp.Options)-1 {
						state.FocusedIndex++
					}
					
					maxH := drp.MaxListHeight
					if maxH <= 0 {
						maxH = 5
					}
					if maxH > len(drp.Options) {
						maxH = len(drp.Options)
					}
					
					if state.FocusedIndex >= state.ScrollOffset + maxH {
						state.ScrollOffset = state.FocusedIndex - maxH + 1
					}
					if state.FocusedIndex < state.ScrollOffset {
						state.ScrollOffset = 0
					}
				}
			} else if ev.Key() == tcell.KeyPgUp {
				if state.Open {
					maxH := drp.MaxListHeight
					if maxH <= 0 {
						maxH = 5
					}
					if maxH > len(drp.Options) {
						maxH = len(drp.Options)
					}
					
					state.FocusedIndex -= maxH
					if state.FocusedIndex < 0 {
						state.FocusedIndex = 0
					}
					
					state.ScrollOffset -= maxH
					if state.ScrollOffset < 0 {
						state.ScrollOffset = 0
					}
				}
			} else if ev.Key() == tcell.KeyPgDn {
				if state.Open {
					maxH := drp.MaxListHeight
					if maxH <= 0 {
						maxH = 5
					}
					if maxH > len(drp.Options) {
						maxH = len(drp.Options)
					}
					
					state.FocusedIndex += maxH
					if state.FocusedIndex >= len(drp.Options) {
						state.FocusedIndex = len(drp.Options) - 1
					}
					
					state.ScrollOffset += maxH
					if state.ScrollOffset + maxH > len(drp.Options) {
						state.ScrollOffset = len(drp.Options) - maxH
						if state.ScrollOffset < 0 {
							state.ScrollOffset = 0
						}
					}
				}
			}
		} else if _, ok := focusedNode.(*ScrollView); ok {
			stateObj, ok := a.componentStates[a.focusedID]
			var state *ScrollViewState
			if !ok {
				state = &ScrollViewState{}
				a.componentStates[a.focusedID] = state
			} else {
				state = stateObj.(*ScrollViewState)
			}

			if ev.Key() == tcell.KeyUp {
				state.ScrollOffset--
				if state.ScrollOffset < 0 {
					state.ScrollOffset = 0
				}
			} else if ev.Key() == tcell.KeyDown {
				state.ScrollOffset++
			} else if ev.Key() == tcell.KeyPgUp {
				state.ScrollOffset -= 10
				if state.ScrollOffset < 0 {
					state.ScrollOffset = 0
				}
			} else if ev.Key() == tcell.KeyPgDn {
				state.ScrollOffset += 10
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
	return false
}
