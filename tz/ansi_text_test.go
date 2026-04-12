package tz

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// parseANSIString
// ---------------------------------------------------------------------------

func TestParseANSIString_PlainText(t *testing.T) {
	rows := parseANSIString("hello")
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0]) != 5 {
		t.Fatalf("expected 5 runes, got %d", len(rows[0]))
	}
	if rows[0][0].R != 'h' {
		t.Errorf("expected 'h', got %q", rows[0][0].R)
	}
}

func TestParseANSIString_Multiline(t *testing.T) {
	rows := parseANSIString("foo\nbar\nbaz")
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
	for i, want := range []string{"foo", "bar", "baz"} {
		got := string(runesFrom(rows[i]))
		if got != want {
			t.Errorf("row %d: expected %q, got %q", i, want, got)
		}
	}
}

func TestParseANSIString_BasicForegroundColor(t *testing.T) {
	// ESC[31m = red foreground
	rows := parseANSIString("\x1b[31mRed\x1b[0m")
	if len(rows) != 1 || len(rows[0]) != 3 {
		t.Fatalf("expected 1 row with 3 runes, got %d rows", len(rows))
	}
	fg, _, _ := rows[0][0].Style.Decompose()
	if fg != tcell.ColorMaroon {
		t.Errorf("expected ColorMaroon (ANSI red), got %v", fg)
	}
	// After reset the style should be default
	resetRows := parseANSIString("\x1b[31mRed\x1b[0mNormal")
	fg2, _, _ := resetRows[0][3].Style.Decompose()
	if fg2 != tcell.ColorDefault {
		t.Errorf("expected ColorDefault after reset, got %v", fg2)
	}
}

func TestParseANSIString_BrightForeground(t *testing.T) {
	// ESC[91m = bright red (high-intensity)
	rows := parseANSIString("\x1b[91mHi\x1b[0m")
	fg, _, _ := rows[0][0].Style.Decompose()
	if fg != tcell.ColorRed {
		t.Errorf("expected ColorRed (bright red), got %v", fg)
	}
}

func TestParseANSIString_256Color(t *testing.T) {
	// ESC[38;5;196m = 256-color index 196 (bright red)
	rows := parseANSIString("\x1b[38;5;196mX\x1b[0m")
	fg, _, _ := rows[0][0].Style.Decompose()
	if fg != tcell.PaletteColor(196) {
		t.Errorf("expected PaletteColor(196), got %v", fg)
	}
}

func TestParseANSIString_TrueColor(t *testing.T) {
	// ESC[38;2;255;100;0m = RGB(255,100,0)
	rows := parseANSIString("\x1b[38;2;255;100;0mX\x1b[0m")
	fg, _, _ := rows[0][0].Style.Decompose()
	want := tcell.NewRGBColor(255, 100, 0)
	if fg != want {
		t.Errorf("expected %v, got %v", want, fg)
	}
}

func TestParseANSIString_BackgroundColor(t *testing.T) {
	// ESC[42m = green background
	rows := parseANSIString("\x1b[42mX\x1b[0m")
	_, bg, _ := rows[0][0].Style.Decompose()
	if bg != tcell.ColorGreen {
		t.Errorf("expected ColorGreen background, got %v", bg)
	}
}

func TestParseANSIString_Bold(t *testing.T) {
	// ESC[1m = bold
	rows := parseANSIString("\x1b[1mB\x1b[0m")
	_, _, attrs := rows[0][0].Style.Decompose()
	if attrs&tcell.AttrBold == 0 {
		t.Errorf("expected bold attribute")
	}
}

func TestParseANSIString_MultipleParams(t *testing.T) {
	// ESC[1;31m = bold + red
	rows := parseANSIString("\x1b[1;31mX\x1b[0m")
	fg, _, attrs := rows[0][0].Style.Decompose()
	if fg != tcell.ColorMaroon {
		t.Errorf("expected ColorMaroon, got %v", fg)
	}
	if attrs&tcell.AttrBold == 0 {
		t.Errorf("expected bold attribute")
	}
}

func TestParseANSIString_OSCSequenceSkipped(t *testing.T) {
	// OSC hyperlink: ESC]8;;https://example.com\x07text ESC]8;;\x07
	input := "\x1b]8;;https://example.com\x07link\x1b]8;;\x07"
	rows := parseANSIString(input)
	got := string(runesFrom(rows[0]))
	if got != "link" {
		t.Errorf("expected OSC sequence skipped, got %q", got)
	}
}

