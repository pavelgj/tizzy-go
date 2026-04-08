# Creating New Components

This guide explains how to create and test new components in Splotch.

## The Node Interface

All components in Splotch must implement the `Node` interface defined in `splotch/app.go`:

```go
type Node interface {
    Layout(x, y int, c Constraints) LayoutResult
    Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any)
    GetStyle() Style
    node() // Marker method
}
```

## Step-by-Step Guide

### 1. Define the Struct

Create a struct that holds the properties (props) of your component. It should usually include a `Style` field.

```go
type MyComponent struct {
    Style Style
    Label string
    // Other props...
}
```

### 2. Implement the Marker Method

Implement the empty `node()` method to satisfy the interface.

```go
func (m *MyComponent) node() {}
```

### 3. Create a Constructor

Provide a constructor function.
- If the component is **stateless** (controlled), it doesn't need `RenderContext`.
- If the component uses **hooks** (internal state or effects), it **must** accept `ctx *RenderContext` as its first argument.

```go
// Stateless example
func NewMyComponent(style Style, label string) *MyComponent {
    return &MyComponent{Style: style, Label: label}
}

// Stateful example (using hooks or auto-generating IDs)
func NewMyStatefulComponent(ctx *RenderContext, style Style, label string) *MyStatefulComponent {
    // Auto-generate ID if not provided (required for focusable components)
    if style.ID == "" {
        style.ID = fmt.Sprintf("hook-%d", ctx.hookIndex)
        ctx.hookIndex++
    }
    
    // You can also call UseState here if needed
    
    return &MyStatefulComponent{Style: style, Label: label}
}
```

### 4. Implement Layout

The `Layout` method calculates the size and position of the component. It receives the suggested `x` and `y` coordinates and layout `Constraints`.

```go
func (m *MyComponent) Layout(x, y int, c Constraints) LayoutResult {
    // Calculate width and height based on content and constraints
    w := len(m.Label)
    h := 1
    
    // Apply styles (padding, border, width overrides)
    // ...
    
    return LayoutResult{
        Node: m,
        X:    x + m.Style.Margin.Left,
        Y:    y + m.Style.Margin.Top,
        W:    w,
        H:    h,
    }
}
```

### 5. Implement Render

The `Render` method draws the component to the `Grid`.

```go
func (m *MyComponent) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
    focused := m.Style.ID != "" && m.Style.ID == focusedID
    
    style := tcell.StyleDefault.Foreground(m.Style.Color).Background(m.Style.Background)
    if focused {
        style = style.Background(tcell.ColorYellow) // Example focus style
    }
    
    // Use drawText or directly grid.SetContent
    drawText(grid, layout.X, layout.Y, m.Label, style)
}
```

## Testing Components

Tests are crucial to ensure components render correctly and handle layout constraints.

### Layout Tests

Verify that the component calculates its dimensions correctly.

```go
func TestLayoutMyComponent(t *testing.T) {
    comp := NewMyComponent(Style{}, "Test")
    res := Layout(comp, 0, 0, Constraints{MaxW: 100, MaxH: 100})
    
    if res.W != 4 {
        t.Errorf("Expected width 4, got %d", res.W)
    }
}
```

### Render Tests

Use `tcell.NewSimulationScreen` to verify the visual output. Use the `renderToScreen` helper available in the `splotch` package tests (defined in `render_test.go`) to render the node to the simulation screen.

```go
func TestRenderMyComponent(t *testing.T) {
    s := tcell.NewSimulationScreen("")
    s.Init()
    s.SetSize(20, 2)
    
    comp := NewMyComponent(Style{}, "Test")
    layout := Layout(comp, 0, 0, Constraints{MaxW: 20, MaxH: 2})
    
    // Helper from render_test.go
    renderToScreen(s, layout, "", nil)
    s.Show()
    
    // Verify content
    mainc, _, _, _ := s.GetContent(0, 0)
    if mainc != 'T' {
        t.Errorf("Expected 'T' at 0,0, got '%c'", mainc)
    }
}
```
