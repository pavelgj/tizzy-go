package splotch

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gdamore/tcell/v2"
)

// App manages the terminal screen and the event loop.
type App struct {
	screen          tcell.Screen
	focusedID       string
	componentStates map[string]any
	previousGrid    *Grid
	activeCleanups  map[string]func()
	dirty           bool
	previousRoot    Node
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
		activeCleanups:  make(map[string]func()),
	}, nil
}

// EffectRecord stores an effect and its ID.
type EffectRecord struct {
	ID     string
	Effect func() func()
}

// RenderContext provides hooks and state scoping during rendering.
type RenderContext struct {
	app       *App
	effects   []EffectRecord
	hookIndex int
}

// UseState retrieves or initializes state.
// It returns the current state and a setter function.
func (ctx *RenderContext) UseState(initial any) (any, func(any)) {
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	state := ctx.app.componentStates[id]
	if state == nil {
		state = initial
		ctx.app.componentStates[id] = state
	}
	setter := func(newVal any) {
		ctx.app.componentStates[id] = newVal
		ctx.app.dirty = true
	}
	return state, setter
}

// UseState is a type-safe wrapper around RenderContext.UseState.
func UseState[T any](ctx *RenderContext, initial T) (T, func(T)) {
	stateObj, setter := ctx.UseState(initial)
	return stateObj.(T), func(newVal T) {
		setter(newVal)
	}
}

// UseEffect registers a lifecycle effect.
// The effect function should return a cleanup function (or nil).
func (ctx *RenderContext) UseEffect(effect func() func()) {
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	ctx.effects = append(ctx.effects, EffectRecord{ID: id, Effect: effect})
}