func TestParseANSIString_EmptyString(t *testing.T) {
	rows := parseANSIString("")
	// Should return one empty row (not panic).
	if len(rows) != 1 {
		t.Fatalf("expected 1 row for empty string, got %d", len(rows))
	}
}

func TestParseANSIString_TrailingNewline(t *testing.T) {
	// A trailing newline (common in glamour output) does NOT produce a
	// spurious empty final row — we only flush a final row if there is
	// content after the last '\n'.
	rows := parseANSIString("line1\nline2\n")
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows (no empty trailing row), got %d", len(rows))
	}
}

// ---------------------------------------------------------------------------
// ANSIText layout
// ---------------------------------------------------------------------------

func TestANSITextLayout_SingleLine(t *testing.T) {
	node := NewANSIText(Style{}, "\x1b[31mHello\x1b[0m")
	res := Layout(node, 5, 10, Constraints{MaxW: 80, MaxH: 24})
	if res.X != 5 || res.Y != 10 {
		t.Errorf("expected (5,10), got (%d,%d)", res.X, res.Y)
	}
	if res.W != 5 {
		t.Errorf("expected W=5, got %d", res.W)
	}
	if res.H != 1 {
		t.Errorf("expected H=1, got %d", res.H)
	}
}

func TestANSITextLayout_Multiline(t *testing.T) {
	node := NewANSIText(Style{}, "Hello\nWorld!")
	res := Layout(node, 0, 0, Constraints{MaxW: 80, MaxH: 24})
	if res.W != 6 { // "World!" is 6 chars
		t.Errorf("expected W=6, got %d", res.W)
	}
	if res.H != 2 {
		t.Errorf("expected H=2, got %d", res.H)
	}
}

func TestANSITextLayout_ExplicitWidth(t *testing.T) {
	node := NewANSIText(Style{Width: 20}, "Hi")
	res := Layout(node, 0, 0, Constraints{MaxW: 80, MaxH: 24})
	if res.W != 20 {
		t.Errorf("expected explicit W=20, got %d", res.W)
	}
}

func TestANSITextLayout_Padding(t *testing.T) {
	node := NewANSIText(Style{Padding: Padding{Left: 2, Right: 2, Top: 1, Bottom: 1}}, "Hi")
	res := Layout(node, 0, 0, Constraints{MaxW: 80, MaxH: 24})
	if res.W != 6 { // 2 + 2 + 2
		t.Errorf("expected W=6, got %d", res.W)
	}
	if res.H != 3 { // 1 + 1 + 1
		t.Errorf("expected H=3, got %d", res.H)
	}
}

// ---------------------------------------------------------------------------
// ANSIText render
// ---------------------------------------------------------------------------

func TestANSITextRender_ColorsReachGrid(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 5)

	node := NewANSIText(Style{}, "\x1b[31mHi\x1b[0m")
	layout := Layout(node, 0, 0, Constraints{MaxW: 20, MaxH: 5})
	renderToScreen(s, layout, "", nil)
	s.Show()

	ch, style, _ := s.Get(0, 0)
	if ch != "H" {
		t.Errorf("expected 'H' at (0,0), got %q", ch)
	}
	fg, _, _ := style.Decompose()
	if fg != tcell.ColorMaroon {
		t.Errorf("expected ColorMaroon, got %v", fg)
	}
}

func TestANSITextRender_MultilinePositioning(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	s.SetSize(20, 5)

	node := NewANSIText(Style{}, "AB\nCD")
	layout := Layout(node, 0, 0, Constraints{MaxW: 20, MaxH: 5})
	renderToScreen(s, layout, "", nil)
	s.Show()

	ch0, _, _ := s.Get(0, 0)
	ch1, _, _ := s.Get(1, 0)
	ch2, _, _ := s.Get(0, 1)
	ch3, _, _ := s.Get(1, 1)

	if ch0 != "A" || ch1 != "B" || ch2 != "C" || ch3 != "D" {
		t.Errorf("expected AB/CD, got %q%q/%q%q", ch0, ch1, ch2, ch3)
	}
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func runesFrom(row []styledRune) []rune {
	out := make([]rune, len(row))
	for i, sr := range row {
		out[i] = sr.R
	}
	return out
}
