package tz

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutSpinner(t *testing.T) {
	ctx := makeTestContext()
	spinner := NewSpinner(ctx, Style{})
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

	ctx := makeTestContext()
	spinner := NewSpinner(ctx, Style{})
	layout := Layout(spinner, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	str, _, _ := s.Get(0, 0)
	validFrames := map[string]bool{"|": true, "/": true, "-": true, "\\": true}
	if !validFrames[str] {
		t.Errorf("Expected one of '|', '/', '-', '\\', got '%s'", str)
	}
}
