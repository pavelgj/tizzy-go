# Tizzy

A declarative Terminal User Interface (TUI) library for Go, inspired by React's component model and CSS Flexbox layout.

## Features

- **Declarative UI**: Build UIs by composing components that return a virtual tree of nodes.
- **Flexbox Layout**: Easy alignment and distribution of space with `row` and `column` directions, `JustifyContent`, and `FillWidth`/`FillHeight`.
- **Isolated State**: Support for struct-based components with isolated mutable state.
- **Rich Interactions**: Focus management, keyboard navigation, and mouse support.
- **Extensible**: Easy to create custom components.

## Quick Start

Here is a minimal example of a counter application:

```go
package main

import (
	"log"
	"strconv"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		count, setCount := tizzy.UseState(ctx, 0)

		return tizzy.NewBox(
			tizzy.Style{
				Border:        true,
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				FlexDirection: "column",
				FillWidth:     true,
			},
			tizzy.NewText(tizzy.Style{}, "Clicks: "+strconv.Itoa(count)),
			tizzy.NewButton(tizzy.Style{Focusable: true}, "Increment", func() {
				setCount(count + 1)
			}),
		)
	}

	if err := app.Run(render, func(ev tcell.Event) {}); err != nil {
		log.Fatal(err)
	}
}
```

## Core Concepts

### Components

Tizzy supports two patterns for components:

1.  **Functional Components**: Pure functions that take props and return a `Node`. Best for stateless or read-only UIs.
2.  **Struct Components**: Structs that hold state and have a `Render() Node` method. Best for stateful, interactive parts of the UI.

### Layout

Layout is handled by the `Box` component using a Flexbox-style system.

- `FlexDirection`: "row" or "column".
- `FillWidth` / `FillHeight`: Fill available space.
- `JustifyContent`: "flex-start", "center", "flex-end".

## Styling

All components take a `Style` struct to define their appearance and layout.

### Style Properties

- `ID` (`string`): Unique identifier for the component. Required for stateful hooks and focus management.
- `Focusable` (`bool`): If true, the component can receive focus.
- `Multiline` (`bool`): For text components, allows text to wrap.
- `Width` (`int`): Fixed width in cells.
- `Height` (`int`): Fixed height in cells.
- `MaxHeight` (`int`): Maximum height in cells.
- `FlexDirection` (`string`): "row" or "column" (default). Used by `Box`.
- `JustifyContent` (`string`): "flex-start", "center", "flex-end". Used by `Box`.
- `Border` (`bool`): If true, draws a border around the component.
- `Padding` (`Padding`): Inward spacing.
- `Margin` (`Margin`): Outward spacing.
- `Color` (`tcell.Color`): Foreground color.
- `Background` (`tcell.Color`): Background color.
- `FocusColor` (`tcell.Color`): Foreground color when focused. Falls back to yellow if not set.
- `FocusBackground` (`tcell.Color`): Background color when focused.
- `FillWidth` (`bool`): If true, fills available width.
- `FillHeight` (`bool`): If true, fills available height.
- `GridRow`, `GridCol`, `GridRowSpan`, `GridColSpan` (`int`): Used by `GridBox` layout.

## Hooks and Lifecycle

Tizzy supports React-like hooks for state management and lifecycle effects within the render function.

### UseState

Allows components to have local state that persists across renders. Tizzy provides a generic wrapper for type safety.

```go
app.Run(func(ctx *tizzy.RenderContext) tizzy.Node {
    count, setCount := tizzy.UseState(ctx, 0) // T is inferred as int

    return tizzy.NewButton(tizzy.Style{}, "Clicks: "+strconv.Itoa(count), func() {
        setCount(count + 1)
    })
}, func(ev tcell.Event) {})
```

### UseEffect

Allows performing side effects (like starting background tasks) when a component mounts, and cleaning up when it unmounts.

```go
app.Run(func(ctx *tizzy.RenderContext) tizzy.Node {
    ctx.UseEffect(func() func() {
        // OnInit / OnMount
        // Start a background goroutine or timer...

        return func() {
            // OnUnmount
            // Stop the goroutine or cleanup resources...
        }
    })

    return tizzy.NewText(tizzy.Style{}, "I keep track of my lifecycle")
}, func(ev tcell.Event) {})
```

