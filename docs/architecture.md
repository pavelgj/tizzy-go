# Tizzy Architecture and Philosophy

Tizzy is a declarative Terminal User Interface (TUI) library for Go. It brings the component model and state-management patterns familiar from modern web frameworks (like React) to the world of terminal applications.

## Philosophy

1. **Declarative over Imperative**: You describe *what* the UI should look like based on current state, not *how* to transition between states. Tizzy handles rendering and updates.
2. **Component-Based**: UIs are built by composing small, reusable components. Complex screens are broken down into independent, manageable pieces.
3. **Hooks for State and Lifecycle**: State and side effects are handled with hooks (`UseState`, `UseEffect`) inside the render function, keeping state close to where it is used.
4. **Flexbox-Inspired Layout**: Space distribution follows familiar concepts: `FlexDirection`, `JustifyContent`, `FillWidth`/`FillHeight`.
5. **Clean Core / Component Separation**: The framework core (`app.go`) is generic. All component-specific logic (overlay rendering, focus scoping, cursor positioning) lives in the component files, exposed to the framework through narrow interfaces.

---

## Core Architecture

Tizzy operates on a reactive render loop controlled by the `App` struct.

### 1. The Node Interface

The fundamental building block is the `Node` interface (`tz/node.go`):

```go
type Node interface {
    GetStyle() Style
}
```

`Node` is intentionally minimal. Optional behaviours are expressed through additional interfaces that the framework checks at runtime:

| Interface | Purpose |
|---|---|
| `Layoutable` | Node computes its own layout |
| `Renderable` | Node draws itself to the grid |
| `ParentNode` | Node has child nodes (`GetChildren`) |
| `Focusable` | Node can receive keyboard focus |
| `EventHandler` | Node handles keyboard/mouse events |
| `FocusScope` | Node restricts which children participate in Tab traversal |
| `FocusGainHandler` | Node reacts when it gains focus |
| `CursorProvider` | Node manages the terminal hardware cursor |
| `CustomHitTester` | Node overrides mouse hit-testing for its children |
| `Dismissable` | Node closes/resets when another component gains focus |

New components only implement the interfaces they actually need. Adding a component never requires editing the core framework.

### 2. The Render Context and Hooks

A `RenderContext` is passed to the root render function each frame. It is the gateway to the hooks system:

- **`UseState(ctx, initial)`**: Persists state across renders, keyed by call order. Returns the current value and a setter.
- **`UseEffect(ctx, fn)`**: Runs a function on mount; its return value is called on unmount. Used for timers, subscriptions, etc.

State is stored centrally in `App.componentStates` (a `map[string]any`), keyed by auto-generated sequential IDs (`hook-0`, `hook-1`, …). Hook call order must be stable across renders — the same rules as React hooks apply.

### 3. The RenderFrame Pipeline

Every frame executed by `App.RenderFrame` goes through these steps in order:

```
1. Build tree        renderFn(ctx) → Node tree (or reuse previous if !dirty)
2. Layout            Layout(root) → LayoutResult tree (positions + sizes)
3. Collect portals   Walk LayoutResult, find *Portal nodes, compute their real layouts
4. Focusable IDs     Walk node tree; if a TrapFocus portal exists, scope to its subtree
5. Scroll offset     CursorProvider.UpdateScrollOffset for the focused node
6. Render            Render(grid, layout) for the main tree
7. Render portals    Render each portal's content layout (Z-ordered, last on top)
8. Diff & flush      Compare grid against previousGrid, push changed cells to tcell
9. Cursor            CursorProvider.GetCursorPosition → screen.ShowCursor / HideCursor
```

### 4. The Portal Mechanism

Overlays that sit on top of the main UI (modals, popups, dropdown lists) are modelled as **Portal nodes**:

```go
type Portal struct {
    Style          Style
    Child          Node
    X, Y           int      // absolute position; X==-1 → auto-center
    PositionFn     func(screenW, screenH int, mainLayout LayoutResult) (x, y, maxW, maxH int)
    OnOutsideClick func()   // called when user clicks outside the portal bounds
    TrapFocus      bool     // restricts Tab traversal to nodes inside the portal
    PopupMode      bool     // tells TextInput to suppress Enter/Up/Down keys
}
```

