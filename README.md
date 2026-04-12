# Tizzy

A declarative Terminal User Interface (TUI) library for Go, inspired by React's component model and CSS Flexbox layout.

![Hero Image](docs/images/hero.png)

## Features

- **Declarative UI**: Build UIs by composing components that return a virtual tree of nodes.
- **Flexbox Layout**: Easy alignment and distribution of space with `row` and `column` directions, `JustifyContent`, and `FillWidth`/`FillHeight`.
- **Isolated State**: Support for struct-based components with isolated mutable state.
- **Rich Interactions**: Focus management, keyboard navigation, and mouse support.
- **Extensible**: Easy to create custom components.

## How it Compares to Alternatives

Tizzy takes a different approach compared to other popular Go TUI libraries. While [Bubbletea](https://github.com/charmbracelet/bubbletea) relies on the Elm Architecture (Model-Update-View) and [tview](https://github.com/rivo/tview) uses a more traditional widget hierarchy, Tizzy brings a **declarative, React-like component model** to Go TUIs. It features React-inspired hooks (like `UseState` and `UseEffect`) for local state management and a layout system heavily inspired by CSS Flexbox. This makes it highly intuitive for developers coming from modern web frameworks.

## Screenshots

### Kitchen Sink
Showing various form controls and state management.

![Kitchen Sink Forms](docs/images/kitchensink-forms.png)
![Kitchen Sink Buttons](docs/images/kitchensink-buttons-state.png)

### Overlays & Dialogs
Showing modals and dropdowns.

![Modal Dialog](docs/images/modal-dialog.png)
![Dropdown](docs/images/dropdown.png)

## Quick Start

Here is a minimal example of a counter application:

```go
package main

import (
	"log"
	"strconv"
	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		count, setCount := tz.UseState(ctx, 0)

		return tz.NewBox(
			tz.Style{
				Border:        true,
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				FlexDirection: "column",
			},
			tz.NewText(tz.Style{}, "Clicks: "+strconv.Itoa(count)),
			tz.NewButton(tz.Style{Focusable: true, ID: "btn-increment"}, "Increment", func() {
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

Every component implements the `Node` interface (`GetStyle() Style`) and optionally implements additional interfaces for layout, rendering, event handling, focus, and overlays. The framework detects capabilities at runtime — a component only needs to implement what it actually uses.

Tizzy supports two authoring patterns:

1.  **Functional Components**: Pure functions that take props and return a `Node`. Best for stateless or read-only UI.
2.  **Stateful Components**: Constructor functions that call `UseState`/`UseEffect` hooks and return a `Node`. Best for interactive components that own internal state.

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
app.Run(func(ctx *tz.RenderContext) tz.Node {
    count, setCount := tz.UseState(ctx, 0) // T is inferred as int

    return tz.NewButton(tz.Style{}, "Clicks: "+strconv.Itoa(count), func() {
        setCount(count + 1)
    })
}, func(ev tcell.Event) {})
```

### UseEffect

Allows performing side effects (like starting background tasks) when a component mounts, and cleaning up when it unmounts.

```go
app.Run(func(ctx *tz.RenderContext) tz.Node {
    ctx.UseEffect(func() func() {
        // OnInit / OnMount
        // Start a background goroutine or timer...

        return func() {
            // OnUnmount
            // Stop the goroutine or cleanup resources...
        }
    })

    return tz.NewText(tz.Style{}, "I keep track of my lifecycle")
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
tz.NewBox(
    tz.Style{FlexDirection: "row"},
    child1,
    child2,
)
```

#### ScrollView

A container that allows scrolling its content if it exceeds available size.

```go
tz.NewScrollView(
    ctx,
    tz.Style{Height: 10},
    largeContentNode,
)
```

#### Modal

A centered overlay that traps focus. Returns `nil` when closed (safe to include in `NewBox` children unconditionally).

```go
tz.NewModal(
    ctx,
    tz.Style{Background: tcell.ColorBlue},
    modalContentNode,
    isOpen,
)
```

#### Popup

A floating overlay anchored to an explicit screen position. Returns `nil` when closed.

```go
tz.NewPopup(
    ctx,
    tz.Style{Border: true, Background: tcell.ColorGray},
    popupContentNode,
    x, y,   // absolute screen position
    isOpen,
)

### Basic Components

#### Text

Displays static text.

```go
tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Hello World")
```

#### Button

A clickable button.

```go
tz.NewButton(tz.Style{Focusable: true, ID: "my-button"}, "Click Me", func() {
    // handle click
})
```

#### TextInput

A single-line text input field.

```go
tz.NewTextInput(ctx, tz.Style{Focusable: true}, "initial value", func(newValue string) {
    // handle change
})
```

### Selection Components

#### List

Displays a list of selectable items.

```go
tz.NewList(
    ctx,
    tz.Style{Focusable: true},
    "state-key",          // key: resets cursor/scroll when it changes
    items,                // []any
    0,                    // initial selected index (-1 for none)
    func(item any, index int, selected bool, cursor bool) tz.Node {
        return tz.NewListItem(item.(string), selected, cursor)
    },
    func(idx int) {
        // handle selection (Enter key or click)
    },
)
```

Options on the returned `*List` struct:

- `OnSelectionChange func(int)`: Called when the cursor moves.
- `OnFocus func(state *ListState)`: Called when the list gains focus.

#### Checkbox

A toggleable checkbox.

```go
tz.NewCheckbox(ctx, tz.Style{Focusable: true}, "Enable Feature", true, func(checked bool) {
    // handle change
})
```

#### RadioButton

A mutually exclusive selection button.

```go
tz.NewRadioButton(ctx, tz.Style{Focusable: true}, "Option 1", "value1", isSelected, func(value string) {
    // handle selection
})
```

#### Dropdown

A dropdown menu for selecting from a list.

```go
tz.NewDropdown(
    ctx,
    tz.Style{Focusable: true},
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
tz.NewTabs(
    ctx,
    tz.Style{Focusable: true},
    []tz.Tab{
        {Title: "Home", Content: homeNode},
        {Title: "Settings", Content: settingsNode},
    },
)
```

#### MenuBar

A top-level menu bar with keyboard navigation and Alt-key shortcuts.

```go
tz.NewMenuBar(
    ctx,
    tz.Style{Focusable: true},
    []tz.Menu{
        {
            Title:   "File",
            AltRune: 'f',
            Items: []tz.MenuItem{
                {Label: "New", Action: func() {}},
                {Label: "Exit", Action: func() {}},
            },
        },
    },
)
```

### Feedback & Data

#### ProgressBar

A horizontal progress bar.

```go
tz.NewProgressBar(tz.Style{}, 0.75) // 75%
```

#### Spinner

An animated loading spinner.

```go
tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorCyan})
```

#### Table

A grid table for displaying tabular data.

```go
tz.NewTable(
    tz.Style{},
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

## Further Reading

- [Architecture & Internals](docs/architecture.md) — render pipeline, Portal mechanism, overlay model, interface catalogue
- [Creating New Components](docs/new-component.md) — step-by-step guide with layout, render, event, and overlay examples
- [Contributing](CONTRIBUTING.md) — visual regression testing workflow
