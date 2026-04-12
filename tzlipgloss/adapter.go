// Package tzlipgloss bridges the Charmbracelet Lipgloss styling library with
// Tizzy. Import this package only in applications that already use Lipgloss;
// the core tz package has no dependency on Lipgloss.
//
// The two entry points are:
//
//   - [Color] converts any lipgloss.TerminalColor to a tcell.Color.
//   - [Style] converts the color and text-attribute properties of a
//     lipgloss.Style into a tz.Style.  Layout properties (padding, border,
//     width, height, margin) are intentionally not mapped — express those via
//     tz.Style directly, as Tizzy owns the layout pass.
package tzlipgloss

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
	tz "github.com/pavelgj/tizzy-go/tz"
)

// Color converts a lipgloss.TerminalColor to a tcell.Color.
//
// Conversion rules:
//   - [lipgloss.NoColor] → tcell.ColorDefault
//   - [lipgloss.Color] — hex string ("#RRGGBB", "#RGB") → true-color;
//     numeric string ("196") → 256-color palette index
//   - [lipgloss.ANSIColor] → 256-color palette index
//   - [lipgloss.AdaptiveColor] → resolves the Dark variant (most terminals
//     use dark backgrounds; switch to Light manually if needed)
//   - [lipgloss.CompleteColor] → prefers TrueColor, falls back to ANSI256
//   - [lipgloss.CompleteAdaptiveColor] → resolves Dark.TrueColor / ANSI256
//
// Color strings are parsed directly without going through the Lipgloss/termenv
// renderer, so the result is stable in headless and test environments.
func Color(c lipgloss.TerminalColor) tcell.Color {
	if c == nil {
		return tcell.ColorDefault
	}
	switch v := c.(type) {
	case lipgloss.NoColor:
		return tcell.ColorDefault
	case lipgloss.Color:
		return parseColorString(string(v))
	case lipgloss.ANSIColor:
		return tcell.PaletteColor(int(v))
	case lipgloss.AdaptiveColor:
		// Default to the dark variant; callers on light terminals can use
		// parseColorString(v.Light) directly.
		return parseColorString(v.Dark)
	case lipgloss.CompleteColor:
		if c := parseColorString(v.TrueColor); c != tcell.ColorDefault {
			return c
		}
		return parseColorString(v.ANSI256)
	case lipgloss.CompleteAdaptiveColor:
		if c := parseColorString(v.Dark.TrueColor); c != tcell.ColorDefault {
			return c
		}
		return parseColorString(v.Dark.ANSI256)
	}
	return tcell.ColorDefault
}

// parseColorString converts a Lipgloss color string to a tcell.Color.
// Supports:
//   - "#RRGGBB" and "#RGB" hex strings → true-color
//   - "0"–"255" decimal strings → 256-color palette
//   - "" or unrecognised input → tcell.ColorDefault
func parseColorString(s string) tcell.Color {
	s = strings.TrimSpace(s)
	if s == "" {
		return tcell.ColorDefault
	}
	if strings.HasPrefix(s, "#") {
		hex := s[1:]
		if len(hex) == 3 {
			// Expand shorthand: #RGB → #RRGGBB
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if len(hex) == 6 {
			r, e1 := strconv.ParseUint(hex[0:2], 16, 8)
			g, e2 := strconv.ParseUint(hex[2:4], 16, 8)
			b, e3 := strconv.ParseUint(hex[4:6], 16, 8)
			if e1 == nil && e2 == nil && e3 == nil {
				return tcell.NewRGBColor(int32(r), int32(g), int32(b))
			}
		}
		return tcell.ColorDefault
	}
	// ANSI 256-color index
	n, err := strconv.Atoi(s)
	if err == nil && n >= 0 && n <= 255 {
		return tcell.PaletteColor(n)
	}
	return tcell.ColorDefault
}

// Style converts the color and text-attribute properties of a lipgloss.Style
// into a tz.Style. The returned tz.Style can be further modified by the caller
// before passing it to any Tizzy constructor.
//
// Mapped properties:
//   - Foreground color → tz.Style.Color
//   - Background color → tz.Style.Background
//   - Bold  → preserved in the tcell layer via the Render path (see note below)
//   - Italic, Underline, Strikethrough, Blink, Faint, Reverse
//
// Not mapped (Tizzy owns these):
//   - Padding, Margin, Border, Width, Height, Align
//
// Note on text attributes: tz.Style does not have dedicated bool fields for
// bold/italic/etc. — those are expressed via tcell.Style at render time.
// Because tz.Style is a plain struct passed to constructors, the text
// attributes from Lipgloss are encoded into the tz.Style.TextAttrs field
// (a tcell.AttrMask). Tizzy's built-in Text and ANSIText renderers apply this
// mask when constructing the tcell.Style for each cell.
func Style(ls lipgloss.Style) tz.Style {
	return tz.Style{
		Color:      Color(ls.GetForeground()),
		Background: Color(ls.GetBackground()),
		TextAttrs:  textAttrs(ls),
	}
}

// textAttrs builds a tcell.AttrMask from the text-attribute getters on a
// lipgloss.Style.
func textAttrs(ls lipgloss.Style) tcell.AttrMask {
	var mask tcell.AttrMask
	if ls.GetBold() {
		mask |= tcell.AttrBold
	}
	if ls.GetItalic() {
		mask |= tcell.AttrItalic
	}
	if ls.GetUnderline() {
		mask |= tcell.AttrUnderline
	}
	if ls.GetStrikethrough() {
		mask |= tcell.AttrStrikeThrough
	}
	if ls.GetBlink() {
		mask |= tcell.AttrBlink
	}
	if ls.GetFaint() {
		mask |= tcell.AttrDim
	}
	if ls.GetReverse() {
		mask |= tcell.AttrReverse
	}
	return mask
}
