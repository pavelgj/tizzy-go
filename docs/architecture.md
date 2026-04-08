# Splotch Architecture and Philosophy

Splotch is a declarative Terminal User Interface (TUI) library for Go. It brings the component model and state management patterns familar from modern web frameworks (like React) to the world of terminal applications.

## Philosophy

1.  **Declarative over Imperative**: You describe *what* the UI should look like based on the current state, not *how* to transition the UI from one state to another. Splotch handles the rendering and updating.
2.  **Component-Based**: Interfaces are built by composing small, reusable components. Complex UIs are broken down into manageable, independent pieces.
3.  **Hooks for State and Lifecycle**: State management and side effects are handled using hooks (`UseState`, `UseEffect`) within functional components, keeping state close to where it is used.
4.  **Flexbox-inspired Layout**: Layout is determined by containers (like `Box`) that distribute space using familiar concepts like flexDirection, justifyContent, and fill.

## Core Architecture

Splotch operates on a reactive render loop controlled by the `App` struct.

### 1. The Component Tree (Nodes)

The fundamental building block is the `Node` interface. Every component (whether a primitive like `Text` or a complex composite like `Tabs`) implements this interface.

```go
type Node interface {
    Layout(x, y int, c Constraints) LayoutResult
    Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any)
    GetStyle() Style
}
```

Components are created in the `Render` function, forming a virtual tree of nodes describing the desired UI.

### 2. Render Context and Hooks

The `RenderContext` is passed to the main render function. It is the gateway to Splotch's hooks system:

-   **`UseState`**: Allows functional components to persist state across render cycles. State is keyed by the order of hook calls.
-   **`UseEffect`**: Allows components to perform side effects (e.g., starting background timers, fetching data) when they are first rendered ("mount") and clean up when they are no longer rendered ("unmount").

### 3. Dual-Pass Layout and Rendering

Every frame in Splotch goes through two distinct phases:

1.  **Layout Phase**: Splotch traverses the component tree calling `Layout()` on each node. Nodes calculate their dimensions based on `Constraints` passed from their parents and return a `LayoutResult`.
2.  **Render Phase**: Splotch traverses the tree again, calling `Render()`. Nodes draw themselves to the `Grid` (a wrapper around `tcell.Screen`) based on the coordinates calculated in the layout phase.

### 4. State Persistence and Re-renders

State created via `UseState` is stored centrally in the `App` struct, keyed by auto-generated sequential IDs. When a state setter is called, `App` marks itself as `dirty`. On the next tick, if the app is dirty, it triggers a re-render by calling the user-provided render function to build a new virtual tree.

## Comparison with Other Go TUI Libraries

Most traditional Go TUI libraries (like termui or tview) are object-oriented and imperative: you create long-lived widget instances and manually update their properties.

Libraries like Bubble Tea use the Elm Architecture. While also declarative, Bubble Tea centralizes all state and message routing in a single continuous model, which can lead to boilerplate in large applications.

Splotch takes a middle ground: it is declarative like Bubble Tea, but uses localized state and hooks like React, making it easier to build highly interactive, modular components without a massive central model.
