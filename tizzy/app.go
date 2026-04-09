package tizzy

import (
	"fmt"
	"os"
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

// GetComponentState retrieves state for a component by ID.
func (a *App) GetComponentState(id string) (any, bool) {
	state, ok := a.componentStates[id]
	return state, ok
}

// GetFocusedID returns the ID of the currently focused component.
func (a *App) GetFocusedID() string {
	return a.focusedID
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

// GetFocusedID returns the ID of the currently focused component.
func (ctx *RenderContext) GetFocusedID() string {
	return ctx.app.GetFocusedID()
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
			_ = a.screen.PostEvent(&EventTick{t: time.Now()})
		}
	}()

	for {
		// 1. Get the current UI tree or reuse previous
		var root Node
		if a.dirty || a.previousRoot == nil {
			ctx := &RenderContext{app: a}
			root = renderFn(ctx)
			if err := validateUniqueIDs(root); err != nil {
				return err
			}
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

		var focusableIDs []string
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

		// Update scroll offset for focused TextInput based on up-to-date layout
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
							line, col := offsetToLineCol(input.Value, state.cursorOffset)
							h := res.H - input.Style.Padding.Top - input.Style.Padding.Bottom - borderOffset*2
							if h > 0 {
								if line < state.vScrollOffset {
									state.vScrollOffset = line
								}
								if line >= state.vScrollOffset+h {
									state.vScrollOffset = line - h + 1
								}
							}

							// Horizontal scroll for multiline
							if w > 0 {
								if col < state.scrollOffset {
									state.scrollOffset = col
								}
								if col >= state.scrollOffset+w {
									state.scrollOffset = col - w + 1
								}
							}
						}
					}
				}
			}
		}

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
			drawBorder(grid, modalX, modalY, modalW, modalH, "", style)

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
						if mb.Style.Border {
							borderOffset = 1
						}
						curX := res.X + borderOffset + mb.Style.Padding.Left

						menuX := curX
						for i := 0; i < state.OpenMenuIndex; i++ {
							menuX += len(mb.Menus[i].Title) + 4
						}

						openMenu := mb.Menus[state.OpenMenuIndex]
						listY := res.Y + borderOffset + mb.Style.Padding.Top + 1

						listW := 0
						for _, item := range openMenu.Items {
							if len(item.Label) > listW {
								listW = len(item.Label)
							}
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
						drawBorder(grid, menuX, listY, listW, listH, "", style)

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
			if maxPopupW < 0 {
				maxPopupW = 0
			}
			if maxPopupH < 0 {
				maxPopupH = 0
			}

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
				var res *LayoutResult
				for id, stateObj := range a.componentStates {
					if state, ok := stateObj.(*ModalState); ok && state.Open {
						node := findNodeByID(root, id)
						if modal, ok := node.(*Modal); ok {
							w, h := a.screen.Size()
							maxModalW := w - 4
							maxModalH := h - 4
							if maxModalW < 0 { maxModalW = 0 }
							if maxModalH < 0 { maxModalH = 0 }
							
							modalConstraints := Constraints{
								MaxW: maxModalW,
								MaxH: maxModalH,
							}
							
							modalLayout := Layout(modal.Child, 0, 0, modalConstraints)
							modalW := modalLayout.W + 2
							modalH := modalLayout.H + 2
							
							if modalW > w { modalW = w }
							if modalH > h { modalH = h }
							
							modalX := (w - modalW) / 2
							modalY := (h - modalH) / 2
							
							ml := Layout(modal.Child, modalX+1, modalY+1, modalConstraints)
							res = findLayoutResultByID(ml, a.focusedID)
							if res != nil {
								break
							}
						}
					}
				}
				if res == nil {
					res = findLayoutResultByID(layout, a.focusedID)
				}
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
							a.screen.ShowCursor(res.X+input.Style.Padding.Left+col-state.scrollOffset+borderOffset, res.Y+input.Style.Padding.Top+line+borderOffset-state.vScrollOffset)
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
			a.handleMouseEvent(ev, root, layout)

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

	if node == nil {
		return ids
	}

	if f, ok := node.(Focusable); ok && f.IsFocusable() && node.GetStyle().ID != "" {
		ids = append(ids, node.GetStyle().ID)
	}

	switch n := node.(type) {
	case *Tabs:
		activeIdx := 0
		if n.Style.ID != "" && componentStates != nil {
			if stateObj, ok := componentStates[n.Style.ID]; ok {
				activeIdx = stateObj.(*TabsState).ActiveTab
			}
		}
		if activeIdx >= 0 && activeIdx < len(n.Tabs) {
			ids = append(ids, findFocusableIDs(n.Tabs[activeIdx].Content, componentStates)...)
		}
	default:
		if p, ok := node.(ParentNode); ok {
			for _, child := range p.GetChildren() {
				ids = append(ids, findFocusableIDs(child, componentStates)...)
			}
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
	if node == nil {
		return nil
	}
	if node.GetStyle().ID == id {
		return node
	}
	if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
			if found := findNodeByID(child, id); found != nil {
				return found
			}
		}
	}
	return nil
}

