package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutCheckbox(t *testing.T) {
	ctx := makeTestContext()
	cb := NewCheckbox(ctx, Style{}, "Check", false, nil)
	res := Layout(cb, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 9 || res.H != 1 {
		t.Errorf("Expected W=9, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderCheckbox(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	ctx := makeTestContext()
	cb := NewCheckbox(ctx, Style{}, "Check", true, nil)
	layout := Layout(cb, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	expected := "[x] Check"
	for i, r := range expected {
		str, _, _ := s.Get(i, 0)
		if str != string(r) {
			t.Errorf("Expected '%c' at %d,0, got '%s'", r, i, str)
		}
	}
}
