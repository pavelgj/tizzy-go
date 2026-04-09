package tz

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutText(t *testing.T) {
	text := NewText(Style{}, "Hello")
	res := Layout(text, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 5 || res.H != 1 {
		t.Errorf("Expected W=5, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderColors(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	text := NewText(Style{Color: tcell.ColorRed, Background: tcell.ColorBlue}, "Red on Blue")
	layout := Layout(text, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 2)
	renderToScreen(s, layout, "", nil)
	s.Show()

	str, style, _ := s.Get(0, 0)
	if str != "R" {
		t.Errorf("Expected 'R' at 0,0, got '%s'", str)
	}

	// Check style
	fg, bg, _ := style.Decompose()
	if fg != tcell.ColorRed {
		t.Errorf("Expected foreground ColorRed, got %v", fg)
	}
	if bg != tcell.ColorBlue {
		t.Errorf("Expected background ColorBlue, got %v", bg)
	}
}
