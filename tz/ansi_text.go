package tz

import (
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
)

// ANSIText is a leaf node that renders a string containing ANSI SGR escape
// sequences (e.g. the output of lipgloss.Render or glamour.Render) inside the
// Tizzy layout tree. Visible width is measured by stripping escape codes;
// newlines produce additional rows.
type ANSIText struct {
	Style   Style
	Content string
	parsed  [][]styledRune
}

// NewANSIText creates an ANSIText node from a string that may contain ANSI
// escape sequences. The string is parsed once at construction time.
func NewANSIText(style Style, content string) *ANSIText {
	return &ANSIText{
		Style:   style,
		Content: content,
		parsed:  parseANSIString(content),
	}
}

// GetStyle implements Node.
func (n *ANSIText) GetStyle() Style { return n.Style }

// Layout implements Layoutable. Width is the widest row of visible characters;
// height is the number of lines. An explicit Style.Width overrides the computed
// width.
func (n *ANSIText) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin

	maxW := 0
	for _, row := range n.parsed {
		if len(row) > maxW {
			maxW = len(row)
		}
	}
	h := len(n.parsed)
	if h == 0 {
		h = 1
	}

	w := maxW + pad.Left + pad.Right
	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	return LayoutResult{
		Node: n,
		X:    x + margin.Left,
		Y:    y + margin.Top,
		W:    w,
		H:    h + pad.Top + pad.Bottom,
	}
}

// Render implements Renderable. Each parsed (rune, style) pair is written to
// the grid, clipped to the layout bounds.
func (n *ANSIText) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	baseX := layout.X + n.Style.Padding.Left
	baseY := layout.Y + n.Style.Padding.Top
	maxW := layout.W - n.Style.Padding.Left - n.Style.Padding.Right
	maxH := layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom

	for rowIdx, row := range n.parsed {
		if rowIdx >= maxH {
			break
		}
		y := baseY + rowIdx
		for colIdx, sr := range row {
			if colIdx >= maxW {
				break
			}
			grid.SetContent(baseX+colIdx, y, sr.R, sr.Style)
		}
	}
}

// ---------------------------------------------------------------------------
// ANSI parser
// ---------------------------------------------------------------------------

// styledRune pairs a rune with the tcell.Style that was active when it was
// parsed from an ANSI-escaped string.
type styledRune struct {
	R     rune
	Style tcell.Style
}

// parseANSIString converts a string with ANSI SGR escape sequences into a 2-D
// slice of styled runes (rows × columns). Each '\n' starts a new row.
// Non-SGR escape sequences (cursor movement, OSC hyperlinks, etc.) are skipped.
func parseANSIString(s string) [][]styledRune {
	var rows [][]styledRune
	var cur []styledRune
	style := tcell.StyleDefault
	runes := []rune(s)
	n := len(runes)
	i := 0

	for i < n {
		r := runes[i]

		switch {
		case r == '\n':
			rows = append(rows, cur)
			cur = nil
			i++

		case r == '\r':
			// Swallow bare CR (common in terminal output).
			i++

		case r == '\x1b' && i+1 < n && runes[i+1] == '[':
			// CSI sequence: ESC [ <params> <cmd>
			i += 2
			j := i
			for j < n && (runes[j] < '@' || runes[j] > '~') {
				j++
			}
			if j < n {
				cmd := runes[j]
				params := string(runes[i:j])
				i = j + 1
				if cmd == 'm' {
					style = applySGR(style, params)
				}
				// Other CSI commands (cursor movement, erase, etc.) are ignored.
			} else {
				i = j
			}

		case r == '\x1b' && i+1 < n && runes[i+1] == ']':
			// OSC sequence: ESC ] <payload> BEL  or  ESC ] <payload> ST (ESC \)
			i += 2
			for i < n {
				if runes[i] == '\x07' {
					i++
					break
				}
				if runes[i] == '\x1b' && i+1 < n && runes[i+1] == '\\' {
					i += 2
					break
				}
				i++
			}

		case r == '\x1b':
			// Other 2-byte escape sequences — skip ESC + next byte.
			i += 2

		default:
			cur = append(cur, styledRune{R: r, Style: style})
			i++
		}
	}

	// Flush the last line (even if it had no trailing newline).
	if len(cur) > 0 || len(rows) == 0 {
		rows = append(rows, cur)
	}

	return rows
}

