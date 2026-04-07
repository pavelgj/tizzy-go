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

	Render(s, layout, "", nil)
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
	Render(s, layout, "", nil)
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

func TestRenderTextInputScrolling(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	input := NewTextInput(Style{Width: 5, ID: "input1"}, "1234567890", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 2)
	
	// Set state with scrollOffset = 5
	states := map[string]any{
		"input1": &TextInputState{cursorOffset: 5, scrollOffset: 5},
	}

	Render(s, layout, "input1", states)
	s.Show()

	// Should show "67890"
	expected := "67890"
	for i, c := range expected {
		mainc, _, _, _ := s.GetContent(i, 0)
		if mainc != c {
			t.Errorf("At col %d, expected '%c', got '%c'", i, c, mainc)
		}
	}
}

func TestRenderTextInputMultiline(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	input := NewTextInput(Style{Multiline: true}, "abc\ndef", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 5)
	Render(s, layout, "", nil)
	s.Show()

	// Check line 0 "abc"
	expectedLine0 := "abc"
	for i, c := range expectedLine0 {
		mainc, _, _, _ := s.GetContent(i, 0)
		if mainc != c {
			t.Errorf("At line 0, col %d, expected '%c', got '%c'", i, c, mainc)
		}
	}

	// Check line 1 "def"
	expectedLine1 := "def"
	for i, c := range expectedLine1 {
		mainc, _, _, _ := s.GetContent(i, 1)
		if mainc != c {
			t.Errorf("At line 1, col %d, expected '%c', got '%c'", i, c, mainc)
		}
	}
}