func validateUniqueIDs(node Node) error {
	seen := make(map[string]bool)
	return walkTree(node, func(n Node) error {
		id := n.GetStyle().ID
		if id != "" {
			if seen[id] {
				return fmt.Errorf("duplicate component ID: %s", id)
			}
			seen[id] = true
		}
		return nil
	})
}

func walkTree(node Node, fn func(Node) error) error {
	if node == nil {
		return nil
	}
	if err := fn(node); err != nil {
		return err
	}
	if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
			if err := walkTree(child, fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func findLayoutResultByID(res LayoutResult, id string) *LayoutResult {
	if res.Node.GetStyle().ID == id {
		return &res
	}
	for _, child := range res.Children {
		if found := findLayoutResultByID(child, id); found != nil {
			return found
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

func findListAt(res LayoutResult, x, y int, componentStates map[string]any) *List {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		if l, ok := res.Node.(*List); ok {
			return l
		}

		for _, child := range res.Children {
			if lChild := findListAt(child, x, y, componentStates); lChild != nil {
				return lChild
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

		if tabs, ok := res.Node.(*Tabs); ok {
			activeIdx := 0
			if tabs.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[tabs.Style.ID]; ok {
					activeIdx = stateObj.(*TabsState).ActiveTab
				}
			}
			if activeIdx >= 0 && activeIdx < len(res.Children) {
				if path := findNodePathAt(res.Children[activeIdx], x, y+scrollOffset, componentStates); path != nil {
					return append([]Node{res.Node}, path...)
				}
			}
			return []Node{res.Node}
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

// MarkDirty forces a re-render on the next frame.
func (a *App) MarkDirty() {
	a.dirty = true
}

func (a *App) setFocus(id string, root Node) {
	if a.focusedID == id {
		return
	}
	a.focusedID = id
	a.dirty = true

	a.closeOtherDropdowns(id)

	if id != "" && root != nil {
		node := findNodeByID(root, id)
		if list, ok := node.(*List); ok {
			if list.OnFocus != nil {
				stateObj, ok := a.componentStates[id]
				var state *ListState
				if !ok {
					state = &ListState{}
					a.componentStates[id] = state
				} else {
					state = stateObj.(*ListState)
				}
				list.OnFocus(state)
			}
		}
	}
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

	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if handler, ok := focusedNode.(EventHandler); ok {
			stateObj, ok := a.componentStates[a.focusedID]
			var state any
			if !ok {
				state = handler.DefaultState()
				a.componentStates[a.focusedID] = state
			} else {
				state = stateObj
			}

			res := findLayoutResultByID(layout, a.focusedID)
			var layoutRes LayoutResult
			if res != nil {
				layoutRes = *res
			}

			eventCtx := EventContext{
				Layout:    layoutRes,
				PopupOpen: popupOpen,
			}

			if handler.HandleEvent(ev, state, eventCtx) {
				a.dirty = true
				return false
			}
		}
	}

	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		if popupOpen && ev.Key() == tcell.KeyEscape {
			// Let falling through to updateFn handle closing the popup
			return false
		}
		return true // Request exit
	}

	if ev.Key() == tcell.KeyTab {
		a.setFocus(nextFocus(a.focusedID, focusableIDs), root)
	}
	if ev.Key() == tcell.KeyBacktab {
		a.setFocus(prevFocus(a.focusedID, focusableIDs), root)
	}

	return false
}

func (a *App) Stop() {
	a.screen.Fini()
	os.Exit(0)
}

// MouseEvent is an interface for tcell.EventMouse to allow mocking in tests.
type MouseEvent interface {
	Position() (int, int)
	Buttons() tcell.ButtonMask
}

func (a *App) handleMouseEvent(ev MouseEvent, root Node, layout LayoutResult) bool {
	mx, my := ev.Position()

	// Handle generic overlays
	for id, stateObj := range a.componentStates {
		if openable, ok := stateObj.(OpenableState); ok && openable.IsOpen() {
			node := findNodeByID(root, id)
			if handler, ok := node.(OverlayHandler); ok {
				res := findLayoutResultByID(layout, id)
				var compLayout LayoutResult
				if res != nil {
					compLayout = *res
				}
				
				eventCtx := EventContext{
					Layout: compLayout,
				}
				
				// Special case for Modal layout calculation
				if modal, ok := node.(*Modal); ok {
					w, h := a.screen.Size()
					maxModalW := w - 4
					maxModalH := h - 4
					if maxModalW < 0 { maxModalW = 0 }
					if maxModalH < 0 { maxModalH = 0 }
					
					modalConstraints := Constraints{
						MaxW: maxModalW,
						MaxH: maxModalH,
					}
					
					modalLayout := Layout(modal.Child, 0, 0, modalConstraints)
					modalW := modalLayout.W + 2
					modalH := modalLayout.H + 2
					
					if modalW > w { modalW = w }
					if modalH > h { modalH = h }
					
					modalX := (w - modalW) / 2
					modalY := (h - modalH) / 2
					
					eventCtx.OverlayLayout = Layout(modal.Child, modalX+1, modalY+1, modalConstraints)
				}
				
				if tcellEv, ok := ev.(tcell.Event); ok {
					handled, searchLayout := handler.HandleOverlayEvent(tcellEv, stateObj, eventCtx)
					if handled {
						a.dirty = true
						return true
					}
					if searchLayout != nil {
						path := findNodePathAt(*searchLayout, mx, my, a.componentStates)
						if len(path) > 0 {
							a.dispatchEventToPath(path, tcellEv, root, *searchLayout)
							return true
						}
					}
				}
			}
		}
	}

	if ev.Buttons()&tcell.Button1 != 0 {
		handled := false



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

			debugLog(fmt.Sprintf("handleMouseEvent: openMenuBar=%v", openMenuBar))

			if openMenuBar != nil {
				stateObj := a.componentStates[openMenuBarID]
				state := stateObj.(*MenuBarState)
				if state.OpenMenuIndex >= 0 {
					res := findLayoutResultByID(layout, openMenuBarID)
					if res != nil {
						borderOffset := 0
						if openMenuBar.Style.Border {
							borderOffset = 1
						}
						curX := res.X + borderOffset + openMenuBar.Style.Padding.Left

						menuX := curX
						for i := 0; i < state.OpenMenuIndex; i++ {
							menuX += len(openMenuBar.Menus[i].Title) + 4
						}

						openMenu := openMenuBar.Menus[state.OpenMenuIndex]
						listY := res.Y + borderOffset + openMenuBar.Style.Padding.Top + 1
						listW := 0
						for _, item := range openMenu.Items {
							if len(item.Label) > listW {
								listW = len(item.Label)
							}
						}
						listW += 4                       // +2 for padding, +2 for borders
						listH := len(openMenu.Items) + 2 // +2 for borders

						debugLog(fmt.Sprintf("handleMouseEvent: mx=%d, my=%d, menuX=%d, listY=%d, listW=%d, listH=%d", mx, my, menuX, listY, listW, listH))
						if mx >= menuX && mx < menuX+listW && my >= listY && my < listY+listH {
							clickedIndex := my - listY - 1 // -1 for top border
							if clickedIndex >= 0 && clickedIndex < len(openMenu.Items) {
								item := openMenu.Items[clickedIndex]
								if item.Action != nil {
									item.Action()
									state.OpenMenuIndex = -1
									a.dirty = true
								}
							}
							handled = true
						} else {
							// Click outside menu closes it
							state.OpenMenuIndex = -1
							a.dirty = true
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
					// Do nothing
				} else {
					if tcellEv, ok := ev.(tcell.Event); ok {
						handled = a.dispatchEventToPath(path, tcellEv, root, layout)
					}
				}
			}
		}
	} else if ev.Buttons()&tcell.Button4 != 0 || ev.Buttons()&tcell.WheelUp != 0 { // Wheel Up
		path := findNodePathAt(layout, mx, my, a.componentStates)
		if len(path) > 0 {
			targetNode := path[len(path)-1]
			if handler, ok := targetNode.(EventHandler); ok {
				state := a.componentStates[targetNode.GetStyle().ID]
				res := findLayoutResultByID(layout, targetNode.GetStyle().ID)
				var targetLayout LayoutResult
				if res != nil {
					targetLayout = *res
				}
				eventCtx := EventContext{
					Layout: targetLayout,
				}
				if tcellEv, ok := ev.(tcell.Event); ok {
					if handler.HandleEvent(tcellEv, state, eventCtx) {
						a.dirty = true
					}
				}
			}
		}
	} else if ev.Buttons()&tcell.Button5 != 0 || ev.Buttons()&tcell.WheelDown != 0 { // Wheel Down
		path := findNodePathAt(layout, mx, my, a.componentStates)
		if len(path) > 0 {
			targetNode := path[len(path)-1]
			if handler, ok := targetNode.(EventHandler); ok {
				state := a.componentStates[targetNode.GetStyle().ID]
				res := findLayoutResultByID(layout, targetNode.GetStyle().ID)
				var targetLayout LayoutResult
				if res != nil {
					targetLayout = *res
				}
				eventCtx := EventContext{
					Layout: targetLayout,
				}
				if tcellEv, ok := ev.(tcell.Event); ok {
					if handler.HandleEvent(tcellEv, state, eventCtx) {
						a.dirty = true
					}
				}
			}
		}
	}
	return true
}

func (a *App) dispatchEventToPath(path []Node, ev tcell.Event, root Node, searchLayout LayoutResult) bool {
	if len(path) == 0 {
		return false
	}
	clickedNode := path[len(path)-1]

	var nodeStyle Style
	var focusableNode Node
	for i := len(path) - 1; i >= 0; i-- {
		n := path[i]
		nodeStyle = n.GetStyle()
		if nodeStyle.Focusable && nodeStyle.ID != "" {
			focusableNode = n
			break
		}
	}

	if focusableNode != nil {
		a.setFocus(nodeStyle.ID, root)
	}

	handled := false
	if handler, ok := clickedNode.(EventHandler); ok {
		state := a.componentStates[clickedNode.GetStyle().ID]
		res := findLayoutResultByID(searchLayout, clickedNode.GetStyle().ID)
		var clickedLayout LayoutResult
		if res != nil {
			clickedLayout = *res
		}
		eventCtx := EventContext{
			Layout: clickedLayout,
		}
		if handler.HandleEvent(ev, state, eventCtx) {
			a.dirty = true
		}
		handled = true
	}
	return handled
}
