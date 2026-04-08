package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutSpinner(t *testing.T) {
	spinner := NewSpinner(Style{})
	res := Layout(spinner, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 1 || res.H != 1 {
		t.Errorf("Expected W=1, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderSpinner(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	spinner := NewSpinner(Style{})
	layout := Layout(spinner, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	mainc, _, _, _ := s.GetContent(0, 0)
	validFrames := map[rune]bool{'|': true, '/': true, '-': true, '\\': true}
	if !validFrames[mainc] {
		t.Errorf("Expected one of '|', '/', '-', '\\', got '%c'", mainc)
	}
}
