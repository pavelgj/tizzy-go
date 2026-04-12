# Creating New Components

This guide explains how to create custom components in Tizzy and how to test them.

## The Node Interface

Every component in Tizzy implements the `Node` interface defined in `tz/node.go`:

```go
type Node interface {
    GetStyle() Style
}
```

That is the only _required_ interface. All other behaviours — layout, rendering, event handling, focus, cursor management, overlays — are expressed through **optional interfaces** that the framework detects at runtime. Implement only what your component needs.

## Optional Interfaces

| Interface | Method(s) | When to implement |
|---|---|---|
| `Layoutable` | `Layout(x, y int, c Constraints) LayoutResult` | Component has custom size/position logic |
| `Renderable` | `Render(grid, layout, focusedID, componentStates)` | Component draws itself |
| `ParentNode` | `GetChildren() []Node` | Component has child nodes |
| `Focusable` | `IsFocusable() bool` | Component can receive keyboard focus |
| `EventHandler` | `HandleEvent`, `DefaultState` | Component handles keyboard or mouse events |
| `FocusScope` | `FocusableChildren(states) []Node` | Component controls which children participate in Tab traversal (e.g. Tabs) |
| `FocusGainHandler` | `OnFocusGained(state any)` | Component reacts when it gains focus |
| `CursorProvider` | `UpdateScrollOffset`, `GetCursorPosition` | Component manages the hardware terminal cursor (e.g. TextInput) |
| `Dismissable` | `Dismiss(state any)` | Component resets/closes when another component gains focus (e.g. Dropdown) |
| `CustomHitTester` | `FindNodePathAt(x, y, res, states) []Node` | Component overrides mouse hit-testing for its children |

---

## Step-by-Step Guide

### 1. Define the Struct

```go
type MyComponent struct {
    Style   Style
    Label   string
    OnClick func()
}
```

### 2. Implement `GetStyle`

```go
func (m *MyComponent) GetStyle() Style {
    return m.Style
}
```

### 3. Write a Constructor

- If the component is **stateless**, no `RenderContext` is needed.
- If the component uses **hooks** (state, effects) or needs an auto-generated ID, accept `ctx *RenderContext` as the first argument.

```go
// Stateless
func NewMyComponent(style Style, label string, onClick func()) *MyComponent {
    return &MyComponent{Style: style, Label: label, OnClick: onClick}
}

// Stateful — needs a unique ID for focus and state storage
func NewMyStatefulComponent(ctx *RenderContext, style Style, label string) *MyStatefulComponent {
    stateObj, _ := UseState(ctx, &MyState{})
    if style.ID == "" {
        style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex-1)
    }
    return &MyStatefulComponent{Style: style, Label: label, state: stateObj.(*MyState)}
}
```

> **Hook ordering rule**: `UseState` and `UseEffect` must be called in the same order every render — no calls inside `if` blocks or loops. One hook slot is always consumed so that sibling components keep stable indices.

### 4. Implement `Layoutable`

The `Layout` method calculates the component's position and dimensions.

```go
func (m *MyComponent) Layout(x, y int, c Constraints) LayoutResult {
    w := len(m.Label) + m.Style.Padding.Left + m.Style.Padding.Right
    h := 1
    if m.Style.Width > 0 {
        w = m.Style.Width
    }
    return LayoutResult{
        Node: m,
        X:    x + m.Style.Margin.Left,
        Y:    y + m.Style.Margin.Top,
        W:    w,
        H:    h,
    }
}
```

### 5. Implement `Renderable`

The `Render` method draws the component to the `Grid` using the layout result from step 4.

```go
func (m *MyComponent) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
    focused := m.Style.ID != "" && m.Style.ID == focusedID
    style := tcell.StyleDefault.Foreground(m.Style.Color).Background(m.Style.Background)
    if focused {
        style = style.Foreground(tcell.ColorYellow)
    }
    drawText(grid, layout.X+m.Style.Padding.Left, layout.Y, m.Label, style)
}
```