Key properties:

- A Portal produces **zero size** in the main layout tree — it does not displace sibling nodes.
- The framework walks the main `LayoutResult` tree to collect all portals, then computes each portal's real screen layout (applying `PositionFn` or centering) separately.
- Portal content is rendered in Z-order after the main pass.
- **Mouse events** check portals top-to-bottom before the main tree. Clicks inside the topmost portal are dispatched to its content; clicks outside trigger `OnOutsideClick`.
- **`TrapFocus: true`** causes the framework to compute focusable IDs from the portal's subtree only, preventing Tab from reaching nodes behind the overlay.
- **`PopupMode: true`** sets `EventContext.PopupOpen = true` for the current frame so underlying `TextInput` components know to suppress navigation keys.

`NewModal` and `NewPopup` are convenience constructors that return a Portal (or `nil` when closed). Because `NewBox` silently drops `nil` children, callers can always include them unconditionally:

```go
return tz.NewBox(style,
    mainContent,
    tz.NewModal(ctx, modalStyle, modalBody, isOpen),  // nil when closed
    tz.NewPopup(ctx, popupStyle, popupBody, x, y, isOpen),
)
```

### 5. Event Routing

**Keyboard events** flow to the currently focused component via `EventHandler.HandleEvent`. If the component does not handle the event (returns `false`), the framework processes Tab/Shift-Tab focus movement and Escape (which suppresses app exit if any portal is open).

**Mouse click events** are routed in Z-order:

1. Portals (topmost first): inside → dispatch to content; outside → `OnOutsideClick`.
2. Main tree: `findNodePathAt` walks the `LayoutResult` tree to build a path of nodes under the cursor, then `dispatchEventToPath` sends the event to the leaf node.

**Wheel events** follow the same Z-order but always dispatch to the leaf node under the cursor.

### 6. Focus Management

Focusable IDs are collected each frame by walking the node tree via `ParentNode.GetChildren`. The walk respects:

- `FocusScope.FocusableChildren` — lets a component restrict which children are reachable (used by `Tabs` to expose only the active tab's content).
- `*Portal` nodes — skipped during normal traversal; when a `TrapFocus` portal is present, only that portal's subtree is searched.

When focus changes, `FocusGainHandler.OnFocusGained` is called on the newly focused node (used by `List` to call its `OnFocus` callback).

---

## File Map

| File | Role |
|---|---|
| `tz/app.go` | `App`, `RenderFrame`, event loop, focus management |
| `tz/node.go` | Core interfaces, `Style`, `Box` |
| `tz/layout.go` | `Layout()` dispatcher, `LayoutResult`, `Constraints` |
| `tz/render.go` | `Render()` dispatcher, `drawBorder`, `drawText` |
| `tz/portal.go` | `Portal` node, `collectPortals`, `computePortalLayout` |
| `tz/modal.go` | `NewModal` factory (returns Portal or nil) |
| `tz/popup.go` | `NewPopup` factory (returns Portal or nil) |
| `tz/dropdown.go` | `Dropdown` widget + Portal-based list overlay |
| `tz/menubar.go` | `MenuBar` widget + Portal-based menu overlay |
| `tz/text_input.go` | `TextInput`, implements `CursorProvider` |
| `tz/list.go` | `List`, implements `FocusGainHandler` |
| `tz/tabs.go` | `Tabs`, implements `FocusScope` + `CustomHitTester` |
| `tz/grid.go` | In-memory cell grid written to `tcell.Screen` |

---

## Comparison with Other Go TUI Libraries

Most traditional Go TUI libraries (like termui or tview) are object-oriented and imperative: you create long-lived widget instances and manually update their properties.

Bubble Tea uses the Elm Architecture. While also declarative, it centralises all state and message routing in a single model, which can create boilerplate in large applications.

Tizzy takes a middle ground: declarative like Bubble Tea, but with localised state and hooks like React, making it easier to build highly interactive, modular components without a massive central update function.
