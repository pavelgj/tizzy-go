package tz

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutButton(t *testing.T) {
	btn := NewButton(Style{}, "Click", func() {})
	res := Layout(btn, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 9 || res.H != 1 {
		t.Errorf("Expected W=9, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderButton(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	btn := NewButton(Style{}, "Click", func() {})
	layout := Layout(btn, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	expected := "[ Click ]"
	for i, c := range expected {
		str, _, _ := s.Get(i, 0)
		if str != string(c) {
			t.Errorf("At col %d, expected '%c', got '%s'", i, c, str)
		}
	}
}

func TestRenderButtonFocusColors(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	btn := NewButton(Style{
		ID:              "btn",
		Focusable:       true,
		FocusColor:      tcell.ColorWhite,
		FocusBackground: tcell.ColorBlue,
	}, "Go", func() {})
	layout := Layout(btn, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "btn", nil)
	s.Show()

	// First cell of "[ Go ]" should carry the custom focus style.
	_, _, style, _ := s.GetContent(0, 0)
	fg, bg, _ := style.Decompose()
	if fg != tcell.ColorWhite {
		t.Errorf("Expected focused fg=White, got %v", fg)
	}
	if bg != tcell.ColorBlue {
		t.Errorf("Expected focused bg=Blue, got %v", bg)
	}
}

func TestRenderButtonDefaultFocusColors(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	// No FocusColor/FocusBackground set — should fall back to black/yellow.
	btn := NewButton(Style{ID: "btn", Focusable: true}, "Go", func() {})
	layout := Layout(btn, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "btn", nil)
	s.Show()

	_, _, style, _ := s.GetContent(0, 0)
	fg, bg, _ := style.Decompose()
	if fg != tcell.ColorBlack {
		t.Errorf("Expected default focused fg=Black, got %v", fg)
	}
	if bg != tcell.ColorYellow {
		t.Errorf("Expected default focused bg=Yellow, got %v", bg)
	}
}

func TestButtonDisabledNotFocusable(t *testing.T) {
	btn := NewButton(Style{ID: "btn", Focusable: true}, "Click", func() {}).WithDisabled(true)
	if btn.IsFocusable() {
		t.Error("Disabled button should not be focusable")
	}
}

func TestButtonDisabledIgnoresEvents(t *testing.T) {
	clicked := false
	btn := NewButton(Style{}, "Click", func() { clicked = true }).WithDisabled(true)

	ev := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	consumed := btn.HandleEvent(ev, nil, EventContext{})

	if consumed {
		t.Error("Disabled button should not consume events")
	}
	if clicked {
		t.Error("Disabled button should not fire OnClick")
	}
}

func TestButtonCenteredLabel(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	// Width=20, label "[ Hi ]" = 6 chars, so offset = (20-6)/2 = 7
	btn := NewButton(Style{Width: 20}, "Hi", func() {})
	layout := Layout(btn, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(30, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	expectedOffset := (20 - len("[ Hi ]")) / 2
	str, _, _ := s.Get(expectedOffset, 0)
	if str != "[" {
		t.Errorf("Expected '[' at col %d (centered), got '%s'", expectedOffset, str)
	}
}