### 6. Implement `EventHandler` (interactive components)

```go
func (m *MyComponent) DefaultState() any {
    return nil // or &MyState{} if the component has internal state
}

func (m *MyComponent) HandleEvent(ev tcell.Event, state any, ctx EventContext) bool {
    key, ok := ev.(*tcell.EventKey)
    if !ok {
        return false
    }
    if key.Key() == tcell.KeyEnter && m.OnClick != nil {
        m.OnClick()
        return true // return true = re-render needed
    }
    return false
}
```

---

## Creating Overlay Components

Use a **Portal** when your component needs to render content above the main UI at an absolute screen position (dialog, tooltip, completion popup).

```go
func NewMyDialog(ctx *RenderContext, style Style, content Node, isOpen bool) Node {
    // Consume one hook slot so sibling hooks keep stable indices.
    id := fmt.Sprintf("hook-%d", ctx.hookIndex)
    ctx.hookIndex++

    if !isOpen {
        return nil // NewBox drops nil children automatically
    }

    if style.ID == "" {
        style.ID = id
    }

    return &Portal{
        Style:     style,
        Child:     NewBox(Style{Border: true, Background: style.Background}, content),
        X:         -1, // -1 = auto-center on screen
        Y:         -1,
        TrapFocus: true, // restrict Tab to dialog content
    }
}
```

For overlays whose position depends on the triggering component's screen position (like a dropdown list appearing below a button), use `PositionFn`:

```go
myID := style.ID
portal := &Portal{
    Child: listNode,
    PositionFn: func(screenW, screenH int, mainLayout LayoutResult) (x, y, maxW, maxH int) {
        res := findLayoutResultByID(mainLayout, myID)
        if res == nil {
            return 0, 0, screenW, screenH
        }
        return res.X, res.Y + res.H, res.W, screenH - (res.Y + res.H)
    },
    OnOutsideClick: func() { state.Open = false },
}
```

See `tz/dropdown.go` and `tz/menubar.go` for complete Portal-based overlay examples.

---

## Testing Components

### Layout Tests

```go
func TestMyComponentLayout(t *testing.T) {
    comp := NewMyComponent(Style{}, "Hello", nil)
    res := Layout(comp, 0, 0, Constraints{MaxW: 80, MaxH: 24})

    if res.W != 5 {
        t.Errorf("expected width 5, got %d", res.W)
    }
    if res.H != 1 {
        t.Errorf("expected height 1, got %d", res.H)
    }
}
```

### Render Tests

Use `tcell.NewSimulationScreen` and the `renderToScreen` helper (defined in `tz/render_test.go`):

```go
func TestMyComponentRender(t *testing.T) {
    s := tcell.NewSimulationScreen("")
    s.Init()
    s.SetSize(20, 2)

    comp := NewMyComponent(Style{}, "Hello", nil)
    layout := Layout(comp, 0, 0, Constraints{MaxW: 20, MaxH: 2})
    renderToScreen(s, layout, "", nil)
    s.Show()

    mainc, _, _, _ := s.GetContent(0, 0)
    if mainc != 'H' {
        t.Errorf("expected 'H' at 0,0, got %q", mainc)
    }
}
```

### Visual Regression Tests

For complex components, add a visual regression test in `tz/visual_test.go`. See `CONTRIBUTING.md` for how the golden-image workflow works.

### Event Tests

```go
func TestMyComponentHandleEvent(t *testing.T) {
    clicked := false
    comp := NewMyComponent(Style{ID: "c"}, "Click me", func() { clicked = true })

    app := &App{componentStates: make(map[string]any)}
    ctx := EventContext{Layout: LayoutResult{Node: comp, W: 10, H: 1}}

    ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
    handled := comp.HandleEvent(ev, nil, ctx)

    if !handled {
        t.Error("expected HandleEvent to return true")
    }
    if !clicked {
        t.Error("expected OnClick to be called")
    }
    _ = app
}
```
