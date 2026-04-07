package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRenderBorder(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	box := NewBox(Style{Border: true},
		NewText(Style{}, "Hi"),
	)

	// Layout calculates size: 2 chars for "Hi" + 2 for borders = 4 width.
	// Height: 1 line + 2 for borders = 3 height.
	layout := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	
	// Simulate screen size
	s.SetSize(10, 10)

	Render(s, layout, "")
	s.Show()

	// Check top-left corner
	mainc, _, _, _ := s.GetContent(0, 0)
	if mainc != '┌' {
		t.Errorf("Expected '┌' at 0,0, got '%c'", mainc)
	}

	// Check top border
	mainc, _, _, _ = s.GetContent(1, 0)
	if mainc != '─' {
		t.Errorf("Expected '─' at 1,0, got '%c'", mainc)
	}

	// Check text content position (should be at 1,1)
	mainc, _, _, _ = s.GetContent(1, 1)
	if mainc != 'H' {
		t.Errorf("Expected 'H' at 1,1, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(2, 1)
	if mainc != 'i' {
		t.Errorf("Expected 'i' at 2,1, got '%c'", mainc)
	}

	// Check bottom-right corner
	mainc, _, _, _ = s.GetContent(3, 2)
	if mainc != '┘' {
		t.Errorf("Expected '┘' at 3,2, got '%c'", mainc)
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
	Render(s, layout, "")
	s.Show()

	mainc, _, style, _ := s.GetContent(0, 0)
	if mainc != 'R' {
		t.Errorf("Expected 'R' at 0,0, got '%c'", mainc)
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