> [!NOTE]
> `UseEffect` relies on call order to identify effects across renders. Do not call hooks inside loops or conditions.

## Components Reference

Here are all the available components in Tizzy:

### Containers & Layout

#### Box

The fundamental layout container.

```go
tizzy.NewBox(
    tizzy.Style{FlexDirection: "row"},
    child1,
    child2,
)
```

#### ScrollView

A container that allows scrolling its content if it exceeds available size.

```go
tizzy.NewScrollView(
    ctx,
    tizzy.Style{Height: 10},
    largeContentNode,
)
```

#### Modal

An overlay that traps focus and blocks interaction with the background. Controlled by `isOpen` boolean.

```go
tizzy.NewModal(
    ctx,
    tizzy.Style{},
    modalContentNode,
    isOpen,
)
```

### Basic Components

#### Text

Displays static text.

```go
tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Hello World")
```

#### Button

A clickable button.

```go
tizzy.NewButton(tizzy.Style{Focusable: true}, "Click Me", func() {
    // handle click
})
```

#### TextInput

A single-line text input field.

```go
tizzy.NewTextInput(ctx, tizzy.Style{Focusable: true}, "initial value", func(newValue string) {
    // handle change
})
```

### Selection Components

#### List

Displays a list of selectable items.

```go
tizzy.NewList(
    ctx,
    tizzy.Style{Focusable: true},
    "state-key", // Key to identify list state across renders
    items,       // []any
    func(item any, index int, selected bool, cursor bool) tizzy.Node {
        return tizzy.NewListItem(label, selected, cursor)
    },
    func(idx int) {
        // handle selection
    },
)
```

Options on the returned `*List` struct:

- `OnSelectionChange func(int)`: Called when the cursor moves.
- `OnFocus func(state *ListState)`: Called when the list gains focus.

#### Checkbox

A toggleable checkbox.

```go
tizzy.NewCheckbox(ctx, tizzy.Style{Focusable: true}, "Enable Feature", true, func(checked bool) {
    // handle change
})
```

#### RadioButton

A mutually exclusive selection button.

```go
tizzy.NewRadioButton(ctx, tizzy.Style{Focusable: true}, "Option 1", "value1", isSelected, func(value string) {
    // handle selection
})
```

#### Dropdown

A dropdown menu for selecting from a list.

```go
tizzy.NewDropdown(
    ctx,
    tizzy.Style{Focusable: true},
    []string{"Option A", "Option B", "Option C"},
    selectedIndex,
    func(idx int) {
        // handle selection
    },
)
```

### Navigation

#### Tabs

A tabbed interface for switching between views.

```go
tizzy.NewTabs(
    ctx,
    tizzy.Style{Focusable: true},
    []tizzy.Tab{
        {Title: "Home", Content: homeNode},
        {Title: "Settings", Content: settingsNode},
    },
)
```

#### MenuBar

A top-level specific menu bar with support for Alt shortcuts and dropdowns. Uses internal hooks for state.

```go
tizzy.NewMenuBar(
    ctx,
    tizzy.Style{Focusable: true},
    []tizzy.Menu{
        {
            Title: "File",
            Items: []tizzy.MenuItem{
                {Title: "New", Action: func() {}},
                {Title: "Exit", Action: func() {}},
            },
        },
    },
)
```

### Feedback & Data

#### ProgressBar

A horizontal progress bar.

```go
tizzy.NewProgressBar(tizzy.Style{}, 0.75) // 75%
```

#### Spinner

An animated loading spinner.

```go
tizzy.NewSpinner(ctx, tizzy.Style{Color: tcell.ColorCyan})
```

#### Table

A grid table for displaying tabular data.

```go
tizzy.NewTable(
    tizzy.Style{},
    []string{"Name", "Age", "Role"},
    [][]string{
        {"Alice", "30", "Dev"},
        {"Bob", "25", "Designer"},
    },
)
```

## Installation

```bash
go get github.com/pavelgj/tizzy-go
```
