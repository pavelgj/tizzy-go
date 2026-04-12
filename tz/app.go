package tz

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
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
	dirty           atomic.Bool
	previousRoot    Node
	portals         []collectedPortal // portals collected after each layout pass
	mu              sync.Mutex
	scheduler       *animationScheduler
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
	a := &App{
		screen:          s,
		componentStates: make(map[string]any),
		activeCleanups:  make(map[string]func()),
		scheduler:       newAnimationScheduler(),
	}
	a.dirty.Store(true)
	return a
}

func (a *App) RenderFrame(renderFn func(ctx *RenderContext) Node) (*Grid, Node, LayoutResult, []string, error) {
	// 1. Build (or reuse) the UI tree.
	var root Node
	if a.dirty.Load() || a.previousRoot == nil {
		ctx := &RenderContext{app: a}
		root = renderFn(ctx)
		if err := validateUniqueIDs(root); err != nil {
			return nil, nil, LayoutResult{}, nil, err
		}
		a.previousRoot = root
		a.dirty.Store(false)

		// Collect all effect IDs requested in this frame.
		requestedEffects := make(map[string]bool)
		for _, eff := range ctx.effects {
			requestedEffects[eff.ID] = true
		}

		// Detect unmounts and run cleanups.
		for id, cleanup := range a.activeCleanups {
			if !requestedEffects[id] {
				if cleanup != nil {
					cleanup()
				}
				delete(a.activeCleanups, id)
			}
		}

		// Detect mounts and run effects.
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

	a.mu.Lock()
	statesCopy := make(map[string]any, len(a.componentStates))
	for k, v := range a.componentStates {
		statesCopy[k] = v
	}
	a.mu.Unlock()

	// 2. Layout the main tree.
	w, h := a.screen.Size()
	layout := Layout(root, 0, 0, Constraints{MaxW: w, MaxH: h})

	// 3. Collect portals from the layout tree (zero-size stubs in main layout,
	//    but Portal.GetChildren lets findNodeByID reach inside them).
	var portals []collectedPortal
	collectPortals(layout, w, h, layout, &portals)
	a.portals = portals

	// 4. Determine focusable IDs.  If a TrapFocus portal is present only its
	//    subtree participates in Tab traversal.
	var trapFocusPortal *Portal
	for i := len(portals) - 1; i >= 0; i-- {
		if portals[i].portal.TrapFocus {
			trapFocusPortal = portals[i].portal
			break
		}
	}

	var focusableIDs []string
	if trapFocusPortal != nil {
		focusableIDs = findFocusableIDs(trapFocusPortal.Child, a.componentStates)
		// Ensure the current focus is inside the portal; if not, move it in.
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

	// 5. Update scroll offset for the focused CursorProvider.
	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if provider, ok := focusedNode.(CursorProvider); ok {
			res := a.findFocusedLayout(layout, a.focusedID)
			if res != nil {
				provider.UpdateScrollOffset(*res, a.componentStates[a.focusedID])
			}
		}
	}

	// 6. Render main tree.
	grid := NewGrid(w, h)
	Render(grid, layout, a.focusedID, statesCopy)

	// 7. Render portal content (in Z-order; later portals appear on top).
	for _, cp := range portals {
		Render(grid, cp.content, a.focusedID, statesCopy)
	}

	// 8. Diff and flush to the terminal.
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

	// 10. Position the terminal hardware cursor for the focused CursorProvider.
	if a.focusedID != "" {
		focusedNode := findNodeByID(root, a.focusedID)
		if provider, ok := focusedNode.(CursorProvider); ok {
			res := a.findFocusedLayout(layout, a.focusedID)
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

// findFocusedLayout searches the main layout tree and all portal content
// layouts for the LayoutResult of the currently-focused node.
func (a *App) findFocusedLayout(mainLayout LayoutResult, id string) *LayoutResult {
	// Check portal layouts first (they sit on top).
	for i := len(a.portals) - 1; i >= 0; i-- {
		if res := findLayoutResultByID(a.portals[i].content, id); res != nil {
			return res
		}
	}
	// Fall back to the main tree layout.
	return findLayoutResultByID(mainLayout, id)
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

// Run starts the application loop.
func (a *App) Run(renderFn func(ctx *RenderContext) Node, updateFn func(tcell.Event)) error {
	if err := a.screen.Init(); err != nil {
		return err
	}
	defer a.screen.Fini()

	a.screen.EnableMouse()

	a.screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorReset).Foreground(tcell.ColorReset))

	// Ensure a scheduler exists even when App is constructed directly in tests.
	if a.scheduler == nil {
		a.scheduler = newAnimationScheduler()
	}
	// Single goroutine drives all UseAnimation/UseTween animations. Step
	// functions call state setters → MarkDirty → PostEvent(EventTick).
	go a.scheduler.run(10)

	for {
		_, root, layout, focusableIDs, err := a.RenderFrame(renderFn)
		if err != nil {
			return err
		}

		ev := a.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			if a.handleKeyEvent(ev, root, layout, focusableIDs) {
				return nil
			}
		case *EventTick:
			// Continuous rendering for animations.
		case *tcell.EventMouse:
			a.handleMouseEvent(ev, root, layout)
		case *tcell.EventResize:
			a.screen.Sync()
		}

		if updateFn != nil {
			updateFn(ev)
		}
	}
}

// dismissOthers calls Dismiss on every Dismissable component except the one
// that just gained focus. This lets components like Dropdown close themselves
// when the user focuses something else.
func (a *App) dismissOthers(keepID string, root Node) {
	_ = walkTree(root, func(n Node) error {
		id := n.GetStyle().ID
		if id == "" || id == keepID {
			return nil
		}
		if d, ok := n.(Dismissable); ok {
			if state, ok := a.componentStates[id]; ok {
				d.Dismiss(state)
			}
		}
		return nil
	})
}

type EventTick struct {
	t time.Time
}

func (e *EventTick) When() time.Time {
	return e.t
}

// MarkDirty forces a re-render on the next frame. Safe to call from any
// goroutine. When called from a background goroutine (e.g. an animation step
// function) it posts an EventTick the first time the flag transitions from
// clean to dirty, waking up PollEvent promptly without flooding the queue.
func (a *App) MarkDirty() {
	if !a.dirty.Swap(true) && a.screen != nil {
		_ = a.screen.PostEvent(&EventTick{t: time.Now()})
	}
}

func (a *App) setFocus(id string, root Node) {
	if a.focusedID == id {
		return
	}
	a.focusedID = id
	a.dirty.Store(true)

	a.dismissOthers(id, root)

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

	// Determine whether any popup-mode portal is open (used to suppress
	// underlying TextInput navigation keys).
	popupOpen := false
	for _, cp := range a.portals {
		if cp.portal.PopupMode {
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

			layoutRes := LayoutResult{}
			if res := a.findFocusedLayout(layout, a.focusedID); res != nil {
				layoutRes = *res
			}

			eventCtx := EventContext{
				Layout:    layoutRes,
				PopupOpen: popupOpen,
			}

			if handler.HandleEvent(ev, state, eventCtx) {
				a.dirty.Store(true)
				return false
			}
		}
	}

	if ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC {
		// When any overlay is open, let Escape fall through to updateFn
		// (component or user code handles closing it) rather than exiting.
		if ev.Key() == tcell.KeyEscape && len(a.portals) > 0 {
			return false
		}
		return true // request exit
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
	if a.handlePortalMouseEvent(ev, mx, my, root) {
		return true
	}
	return a.handleMainTreeMouseEvent(ev, mx, my, root, layout)
}

// handlePortalMouseEvent routes mouse events through portals in Z-order
// (topmost portal first). Returns true if the event was consumed.
func (a *App) handlePortalMouseEvent(ev MouseEvent, mx, my int, root Node) bool {
	for i := len(a.portals) - 1; i >= 0; i-- {
		cp := a.portals[i]
		p := cp.portal
		content := cp.content

		inside := mx >= content.X && mx < content.X+content.W &&
			my >= content.Y && my < content.Y+content.H

		if ev.Buttons()&tcell.Button1 != 0 {
			if inside {
				if tcellEv, ok := ev.(tcell.Event); ok {
					path := findNodePathAt(content, mx, my, a.componentStates)
					if len(path) > 0 {
						a.dispatchEventToPath(path, tcellEv, root, content)
					}
				}
				a.dirty.Store(true)
			} else if p.OnOutsideClick != nil {
				p.OnOutsideClick()
				a.dirty.Store(true)
			}
			return true
		}

		if ev.Buttons()&(tcell.WheelUp|tcell.WheelDown) != 0 && inside {
			if tcellEv, ok := ev.(tcell.Event); ok {
				path := findNodePathAt(content, mx, my, a.componentStates)
				if len(path) > 0 {
					targetNode := path[len(path)-1]
					if handler, ok := targetNode.(EventHandler); ok {
						state := a.componentStates[targetNode.GetStyle().ID]
						res := findLayoutResultByID(content, targetNode.GetStyle().ID)
						var targetLayout LayoutResult
						if res != nil {
							targetLayout = *res
						}
						if handler.HandleEvent(tcellEv, state, EventContext{Layout: targetLayout}) {
							a.dirty.Store(true)
						}
					}
				}
			}
			return true
		}
	}
	return false
}

// handleMainTreeMouseEvent routes mouse events through the main layout tree.
// Note: Tabs.FindNodePathAt (CustomHitTester) already filters inactive tab
// content, so no Tabs-specific guard is needed here.
func (a *App) handleMainTreeMouseEvent(ev MouseEvent, mx, my int, root Node, layout LayoutResult) bool {
	if ev.Buttons()&tcell.Button1 != 0 {
		path := findNodePathAt(layout, mx, my, a.componentStates)
		if len(path) > 0 {
			debugLog(fmt.Sprintf("Mouse click at %d,%d. Path len: %d", mx, my, len(path)))
			for i, n := range path {
				debugLog(fmt.Sprintf("  Path[%d]: %T", i, n))
			}
			if tcellEv, ok := ev.(tcell.Event); ok {
				a.dispatchEventToPath(path, tcellEv, root, layout)
			}
		}
	} else if ev.Buttons()&(tcell.Button4|tcell.WheelUp|tcell.Button5|tcell.WheelDown) != 0 {
		path := findNodePathAt(layout, mx, my, a.componentStates)
		if len(path) > 0 {
			a.dispatchWheelToPath(path, ev, layout)
		}
	}
	return true
}

// dispatchWheelToPath delivers a wheel event to the deepest node in the path
// that handles it, bubbling up toward the root until one returns true.
func (a *App) dispatchWheelToPath(path []Node, ev MouseEvent, searchLayout LayoutResult) {
	tcellEv, ok := ev.(tcell.Event)
	if !ok {
		return
	}
	for i := len(path) - 1; i >= 0; i-- {
		targetNode := path[i]
		handler, ok := targetNode.(EventHandler)
		if !ok {
			continue
		}
		state := a.componentStates[targetNode.GetStyle().ID]
		res := findLayoutResultByID(searchLayout, targetNode.GetStyle().ID)
		var targetLayout LayoutResult
		if res != nil {
			targetLayout = *res
		}
		if handler.HandleEvent(tcellEv, state, EventContext{Layout: targetLayout}) {
			a.dirty.Store(true)
			return
		}
	}
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
		if state == nil {
			state = handler.DefaultState()
			a.mu.Lock()
			a.componentStates[clickedNode.GetStyle().ID] = state
			a.mu.Unlock()
		}
		res := findLayoutResultByID(searchLayout, clickedNode.GetStyle().ID)
		var clickedLayout LayoutResult
		if res != nil {
			clickedLayout = *res
		}
		eventCtx := EventContext{
			Layout: clickedLayout,
		}
		if handler.HandleEvent(ev, state, eventCtx) {
			a.dirty.Store(true)
		}
		handled = true
	}
	return handled
}
