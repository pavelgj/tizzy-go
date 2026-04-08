package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutRadioButton(t *testing.T) {
	rb := NewRadioButton(Style{}, "Option", "val", false, nil)
	res := Layout(rb, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 10 || res.H != 1 {
		t.Errorf("Expected W=10, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderRadioButton(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	rb := NewRadioButton(Style{}, "Option", "val", true, nil)
	layout := Layout(rb, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	expected := "(*) Option"
	for i, r := range expected {
		mainc, _, _, _ := s.GetContent(i, 0)
		if mainc != r {
			t.Errorf("Expected '%c' at %d,0, got '%c'", r, i, mainc)
		}
	}
}
