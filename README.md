# Splotch

A declarative Terminal User Interface (TUI) library for Go, inspired by React's component model and CSS Flexbox layout.

## Features

-   **Declarative UI**: Build UIs by composing components that return a virtual tree of nodes.
-   **Flexbox Layout**: Easy alignment and distribution of space with `row` and `column` directions, `JustifyContent`, and `FillWidth`/`FillHeight`.
-   **Isolated State**: Support for struct-based components with isolated mutable state.
-   **Rich Interactions**: Focus management, keyboard navigation, and mouse support.
-   **Extensible**: Easy to create custom components.

## Quick Start

Here is a minimal example of a counter application:

```go
package main

import (
	"log"
	"strconv"
	"splotch/splotch"
)

type Counter struct {
	count int
}

func (c *Counter) Render() splotch.Node {
	return splotch.NewBox(
		splotch.Style{
			Border:        true,
			Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			FlexDirection: "column",
			FillWidth:     true,
		},
		splotch.NewText(splotch.Style{}, "Clicks: "+strconv.Itoa(c.count)),
		splotch.NewButton(splotch.Style{Focusable: true}, "Increment", func() {
			c.count++
		}),
	)
}

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	counter := &Counter{}

	app.Run(func(ctx *splotch.RenderContext) splotch.Node {
		return counter.Render()
	})
}
```

## Core Concepts

### Components
Splotch supports two patterns for components:
1.  **Functional Components**: Pure functions that take props and return a `Node`. Best for stateless or read-only UIs.
2.  **Struct Components**: Structs that hold state and have a `Render() Node` method. Best for stateful, interactive parts of the UI.

### Layout
Layout is handled by the `Box` component using a Flexbox-style system.
-   `FlexDirection`: "row" or "column".
-   `FillWidth` / `FillHeight`: Fill available space.
-   `JustifyContent`: "flex-start", "center", "flex-end".

## Hooks and Lifecycle

Splotch supports React-like hooks for state management and lifecycle effects within the render function.

### UseState
Allows components to have local state that persists across renders.
```go
app.Run(func(ctx *splotch.RenderContext) splotch.Node {
    countObj, setCount := ctx.UseState(0)
    count := countObj.(int)
    
    return splotch.NewButton(splotch.Style{}, "Clicks: "+strconv.Itoa(count), func() {
        setCount(count + 1)
    })
})
```

### UseEffect
Allows performing side effects (like starting background tasks) when a component mounts, and cleaning up when it unmounts.
```go
app.Run(func(ctx *splotch.RenderContext) splotch.Node {
    ctx.UseEffect(func() func() {
        // OnInit / OnMount
        // Start a background goroutine or timer...
        
        return func() {
            // OnUnmount
            // Stop the goroutine or cleanup resources...
        }
    })
    
    return splotch.NewText(splotch.Style{}, "I keep track of my lifecycle")
})
```
> [!NOTE]
> `UseEffect` relies on call order to identify effects across renders. Do not call hooks inside loops or conditions.

## Components Reference

Here are all the available components in Splotch:

### Containers & Layout

#### Box
The fundamental layout container.
```go
splotch.NewBox(
    splotch.Style{FlexDirection: "row"},
    child1,
    child2,
)
```

#### ScrollView
A container that allows scrolling its content if it exceeds available size.
```go
splotch.NewScrollView(
    splotch.Style{Height: 10},
    largeContentNode,
)
```

#### Modal
An overlay that traps focus and blocks interaction with the background. Controlled by `isOpen` boolean.
```go
splotch.NewModal(
    ctx,
    splotch.Style{},
    modalContentNode,
    isOpen,
)
```

### Basic Components

#### Text
Displays static text.
```go
splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "Hello World")
```

#### Button
A clickable button.
```go
splotch.NewButton(splotch.Style{Focusable: true}, "Click Me", func() {
    // handle click
})
```

#### TextInput
A single-line text input field.
```go
splotch.NewTextInput(splotch.Style{Focusable: true}, "initial value", func(newValue string) {
    // handle change
})
```

### Selection Components

#### Checkbox
A toggleable checkbox.
```go
splotch.NewCheckbox(splotch.Style{Focusable: true}, "Enable Feature", true, func(checked bool) {
    // handle change
})
```

#### RadioButton
A mutually exclusive selection button.
```go
splotch.NewRadioButton(splotch.Style{Focusable: true}, "Option 1", "value1", isSelected, func(value string) {
    // handle selection
})
```

#### Dropdown
A dropdown menu for selecting from a list.
```go
splotch.NewDropdown(
    splotch.Style{Focusable: true},
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
splotch.NewTabs(
    splotch.Style{Focusable: true},
    []splotch.Tab{
        {Title: "Home", Content: homeNode},
        {Title: "Settings", Content: settingsNode},
    },
)
```

#### MenuBar
A top-level specific menu bar with support for Alt shortcuts and dropdowns. Uses internal hooks for state.
```go
splotch.NewMenuBar(
    ctx,
    splotch.Style{Focusable: true},
    []splotch.Menu{
        {
            Title: "File",
            Items: []splotch.MenuItem{
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
splotch.NewProgressBar(splotch.Style{}, 0.75) // 75%
```

#### Spinner
An animated loading spinner.
```go
splotch.NewSpinner(splotch.Style{Color: tcell.ColorCyan})
```

#### Table
A grid table for displaying tabular data.
```go
splotch.NewTable(
    splotch.Style{},
    []string{"Name", "Age", "Role"},
    [][]string{
        {"Alice", "30", "Dev"},
        {"Bob", "25", "Designer"},
    },
)
```

## Installation

```bash
go get splotch/splotch
```

*(Note: Replace with actual import path when available)*

## License
MIT