// Run starts the application loop.
// It takes a function that returns the root Node of the UI,
// and a function to handle events and update state.
func (a *App) Run(renderFn func(ctx *RenderContext) Node, updateFn func(tcell.Event)) error {
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
		// 1. Get the current UI tree or reuse previous
		var root Node
		if a.dirty || a.previousRoot == nil {
			ctx := &RenderContext{app: a}
			root = renderFn(ctx)
			a.previousRoot = root
			a.dirty = false

			// Collect all effect IDs requested in this frame
			requestedEffects := make(map[string]bool)
			for _, eff := range ctx.effects {
				requestedEffects[eff.ID] = true
			}

			// Detect Unmounts and run cleanups (not called implies unmounted)
			for id, cleanup := range a.activeCleanups {
				if !requestedEffects[id] {
					if cleanup != nil {
						cleanup()
					}
					delete(a.activeCleanups, id)
				}
			}

			// Detect Mounts and run effects
			for _, effectRec := range ctx.effects {
				id := effectRec.ID
				if _, active := a.activeCleanups[id]; !active {
					cleanup := effectRec.Effect()
					if cleanup != nil {
						a.activeCleanups[id] = cleanup
					}
				}
			}
		} else {
			root = a.previousRoot
		}

		// Find open modal if any
		var activeModal *Modal
		for id, stateObj := range a.componentStates {
			if state, ok := stateObj.(*ModalState); ok && state.Open {
				node := findNodeByID(root, id)
				if n, ok := node.(*Modal); ok {
					activeModal = n
					break
				}
			}
		}

		focusableIDs := []string{}
		if activeModal != nil {
			focusableIDs = findFocusableIDs(activeModal.Child, a.componentStates)
			// Ensure focus is inside modal
			found := false
			for _, id := range focusableIDs {
				if id == a.focusedID {
					found = true
					break
				}
			}
			if !found && len(focusableIDs) > 0 {
				a.focusedID = focusableIDs[0]
			}
		} else {
			focusableIDs = findFocusableIDs(root, a.componentStates)
			if a.focusedID == "" && len(focusableIDs) > 0 {
				a.focusedID = focusableIDs[0]
			}
		}

		// 2. Compute layout
		w, h := a.screen.Size()
		layout := Layout(root, 0, 0, Constraints{MaxW: w, MaxH: h})

		// 3. Render to grid
		grid := NewGrid(w, h)
		Render(grid, layout, a.focusedID, a.componentStates)

		// Render Modal overlay if active
		if activeModal != nil {
			maxModalW := w - 4
			maxModalH := h - 4
			if maxModalW < 0 {
				maxModalW = 0
			}
			if maxModalH < 0 {
				maxModalH = 0
			}

			modalConstraints := Constraints{
				MaxW: maxModalW,
				MaxH: maxModalH,
			}

			// Layout at 0,0 to find size
			modalLayout := Layout(activeModal.Child, 0, 0, modalConstraints)

			modalW := modalLayout.W + 2
			modalH := modalLayout.H + 2

			if modalW > w {
				modalW = w
			}
			if modalH > h {
				modalH = h
			}

			modalX := (w - modalW) / 2
			modalY := (h - modalH) / 2

			// Layout at correct position
			modalLayout = Layout(activeModal.Child, modalX+1, modalY+1, modalConstraints)

			style := tcell.StyleDefault.Foreground(activeModal.Style.Color).Background(activeModal.Style.Background)
			drawBorder(grid, modalX, modalY, modalW, modalH, style)

			for y := modalY + 1; y < modalY+modalH-1; y++ {
				for x := modalX + 1; x < modalX+modalW-1; x++ {
					grid.SetContent(x, y, ' ', style)
				}
			}

			Render(grid, modalLayout, a.focusedID, a.componentStates)
		}

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

		// Render MenuBar overlays
		for id, stateObj := range a.componentStates {
			if state, ok := stateObj.(*MenuBarState); ok && state.OpenMenuIndex >= 0 {
				debugLog(fmt.Sprintf("Found open MenuBar state for ID: %s, open index: %d", id, state.OpenMenuIndex))
				res := findLayoutResultByID(layout, id)
				if res == nil {
					debugLog(fmt.Sprintf("  Layout result not found for ID: %s", id))
				}
				menuBarNode := findNodeByID(root, id)
				if menuBarNode == nil {
					debugLog(fmt.Sprintf("  Node not found for ID: %s", id))
				}
				if res != nil && menuBarNode != nil {
					if mb, ok := menuBarNode.(*MenuBar); ok {
						borderOffset := 0
						if mb.Style.Border { borderOffset = 1 }
						curX := res.X + borderOffset + mb.Style.Padding.Left
						
						menuX := curX
						for i := 0; i < state.OpenMenuIndex; i++ {
							menuX += len(mb.Menus[i].Title) + 4
						}
						
						openMenu := mb.Menus[state.OpenMenuIndex]
						listY := res.Y + borderOffset + mb.Style.Padding.Top + 1
						
						listW := 0
						for _, item := range openMenu.Items {
							if len(item.Label) > listW { listW = len(item.Label) }
						}
						listW += 4 // +2 for padding, +2 for borders
						
						listH := len(openMenu.Items) + 2 // +2 for top/bottom borders
						
						style := tcell.StyleDefault.Foreground(mb.Style.Color).Background(tcell.ColorBlack)
						
						// Draw Shadow (right and bottom edges only)
						for i := 1; i <= listH; i++ {
							if listY+i < h && menuX+listW < w {
								currentCell := grid.Cells[listY+i][menuX+listW]
								grid.SetContent(menuX+listW, listY+i, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
							}
						}
						for j := 1; j <= listW; j++ {
							if listY+listH < h && menuX+j < w {
								currentCell := grid.Cells[listY+listH][menuX+j]
								grid.SetContent(menuX+j, listY+listH, currentCell.Rune, currentCell.Style.Background(tcell.ColorDarkGray))
							}
						}

						// Fill background
						for i := 0; i < listH; i++ {
							for j := 0; j < listW; j++ {
								if listY+i < h && menuX+j < w {
									grid.SetContent(menuX+j, listY+i, ' ', style)
								}
							}
						}
						
						// Draw Border
						drawBorder(grid, menuX, listY, listW, listH, style)
						
						for i, item := range openMenu.Items {
							itemStyle := style
							if state.FocusedItemIndex == i {
								itemStyle = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
							}
							if item.Disabled {
								itemStyle = itemStyle.Foreground(tcell.ColorGray)
							}
							
							label := " " + item.Label
							for len(label) < listW-2 {
								label += " "
							}
							
							curItemX := menuX + 1
							for _, r := range label {
								if listY+i+1 < h && curItemX < w {
									grid.SetContent(curItemX, listY+i+1, r, itemStyle)
									curItemX++
								}
							}
						}
					}
				}
			}
		}

		// Render Popup overlays
		var activePopups []*Popup
		for id, stateObj := range a.componentStates {
			if state, ok := stateObj.(*PopupState); ok && state.Open {
				node := findNodeByID(root, id)
				if n, ok := node.(*Popup); ok {
					activePopups = append(activePopups, n)
				}
			}
		}

		for _, popup := range activePopups {
			maxPopupW := w - popup.X
			maxPopupH := h - popup.Y
			if maxPopupW < 0 { maxPopupW = 0 }
			if maxPopupH < 0 { maxPopupH = 0 }

			popupConstraints := Constraints{
				MaxW: maxPopupW,
				MaxH: maxPopupH,
			}

			popupLayout := Layout(popup.Child, popup.X, popup.Y, popupConstraints)
			
			// Fill background to prevent see-through
			style := tcell.StyleDefault.Foreground(popup.Style.Color).Background(popup.Style.Background)
			for y := popup.Y; y < popup.Y+popupLayout.H; y++ {
				for x := popup.X; x < popup.X+popupLayout.W; x++ {
					if x < w && y < h {
						grid.SetContent(x, y, ' ', style)
					}
				}
			}

			Render(grid, popupLayout, a.focusedID, a.componentStates)
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
						if input.Cursor != nil {
							state.cursorOffset = *input.Cursor
						}
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
			
			// Handle MenuBar hover-to-switch
			var openMenuBar *MenuBar
			var openMenuBarID string
			for id, stateObj := range a.componentStates {
				if state, ok := stateObj.(*MenuBarState); ok && state.OpenMenuIndex >= 0 {
					node := findNodeByID(root, id)
					if mb, ok := node.(*MenuBar); ok {
						openMenuBar = mb
						openMenuBarID = id
						break
					}
				}
			}
			
			if openMenuBar != nil {
				res := findLayoutResultByID(layout, openMenuBarID)
				if res != nil {
					borderOffset := 0
					if openMenuBar.Style.Border { borderOffset = 1 }
					curX := res.X + borderOffset + openMenuBar.Style.Padding.Left
					curY := res.Y + borderOffset + openMenuBar.Style.Padding.Top
					
					if my == curY {
						for i, menu := range openMenuBar.Menus {
							titleLen := len(menu.Title) + 2
							if mx >= curX && mx < curX+titleLen {
								stateObj := a.componentStates[openMenuBarID]
								state := stateObj.(*MenuBarState)
								if state.OpenMenuIndex != i {
									state.OpenMenuIndex = i
									state.FocusedItemIndex = -1
									a.dirty = true
								}
								break
							}
							curX += titleLen + 2
						}
					}
				}
			}

			if ev.Buttons()&tcell.Button1 != 0 {
				handled := false
				
				// Find open modal if any
				var activeModal *Modal
				for id, stateObj := range a.componentStates {
					if state, ok := stateObj.(*ModalState); ok && state.Open {
						node := findNodeByID(root, id)
						if n, ok := node.(*Modal); ok {
							activeModal = n
							break
						}
					}
				}
				
				if activeModal != nil {
					w, h := a.screen.Size()
					maxModalW := w - 4
					maxModalH := h - 4
					if maxModalW < 0 { maxModalW = 0 }
					if maxModalH < 0 { maxModalH = 0 }
					
					modalConstraints := Constraints{
						MaxW: maxModalW,
						MaxH: maxModalH,
					}
					
					modalLayout := Layout(activeModal.Child, 0, 0, modalConstraints)
					modalW := modalLayout.W + 2
					modalH := modalLayout.H + 2
					
					if modalW > w { modalW = w }
					if modalH > h { modalH = h }
					
					modalX := (w - modalW) / 2
					modalY := (h - modalH) / 2
					
					modalLayout = Layout(activeModal.Child, modalX+1, modalY+1, modalConstraints)
					
					path := findNodePathAt(modalLayout, mx, my, a.componentStates)
					if len(path) > 0 {
						clickedNode := path[len(path)-1]
						
						var nodeStyle Style
						var focusableNode Node
						for i := len(path) - 1; i >= 0; i-- {
							n := path[i]
							var s Style
							switch node := n.(type) {
							case *Text: s = node.Style
							case *TextInput: s = node.Style
							case *Button: s = node.Style
							case *Checkbox: s = node.Style
							case *RadioButton: s = node.Style
							case *Spinner: s = node.Style
							case *ProgressBar: s = node.Style
							case *ScrollView: s = node.Style
							case *Dropdown: s = node.Style
							case *Box: s = node.Style
							case *Modal: s = node.Style
							}
							if s.Focusable && s.ID != "" {
								focusableNode = n
								nodeStyle = s
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
						
						handled = true
					} else {
						// Trap clicks outside modal
						handled = true
					}
				}
				
				if !handled {
					var openMenuBar *MenuBar
					var openMenuBarID string
					for id, stateObj := range a.componentStates {
						if state, ok := stateObj.(*MenuBarState); ok && state.OpenMenuIndex >= 0 {
							node := findNodeByID(root, id)
							if mb, ok := node.(*MenuBar); ok {
								openMenuBar = mb
								openMenuBarID = id
								break
							}
						}
					}
					
					if openMenuBar != nil {
						stateObj := a.componentStates[openMenuBarID]
						state := stateObj.(*MenuBarState)
						if state.OpenMenuIndex >= 0 {
							res := findLayoutResultByID(layout, openMenuBarID)
							if res != nil {
								borderOffset := 0
								if openMenuBar.Style.Border { borderOffset = 1 }
								curX := res.X + borderOffset + openMenuBar.Style.Padding.Left
								
								menuX := curX
								for i := 0; i < state.OpenMenuIndex; i++ {
									menuX += len(openMenuBar.Menus[i].Title) + 4
								}
								
								openMenu := openMenuBar.Menus[state.OpenMenuIndex]
								listY := res.Y + borderOffset + openMenuBar.Style.Padding.Top + 1
								listW := 0
								for _, item := range openMenu.Items {
									if len(item.Label) > listW { listW = len(item.Label) }
								}
								listW += 4 // +2 for padding, +2 for borders
								listH := len(openMenu.Items) + 2 // +2 for borders
								
								if mx >= menuX && mx < menuX+listW && my >= listY && my < listY+listH {
									clickedIndex := my - listY - 1 // -1 for top border
									if clickedIndex >= 0 && clickedIndex < len(openMenu.Items) {
										item := openMenu.Items[clickedIndex]
										if !item.Disabled && item.Action != nil {
											item.Action()
										}
										state.OpenMenuIndex = -1
										a.dirty = true
										handled = true
									}
								}
							}
						}
					}
					
					if !handled {
						menuBar := findMenuBar(root)
						if menuBar != nil && menuBar.Style.ID != "" {
							res := findLayoutResultByID(layout, menuBar.Style.ID)
							if res != nil {
								borderOffset := 0
								if menuBar.Style.Border { borderOffset = 1 }
								curX := res.X + borderOffset + menuBar.Style.Padding.Left
								curY := res.Y + borderOffset + menuBar.Style.Padding.Top
								
								if my == curY {
									for i, menu := range menuBar.Menus {
										titleLen := len(menu.Title) + 2
										if mx >= curX && mx < curX+titleLen {
											stateObj, ok := a.componentStates[menuBar.Style.ID]
											var state *MenuBarState
											if !ok {
												state = &MenuBarState{OpenMenuIndex: -1}
												a.componentStates[menuBar.Style.ID] = state
											} else {
												state = stateObj.(*MenuBarState)
											}
											
											if state.OpenMenuIndex == i {
												state.OpenMenuIndex = -1
											} else {
												state.OpenMenuIndex = i
												state.FocusedItemIndex = -1
											}
											handled = true
											break
										}
										curX += titleLen + 2
									}
								}
							}
						}
					}
				}

				if !handled {
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
				}

				if !handled {
					path := findNodePathAt(layout, mx, my, a.componentStates)
					if len(path) > 0 {
						debugLog(fmt.Sprintf("Mouse click at %d,%d. Path len: %d", mx, my, len(path)))
						if len(path) > 0 {
							debugLog(fmt.Sprintf("Leaf node: %T", path[len(path)-1]))
							for i, n := range path {
								debugLog(fmt.Sprintf("  Path[%d]: %T", i, n))
							}
						}

						// Filter out clicks targeting hidden tab content
						validPath := true
						for i := 0; i < len(path)-1; i++ {
							if tabs, ok := path[i].(*Tabs); ok && tabs.Style.ID != "" {
								stateObj, ok := a.componentStates[tabs.Style.ID]
								activeIdx := 0
								if ok {
									activeIdx = stateObj.(*TabsState).ActiveTab
								}
								activeChild := tabs.Tabs[activeIdx].Content
								if path[i+1] != activeChild {
									validPath = false
									break
								}
							}
						}

						if !validPath {
							handled = true
						} else {
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
								case *Tabs:
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
						if tabs, ok := clickedNode.(*Tabs); ok && tabs.Style.ID != "" {
							res := findLayoutResultByID(layout, tabs.Style.ID)
							if res != nil {
								curX := res.X + tabs.Style.Padding.Left
								curY := res.Y + tabs.Style.Padding.Top
								
								if my == curY {
									for i, tab := range tabs.Tabs {
										labelLen := len(tab.Label) + 4 // "[ " + label + " ]"
										if mx >= curX && mx < curX+labelLen {
											stateObj, ok := a.componentStates[tabs.Style.ID]
											var state *TabsState
											if !ok {
												state = &TabsState{}
												a.componentStates[tabs.Style.ID] = state
											} else {
												state = stateObj.(*TabsState)
											}
											state.ActiveTab = i
											a.dirty = true
											break
										}
										curX += labelLen
									}
								}
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
				}
			} else if ev.Buttons()&tcell.Button4 != 0 || ev.Buttons()&tcell.WheelUp != 0 { // Wheel Up
				dropdownScrolled := false
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
									maxH = 5
								}
								if maxH > len(drp.Options) {
									maxH = len(drp.Options)
								}
								listH := maxH
								
								if mx >= res.X && mx < res.X+listW && my >= listY && my < listY+listH {
									state.ScrollOffset--
									if state.ScrollOffset < 0 {
										state.ScrollOffset = 0
									}
									dropdownScrolled = true
									break
								}
							}
						}
					}
				}
				
				if !dropdownScrolled {
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
				}
			} else if ev.Buttons()&tcell.Button5 != 0 || ev.Buttons()&tcell.WheelDown != 0 { // Wheel Down
				dropdownScrolled := false
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
									maxH = 5
								}
								if maxH > len(drp.Options) {
									maxH = len(drp.Options)
								}
								listH := maxH
								
								if mx >= res.X && mx < res.X+listW && my >= listY && my < listY+listH {
									state.ScrollOffset++
									if state.ScrollOffset+maxH > len(drp.Options) {
										state.ScrollOffset = len(drp.Options) - maxH
										if state.ScrollOffset < 0 {
											state.ScrollOffset = 0
										}
									}
									dropdownScrolled = true
									break
								}
							}
						}
					}
				}
				
				if !dropdownScrolled {
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



func findFocusableIDs(node Node, componentStates map[string]any) []string {
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
	case *MenuBar:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
	case *Tabs:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
		activeIdx := 0
		if n.Style.ID != "" && componentStates != nil {
			if stateObj, ok := componentStates[n.Style.ID]; ok {
				activeIdx = stateObj.(*TabsState).ActiveTab
			}
		}
		if activeIdx >= 0 && activeIdx < len(n.Tabs) {
			ids = append(ids, findFocusableIDs(n.Tabs[activeIdx].Content, componentStates)...)
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
	case *ScrollView:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
		ids = append(ids, findFocusableIDs(n.Child, componentStates)...)
	case *Box:
		if n.Style.Focusable && n.Style.ID != "" {
			ids = append(ids, n.Style.ID)
		}
		for _, child := range n.Children {
			ids = append(ids, findFocusableIDs(child, componentStates)...)
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
	case *Tabs:
		if n.Style.ID == id {
			return n
		}
		for _, tab := range n.Tabs {
			if found := findNodeByID(tab.Content, id); found != nil {
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
	case *Dropdown:
		if n.Style.ID == id {
			return n
		}
	case *MenuBar:
		if n.Style.ID == id {
			return n
		}
	case *Modal:
		if n.Style.ID == id {
			return n
		}
		return findNodeByID(n.Child, id)
	case *Popup:
		if n.Style.ID == id {
			return n
		}
		return findNodeByID(n.Child, id)
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
	case *Tabs:
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
	case *MenuBar:
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
	debugLog(fmt.Sprintf("Key Event: Key=%v, Rune=%v, Mod=%v", ev.Key(), ev.Rune(), ev.Modifiers()))

	// Check if any popup is open
	popupOpen := false
	for _, stateObj := range a.componentStates {
		if state, ok := stateObj.(*PopupState); ok && state.Open {
			popupOpen = true
			break
		}
	}

	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		if popupOpen && ev.Key() == tcell.KeyEscape {
			// Let falling through to updateFn handle closing the popup
			return false
		}
		return true // Request exit
	}

	// Handle MenuBar shortcuts (Alt+Letter or Mac fallback)
	runeLower := unicode.ToLower(ev.Rune())
	isMacAlt := false
	if ev.Key() == tcell.KeyRune && ev.Modifiers() == 0 {
		// Map common Mac Option+Letter characters
		if runeLower == 402 { runeLower = 'f'; isMacAlt = true }
		if runeLower == 180 { runeLower = 'e'; isMacAlt = true }
		if runeLower == 729 { runeLower = 'h'; isMacAlt = true }
	}

	if (ev.Modifiers()&tcell.ModAlt != 0 && ev.Key() == tcell.KeyRune) || isMacAlt {
		menuBar := findMenuBar(root)
		
		if menuBar != nil && menuBar.Style.ID != "" {
			for i, menu := range menuBar.Menus {
				if unicode.ToLower(menu.AltRune) == runeLower {
					stateObj, ok := a.componentStates[menuBar.Style.ID]
					var state *MenuBarState
					if !ok {
						state = &MenuBarState{OpenMenuIndex: -1}
						a.componentStates[menuBar.Style.ID] = state
					} else {
						state = stateObj.(*MenuBarState)
					}
					
					if state.OpenMenuIndex == i {
						state.OpenMenuIndex = -1 // Toggle close
					} else {
						state.OpenMenuIndex = i // Open
						state.FocusedItemIndex = -1 // Reset focus
					}
					return false // Handled
				}
			}
		}
	}

	// Handle MenuBar direct letter shortcuts when focused
	if ev.Modifiers()&tcell.ModAlt == 0 && ev.Key() == tcell.KeyRune {
		menuBar := findMenuBar(root)
		if menuBar != nil && menuBar.Style.ID != "" && a.focusedID == menuBar.Style.ID {
			runeLower := unicode.ToLower(ev.Rune())
			for i, menu := range menuBar.Menus {
				if unicode.ToLower(menu.AltRune) == runeLower {
					stateObj, ok := a.componentStates[menuBar.Style.ID]
					var state *MenuBarState
					if !ok {
						state = &MenuBarState{OpenMenuIndex: -1}
						a.componentStates[menuBar.Style.ID] = state
					} else {
						state = stateObj.(*MenuBarState)
					}
					
					state.OpenMenuIndex = i
					state.FocusedItemIndex = -1
					return false
				}
			}
		}
	}

	// Handle MenuBar arrow navigation when open
	menuBar := findMenuBar(root)
	if menuBar != nil && menuBar.Style.ID != "" {
		stateObj, ok := a.componentStates[menuBar.Style.ID]
		if ok {
			state := stateObj.(*MenuBarState)
			if state.OpenMenuIndex >= 0 {
				openMenu := menuBar.Menus[state.OpenMenuIndex]
				if ev.Key() == tcell.KeyDown {
					state.FocusedItemIndex++
					if state.FocusedItemIndex >= len(openMenu.Items) {
						state.FocusedItemIndex = 0
					}
					a.dirty = true
					return false
				} else if ev.Key() == tcell.KeyUp {
					state.FocusedItemIndex--
					if state.FocusedItemIndex < 0 {
						state.FocusedItemIndex = len(openMenu.Items) - 1
					}
					a.dirty = true
					return false
				} else if ev.Key() == tcell.KeyRight {
					state.OpenMenuIndex++
					if state.OpenMenuIndex >= len(menuBar.Menus) {
						state.OpenMenuIndex = 0
					}
					state.FocusedItemIndex = -1
					a.dirty = true
					return false
				} else if ev.Key() == tcell.KeyLeft {
					state.OpenMenuIndex--
					if state.OpenMenuIndex < 0 {
						state.OpenMenuIndex = len(menuBar.Menus) - 1
					}
					state.FocusedItemIndex = -1
					a.dirty = true
					return false
				} else if ev.Key() == tcell.KeyEnter {
					if state.FocusedItemIndex >= 0 && state.FocusedItemIndex < len(openMenu.Items) {
						item := openMenu.Items[state.FocusedItemIndex]
						if !item.Disabled && item.Action != nil {
							item.Action()
						}
						state.OpenMenuIndex = -1
						a.dirty = true
						return false
					}
				} else if ev.Key() == tcell.KeyEscape {
					state.OpenMenuIndex = -1
					a.dirty = true
					return false
				}
			}
		}
	}

	// Handle Tabs keyboard navigation when focused
	tabs := findTabs(root)
	if tabs != nil && tabs.Style.ID != "" && a.focusedID == tabs.Style.ID {
		stateObj, ok := a.componentStates[tabs.Style.ID]
		var state *TabsState
		if !ok {
			state = &TabsState{}
			a.componentStates[tabs.Style.ID] = state
		} else {
			state = stateObj.(*TabsState)
		}

		if ev.Key() == tcell.KeyRight {
			state.ActiveTab++
			if state.ActiveTab >= len(tabs.Tabs) {
				state.ActiveTab = 0
			}
			a.dirty = true
			return false
		} else if ev.Key() == tcell.KeyLeft {
			state.ActiveTab--
			if state.ActiveTab < 0 {
				state.ActiveTab = len(tabs.Tabs) - 1
			}
			a.dirty = true
			return false
		}
	}

	if ev.Key() == tcell.KeyTab {
		a.focusedID = nextFocus(a.focusedID, focusableIDs)
	}
	if ev.Key() == tcell.KeyBacktab {
		a.focusedID = prevFocus(a.focusedID, focusableIDs)
	}
	if ev.Key() == tcell.KeyTab || ev.Key() == tcell.KeyBacktab {
		// Close all MenuBar menus on focus change
		for _, stateObj := range a.componentStates {
			if state, ok := stateObj.(*MenuBarState); ok {
				state.OpenMenuIndex = -1
			}
		}

		a.closeOtherDropdowns(a.focusedID)
		
		if a.focusedID != "" {
			focusedNode := findNodeByID(root, a.focusedID)
			if _, ok := focusedNode.(*MenuBar); ok {
				stateObj, ok := a.componentStates[a.focusedID]
				var state *MenuBarState
				if !ok {
					state = &MenuBarState{OpenMenuIndex: 0}
					a.componentStates[a.focusedID] = state
				} else {
					state = stateObj.(*MenuBarState)
					state.OpenMenuIndex = 0
					state.FocusedItemIndex = -1
				}
			}
		}
	}
	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if input, ok := focusedNode.(*TextInput); ok {
			if popupOpen {
				if ev.Key() == tcell.KeyEnter || ev.Key() == tcell.KeyUp || ev.Key() == tcell.KeyDown {
					return false // Fall through to updateFn
				}
			}

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
			a.dirty = true
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
					if state.FocusedIndex >= state.ScrollOffset+maxH {
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

					if state.FocusedIndex >= state.ScrollOffset+maxH {
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
					if state.ScrollOffset+maxH > len(drp.Options) {
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



func (a *App) Stop() {
	a.screen.Fini()
	os.Exit(0)
}

func findMenuBar(node Node) *MenuBar {
	switch tn := node.(type) {
	case *MenuBar:
		return tn
	case *Box:
		for _, child := range tn.Children {
			if mb := findMenuBar(child); mb != nil {
				return mb
			}
		}
	case *ScrollView:
		return findMenuBar(tn.Child)
	case *Modal:
		return findMenuBar(tn.Child)
	}
	return nil
}

func findTabs(node Node) *Tabs {
	switch tn := node.(type) {
	case *Tabs:
		return tn
	case *Box:
		for _, child := range tn.Children {
			if t := findTabs(child); t != nil {
				return t
			}
		}
	case *ScrollView:
		return findTabs(tn.Child)
	case *Modal:
		return findTabs(tn.Child)
	}
	return nil
}


