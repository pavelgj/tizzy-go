package tz

import (
	"fmt"
	"os"
	"strings"
	"sync"
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
	mu              sync.Mutex
}

// NewApp creates a new App instance.
func NewApp() (*App, error) {
	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	return NewAppWithScreen(s), nil
}

func NewAppWithScreen(s tcell.Screen) *App {
	return &App{
		screen:          s,
		componentStates: make(map[string]any),
		activeCleanups:  make(map[string]func()),
		dirty:           true,
	}
}

func (a *App) RenderFrame(renderFn func(ctx *RenderContext) Node) (*Grid, Node, LayoutResult, []string, error) {
	// 1. Get the current UI tree or reuse previous
	var root Node
	if a.dirty || a.previousRoot == nil {
		ctx := &RenderContext{app: a}
		root = renderFn(ctx)
		if err := validateUniqueIDs(root); err != nil {
			return nil, nil, LayoutResult{}, nil, err
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
	a.mu.Lock()
	statesCopy := make(map[string]any, len(a.componentStates))
	for k, v := range a.componentStates {
		statesCopy[k] = v
	}
	a.mu.Unlock()

	for id, stateObj := range statesCopy {
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

	// Update scroll offset for focused cursor provider based on up-to-date layout
	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if provider, ok := focusedNode.(CursorProvider); ok {
			res := findLayoutResultByID(layout, a.focusedID)
			if res != nil {
				provider.UpdateScrollOffset(*res, a.componentStates[a.focusedID])
			}
		}
	}

	// 3. Render to grid
	grid := NewGrid(w, h)
	Render(grid, layout, a.focusedID, statesCopy)

	// Render overlays: walk the tree and let each component render itself.
	_ = walkTree(root, func(n Node) error {
		id := n.GetStyle().ID
		if id == "" {
			return nil
		}
		stateObj, ok := statesCopy[id]
		if !ok {
			return nil
		}
		openable, ok := stateObj.(OpenableState)
		if !ok || !openable.IsOpen() {
			return nil
		}
		if overlay, ok := n.(OverlayRenderer); ok {
			overlay.RenderOverlay(grid, w, h, layout, a.focusedID, statesCopy)
		}
		return nil
	})

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

	// Handle terminal cursor for focused component
	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if provider, ok := focusedNode.(CursorProvider); ok {
			var res *LayoutResult
			for _, stateObj := range a.componentStates {
				if state, ok := stateObj.(*ModalState); ok && state.Open {
					res = findLayoutResultByID(state.OverlayLayout, a.focusedID)
					if res != nil {
						break
					}
				}
			}
			if res == nil {
				res = findLayoutResultByID(layout, a.focusedID)
			}
			if res != nil {
				cx, cy, show := provider.GetCursorPosition(*res, a.componentStates[a.focusedID])
				if show {
					a.screen.ShowCursor(cx, cy)
				} else {
					a.screen.HideCursor()
				}
			} else {
				a.screen.HideCursor()
			}
		} else {
			a.screen.HideCursor()
		}
	} else {
		a.screen.HideCursor()
	}

	a.screen.Show()

	return grid, root, layout, focusableIDs, nil
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
		_, root, layout, focusableIDs, err := a.RenderFrame(renderFn)
		if err != nil {
			return err
		}

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

	if scope, ok := node.(FocusScope); ok {
		for _, child := range scope.FocusableChildren(componentStates) {
			ids = append(ids, findFocusableIDs(child, componentStates)...)
		}
	} else if p, ok := node.(ParentNode); ok {
		for _, child := range p.GetChildren() {
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

func findNodePathAt(res LayoutResult, x, y int, componentStates map[string]any) []Node {
	if x >= res.X && x < res.X+res.W && y >= res.Y && y < res.Y+res.H {
		if hitTester, ok := res.Node.(CustomHitTester); ok {
			return hitTester.FindNodePathAt(x, y, res, componentStates)
		}

		for _, child := range res.Children {
			if path := findNodePathAt(child, x, y, componentStates); path != nil {
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
		if handler, ok := node.(FocusGainHandler); ok {
			stateObj := a.componentStates[id]
			handler.OnFocusGained(stateObj)
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

			var res *LayoutResult
			for _, stateObj := range a.componentStates {
				if state, ok := stateObj.(*ModalState); ok && state.Open {
					res = findLayoutResultByID(state.OverlayLayout, a.focusedID)
					if res != nil {
						break
					}
				}
			}
			if res == nil {
				res = findLayoutResultByID(layout, a.focusedID)
			}
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
