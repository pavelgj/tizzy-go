package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

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

	renderToScreen(s, layout, "input1", states)
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
	renderToScreen(s, layout, "", nil)
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
