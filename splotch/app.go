package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// App manages the terminal screen and the event loop.
type App struct {
	screen          tcell.Screen
	focusedID       string
	componentStates map[string]any
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

	// Set default style
	a.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))

	for {
		a.screen.Clear()

		// 1. Get the current UI tree
		root := renderFn()

		focusableIDs := findFocusableIDs(root)
		if a.focusedID == "" && len(focusableIDs) > 0 {
			a.focusedID = focusableIDs[0]
		}

		// 2. Compute layout
		w, h := a.screen.Size()
		layout := Layout(root, 0, 0, Constraints{MaxW: w, MaxH: h})

		// 3. Render
		Render(a.screen, layout, a.focusedID, a.componentStates)

		// Handle cursor for focused TextInput
		if a.focusedID != "" {
			focusedNode := findNodeByID(root, a.focusedID)
			if input, ok := focusedNode.(*TextInput); ok {
				res := findLayoutResultByID(layout, a.focusedID)
				if res != nil {
					stateObj, ok := a.componentStates[a.focusedID]
					if ok {
						state := stateObj.(*TextInputState)
						visualOffset := state.cursorOffset - state.scrollOffset
						a.screen.ShowCursor(res.X+input.Style.Padding.Left+visualOffset, res.Y+input.Style.Padding.Top)
					} else {
						a.screen.ShowCursor(res.X+len(input.Value), res.Y)
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
						w := res.W - input.Style.Padding.Left - input.Style.Padding.Right
						if w > 0 {
							if state.cursorOffset < state.scrollOffset {
								state.scrollOffset = state.cursorOffset
							}
							if state.cursorOffset > state.scrollOffset+w {
								state.scrollOffset = state.cursorOffset - w
							}
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
	}
	return nil
}

type TextInputState struct {
	cursorOffset int
	scrollOffset int
}
