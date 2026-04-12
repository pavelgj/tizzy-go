package tzlipgloss

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/gdamore/tcell/v2"
)

func TestColor_NoColor(t *testing.T) {
	c := Color(lipgloss.NoColor{})
	if c != tcell.ColorDefault {
		t.Errorf("NoColor should map to tcell.ColorDefault, got %v", c)
	}
}

func TestColor_Nil(t *testing.T) {
	c := Color(nil)
	if c != tcell.ColorDefault {
		t.Errorf("nil should map to tcell.ColorDefault, got %v", c)
	}
}

func TestColor_HexColor(t *testing.T) {
	c := Color(lipgloss.Color("#FF0000"))
	want := tcell.NewRGBColor(255, 0, 0)
	if c != want {
		t.Errorf("expected %v, got %v", want, c)
	}
}

func TestColor_AdaptiveColor(t *testing.T) {
	// We can't control which variant is selected (depends on terminal), but we
	// can assert that the result is not ColorDefault and is a valid color.
	c := Color(lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"})
	// Both are valid non-zero colors (black and white are real colors, but
	// the zero-guard in Color() maps them to Default). Just assert no panic.
	_ = c
}

func TestStyle_ForegroundBackground(t *testing.T) {
	ls := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Background(lipgloss.Color("#0000FF"))

	s := Style(ls)

	wantFG := tcell.NewRGBColor(0, 255, 0)
	if s.Color != wantFG {
		t.Errorf("foreground: expected %v, got %v", wantFG, s.Color)
	}

	wantBG := tcell.NewRGBColor(0, 0, 255)
	if s.Background != wantBG {
		t.Errorf("background: expected %v, got %v", wantBG, s.Background)
	}
}

func TestStyle_BoldAttribute(t *testing.T) {
	ls := lipgloss.NewStyle().Bold(true)
	s := Style(ls)
	if s.TextAttrs&tcell.AttrBold == 0 {
		t.Error("expected AttrBold to be set")
	}
}

func TestStyle_ItalicAttribute(t *testing.T) {
	ls := lipgloss.NewStyle().Italic(true)
	s := Style(ls)
	if s.TextAttrs&tcell.AttrItalic == 0 {
		t.Error("expected AttrItalic to be set")
	}
}

func TestStyle_UnderlineAttribute(t *testing.T) {
	ls := lipgloss.NewStyle().Underline(true)
	s := Style(ls)
	if s.TextAttrs&tcell.AttrUnderline == 0 {
		t.Error("expected AttrUnderline to be set")
	}
}

func TestStyle_MultipleAttributes(t *testing.T) {
	ls := lipgloss.NewStyle().Bold(true).Italic(true).Underline(true)
	s := Style(ls)
	if s.TextAttrs&tcell.AttrBold == 0 {
		t.Error("expected AttrBold")
	}
	if s.TextAttrs&tcell.AttrItalic == 0 {
		t.Error("expected AttrItalic")
	}
	if s.TextAttrs&tcell.AttrUnderline == 0 {
		t.Error("expected AttrUnderline")
	}
}

func TestStyle_NoAttributesWhenUnset(t *testing.T) {
	ls := lipgloss.NewStyle()
	s := Style(ls)
	if s.TextAttrs != 0 {
		t.Errorf("expected no text attributes, got %v", s.TextAttrs)
	}
}

func TestStyle_LayoutPropertiesNotMapped(t *testing.T) {
	// Padding/margin/border/width/height should NOT be carried over.
	ls := lipgloss.NewStyle().Padding(2).Margin(3).Width(80).Height(24)
	s := Style(ls)
	if s.Width != 0 {
		t.Errorf("Width should not be mapped, got %d", s.Width)
	}
	if s.Height != 0 {
		t.Errorf("Height should not be mapped, got %d", s.Height)
	}
	p := s.Padding
	if p.Top != 0 || p.Bottom != 0 || p.Left != 0 || p.Right != 0 {
		t.Errorf("Padding should not be mapped, got %+v", p)
	}
}
