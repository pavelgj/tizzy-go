package tizzy

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestRenderTextInputScrolling(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Width: 5, ID: "input1"}, "1234567890", nil)
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
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("At col %d, expected '%c', got '%s'", i, c, str)
		}
	}
}

func TestRenderTextInputMultiline(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := makeTestContext()
	input := NewTextInput(ctx, Style{Multiline: true}, "abc\ndef", nil)
	layout := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	// Check line 0 "abc"
	expectedLine0 := "abc"
	for i, c := range expectedLine0 {
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("At line 0, col %d, expected '%c', got '%s'", i, c, str)
		}
	}

	// Check line 1 "def"
	expectedLine1 := "def"
	for i, c := range expectedLine1 {
		str, _, _ := s.Get(i, 1)
		if str != string(c) {
			t.Errorf("At line 1, col %d, expected '%c', got '%s'", i, c, str)
		}
	}
}

func TestTextInputHandleEvent(t *testing.T) {
	input := &TextInput{
		Value: "hello",
		Style: Style{ID: "myinput"},
	}
	state := &TextInputState{cursorOffset: 5}

	// Simulate KeyLeft
	ev := tcell.NewEventKey(tcell.KeyLeft, ' ', tcell.ModNone)
	ctx := EventContext{Layout: LayoutResult{H: 1}}

	handled := input.HandleEvent(ev, state, ctx)
	if !handled {
		t.Errorf("Expected event to be handled")
	}
	if state.cursorOffset != 4 {
		t.Errorf("Expected cursorOffset 4, got %d", state.cursorOffset)
	}

	// Simulate typing a rune
	ev = tcell.NewEventKey(tcell.KeyRune, '!', tcell.ModNone)
	handled = input.HandleEvent(ev, state, ctx)
	if !handled {
		t.Errorf("Expected event to be handled")
	}
	if input.Value != "hell!o" {
		t.Errorf("Expected Value 'hell!o', got '%s'", input.Value)
	}
	if state.cursorOffset != 5 {
		t.Errorf("Expected cursorOffset 5, got %d", state.cursorOffset)
	}
}