// applySGR updates style by processing an SGR parameter string such as "1;38;2;255;0;0".
func applySGR(style tcell.Style, params string) tcell.Style {
	if params == "" {
		return tcell.StyleDefault
	}
	parts := strings.Split(params, ";")
	i := 0
	for i < len(parts) {
		code, err := strconv.Atoi(parts[i])
		if err != nil {
			i++
			continue
		}
		switch {
		case code == 0:
			style = tcell.StyleDefault
		case code == 1:
			style = style.Bold(true)
		case code == 2:
			style = style.Dim(true)
		case code == 3:
			style = style.Italic(true)
		case code == 4:
			style = style.Underline(true)
		case code == 5:
			style = style.Blink(true)
		case code == 7:
			style = style.Reverse(true)
		case code == 9:
			style = style.StrikeThrough(true)
		case code == 22:
			style = style.Bold(false).Dim(false)
		case code == 23:
			style = style.Italic(false)
		case code == 24:
			style = style.Underline(false)
		case code == 25:
			style = style.Blink(false)
		case code == 27:
			style = style.Reverse(false)
		case code == 29:
			style = style.StrikeThrough(false)
		case code >= 30 && code <= 37:
			style = style.Foreground(ansiBasicColor(code-30, false))
		case code == 38:
			c, skip := parseExtColor(parts, i+1)
			style = style.Foreground(c)
			i += skip
		case code == 39:
			style = style.Foreground(tcell.ColorDefault)
		case code >= 40 && code <= 47:
			style = style.Background(ansiBasicColor(code-40, false))
		case code == 48:
			c, skip := parseExtColor(parts, i+1)
			style = style.Background(c)
			i += skip
		case code == 49:
			style = style.Background(tcell.ColorDefault)
		case code >= 90 && code <= 97:
			style = style.Foreground(ansiBasicColor(code-90, true))
		case code >= 100 && code <= 107:
			style = style.Background(ansiBasicColor(code-100, true))
		}
		i++
	}
	return style
}

// parseExtColor reads a 256-color (5;n) or true-color (2;r;g;b) spec starting
// at parts[idx]. Returns the resolved color and the number of extra parts
// consumed beyond idx (so the caller can advance its index by that amount).
func parseExtColor(parts []string, idx int) (tcell.Color, int) {
	if idx >= len(parts) {
		return tcell.ColorDefault, 0
	}
	mode, err := strconv.Atoi(parts[idx])
	if err != nil {
		return tcell.ColorDefault, 0
	}
	switch mode {
	case 5: // 256-color: 38;5;n
		if idx+1 >= len(parts) {
			return tcell.ColorDefault, 1
		}
		n, err := strconv.Atoi(parts[idx+1])
		if err != nil {
			return tcell.ColorDefault, 1
		}
		return tcell.PaletteColor(n), 2
	case 2: // true-color: 38;2;r;g;b
		if idx+3 >= len(parts) {
			return tcell.ColorDefault, 1
		}
		rv, e1 := strconv.Atoi(parts[idx+1])
		gv, e2 := strconv.Atoi(parts[idx+2])
		bv, e3 := strconv.Atoi(parts[idx+3])
		if e1 != nil || e2 != nil || e3 != nil {
			return tcell.ColorDefault, 1
		}
		return tcell.NewRGBColor(int32(rv), int32(gv), int32(bv)), 4
	}
	return tcell.ColorDefault, 1
}

// ansiBasicColor maps a 0–7 ANSI color offset to a tcell.Color.
// bright=true gives the high-intensity (90–97 / 100–107) variants.
func ansiBasicColor(n int, bright bool) tcell.Color {
	normal := [8]tcell.Color{
		tcell.ColorBlack, tcell.ColorMaroon, tcell.ColorGreen, tcell.ColorOlive,
		tcell.ColorNavy, tcell.ColorPurple, tcell.ColorTeal, tcell.ColorSilver,
	}
	hi := [8]tcell.Color{
		tcell.ColorGray, tcell.ColorRed, tcell.ColorLime, tcell.ColorYellow,
		tcell.ColorBlue, tcell.ColorFuchsia, tcell.ColorAqua, tcell.ColorWhite,
	}
	if n < 0 || n > 7 {
		return tcell.ColorDefault
	}
	if bright {
		return hi[n]
	}
	return normal[n]
}
