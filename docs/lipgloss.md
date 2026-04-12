# Lipgloss Interoperability Proposal

## Motivation

The [Charmbracelet](https://github.com/charmbracelet) ecosystem (Lipgloss, Glamour, Huh) is widely used in the Go TUI community. A common objection to adopting Tizzy is that it is standalone: a developer who already uses Lipgloss for styling or Glamour for markdown rendering cannot carry that investment into Tizzy. They must abandon familiar tools and re-learn Tizzy-specific equivalents.

This proposal describes two targeted additions that close the gap without changing Tizzy's architecture:

1. **`ANSIText` node** — renders a string containing ANSI escape sequences (e.g. the output of `lipgloss.Render` or `glamour.Render`) directly inside the Tizzy layout tree.
2. **`LipglossStyle` adapter** — converts a `lipgloss.Style` to a `tz.Style` so that existing design tokens can be reused across both libraries.

Neither addition requires Tizzy to take a dependency on Lipgloss or Glamour. Both are opt-in.

---

## Feature 1: `ANSIText` Node

### The problem

Tizzy's `Grid` stores `(rune, tcell.Style)` pairs. It has no concept of ANSI escape sequences. Passing the output of `lipgloss.Render("hello")` to `tz.NewText` would display raw escape codes as garbage characters.

### Proposed API

```go
// NewANSIText creates a node that renders a string containing ANSI escape
// sequences. The string may contain newlines; each line becomes a row in the
// layout. Visible width is calculated by stripping escape codes.
func NewANSIText(style tz.Style, ansiString string) tz.Node
```

### Usage examples

**Lipgloss-styled label inside a Tizzy layout:**

```go
import (
    "github.com/charmbracelet/lipgloss"
    "github.com/pavelgj/tizzy-go/tz"
)

boldRed := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF5555"))

func render(ctx *tz.RenderContext) tz.Node {
    return tz.NewBox(
        tz.Style{FlexDirection: "column", Border: true},
        tz.NewANSIText(tz.Style{}, boldRed.Render("Error")),
        tz.NewText(tz.Style{}, "Something went wrong."),
    )
}
```

**Glamour markdown in a scrollable pane:**

```go
import (
    "github.com/charmbracelet/glamour"
    "github.com/pavelgj/tizzy-go/tz"
)

func render(ctx *tz.RenderContext) tz.Node {
    // Glamour needs a fixed render width; pass the pane width or a sensible default.
    rendered, _ := glamour.Render(markdownContent, "dark")

    return tz.NewScrollView(ctx,
        tz.Style{FillWidth: true, FillHeight: true},
        tz.NewANSIText(tz.Style{}, rendered),
    )
}
```

### Implementation notes

**Layout**: Strip all ANSI escape sequences (CSI `\x1b[...m` sequences) using a small parser, split on `\n`, and measure visible rune width per line. Report `W = max(line widths)` and `H = number of lines`. Wide (East Asian / emoji) runes should count as 2 columns — use `unicode/utf8` + a width table or `golang.org/x/text/width`.

**Render**: Walk the string byte by byte, maintaining a current `tcell.Style` accumulator. On each CSI color/attribute sequence, update the accumulator:

| ANSI code | tcell mapping |
|---|---|
| `38;2;r;g;b` (true-color FG) | `tcell.NewRGBColor(r, g, b)` |
| `48;2;r;g;b` (true-color BG) | `tcell.NewRGBColor(r, g, b)` |
| `38;5;n` (256-color FG) | `tcell.PaletteColor(n)` |
| `48;5;n` (256-color BG) | `tcell.PaletteColor(n)` |
| `30`–`37`, `90`–`97` (basic FG) | `tcell.ColorBlack` … etc. |
| `40`–`47`, `100`–`107` (basic BG) | mapped similarly |
| `1` (bold) | `.Bold(true)` |
| `3` (italic) | `.Italic(true)` |
| `4` (underline) | `.Underline(true)` |
| `0` / `m` (reset) | `tcell.StyleDefault` |

Each visible rune is written to the grid with the accumulated style. The base `tz.Style.Color` and `tz.Style.Background` act as defaults before any ANSI sequence is seen in a given line.

**No external dependency**: The ANSI parser is ~100 lines of pure Go with no imports beyond `strings`, `strconv`, and `unicode/utf8`. Tizzy does not need to import Lipgloss or Glamour.

### Known limitations

- Lipgloss border/padding rendering (where Lipgloss draws the box itself) produces multi-line ANSI strings. `ANSIText` will render the visual output correctly, but Tizzy has no knowledge of the internal structure — focus, hit-testing, and dynamic resizing will not work on such content. For interactive components, use Tizzy's own layout primitives and only use `ANSIText` for the *content* inside them.
- Glamour renders to a fixed width. If the terminal is wider or narrower than that width, the output will not reflow. The caller is responsible for passing the correct render width (typically obtained from the layout constraints in a custom `Layoutable` implementation).

---

## Feature 2: `LipglossStyle` Adapter

### The problem

Developers who maintain a Lipgloss-based design system (a shared `styles.go` with color palettes, typography scale, component variants) must duplicate that work if they want consistent styling in a Tizzy app. A conversion function lets them reference the same source of truth.

### Proposed API

```go
// LipglossStyle converts the color and text-attribute properties of a
// lipgloss.Style into a tz.Style. Layout properties (padding, border, width,
// height, margin) are intentionally not mapped — express those via tz.Style
// directly, as Tizzy owns the layout pass.
func LipglossStyle(ls lipgloss.Style) tz.Style

// LipglossColor converts a lipgloss.TerminalColor (lipgloss.Color,
// lipgloss.AdaptiveColor, lipgloss.CompleteColor) to a tcell.Color.
// Useful when you need just the color without a full style conversion.
func LipglossColor(c lipgloss.TerminalColor) tcell.Color
```

### Usage example

```go
// styles.go — shared design tokens
var (
    Primary   = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
    Muted     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
    Danger    = lipgloss.NewStyle().Foreground(lipgloss.Color("#DC2626"))
)

// tizzy component — reuse the same tokens
func render(ctx *tz.RenderContext) tz.Node {
    return tz.NewBox(
        tz.Style{FlexDirection: "column"},
        tz.NewText(tz.LipglossStyle(Primary).WithID("heading"), "Dashboard"),
        tz.NewText(tz.LipglossStyle(Muted), "Last updated: just now"),
    )
}
```

The hypothetical `.WithID("heading")` call illustrates that `LipglossStyle` returns a `tz.Style` that the caller can further modify — it is not a terminal operation.

### What is mapped

| Lipgloss property | tz.Style field |
|---|---|
| `GetForeground()` | `Color` |
| `GetBackground()` | `Background` |
| Bold, Italic, Underline | (stored in a new `TextAttrs` field on `tz.Style` or passed through to `tcell.Style` at render time — see below) |

### What is intentionally NOT mapped

| Lipgloss property | Reason |
|---|---|
| `GetPadding*()` | Tizzy owns layout; use `tz.Style.Padding` |
| `GetMargin*()` | Same — use `tz.Style.Margin` |
| `GetBorder*()` | Tizzy's border is a bool; Lipgloss's is a struct |
| `GetWidth()` / `GetHeight()` | Tizzy sizes are in terminal cells and managed by the layout pass |
| `GetAlign*()` | Use `tz.Style.JustifyContent` |

### Color conversion

`lipgloss.TerminalColor` is an interface satisfied by:
- `lipgloss.Color` — a string that is either an ANSI 256-color index (`"5"`), a hex color (`"#FF0066"`), or empty
- `lipgloss.AdaptiveColor` — light/dark variants; `LipglossColor` should detect the terminal's background and pick the appropriate variant (delegating to `lipgloss.HasDarkBackground()`)
- `lipgloss.CompleteColor` — explicit true-color + 256-color + ANSI fallbacks; prefer the true-color value

Hex colors convert to `tcell.NewRGBColor(r, g, b)`. ANSI index strings convert to `tcell.PaletteColor(n)`.

### Dependency note

Unlike `ANSIText`, `LipglossStyle` requires importing `github.com/charmbracelet/lipgloss`. To keep this optional, it should live in a sub-package:

```
tz/lipgloss/          — package tzlipgloss
    adapter.go        — LipglossStyle(), LipglossColor()
```

This way the core `tz` package has zero new dependencies, and only apps that actually use Lipgloss pay the import cost.

---

## Relationship between the two features

The two features are independent and complementary:

- Use **`ANSIText`** when you want to embed *rendered output* from Lipgloss or Glamour — useful for rich formatting, markdown, or content generated by third-party Charm-based packages.
- Use **`LipglossStyle`** when you want to *reuse design tokens* (colors, text attributes) from an existing Lipgloss style system without duplicating them in Tizzy constants.

A typical app might use both: `LipglossStyle` for interactive Tizzy components (buttons, inputs, labels) and `ANSIText` for non-interactive display content (help text, log output, markdown documentation).

---

## What this does NOT address

- **Huh**: Huh is a forms library with its own input, select, and confirm components. Bridging Huh and Tizzy would require either wrapping Huh components as Tizzy nodes (very deep integration) or replacing them with Tizzy equivalents. Tizzy already has first-class `TextInput`, `Checkbox`, `Dropdown`, and `RadioButton` components, so this is largely a migration concern rather than a gap.
- **Wish (SSH)**: Tizzy's `NewAppWithScreen` already accepts any `tcell.Screen`, so an SSH-multiplexed screen could be plugged in at the app level. This is an integration point, not a library change.

---

## Prioritization

| Item | Effort | Impact |
|---|---|---|
| `ANSIText` node | Medium (ANSI parser + layout) | High — unlocks Glamour and any Lipgloss-rendered content |
| `LipglossColor` helper | Low (color string parsing) | Medium — useful standalone utility |
| `LipglossStyle` adapter (sub-package) | Low–Medium | Medium — reduces duplication for teams with existing Lipgloss design systems |

Recommendation: implement `ANSIText` first. It addresses the most common concrete use case (embedding Glamour markdown) and requires no external dependency. The `LipglossStyle` adapter is a good follow-up once `ANSIText` is proven out.
