package tizzy

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutProgressBar(t *testing.T) {
	pb := NewProgressBar(Style{Width: 30}, 0.5)
	res := Layout(pb, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 30 || res.H != 1 {
		t.Errorf("Expected W=30, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestRenderProgressBar(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	pb := NewProgressBar(Style{Width: 10}, 0.5)
	layout := Layout(pb, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(20, 5)
	renderToScreen(s, layout, "", nil)
	s.Show()

	for i := 0; i < 5; i++ {
		str, _, _ := s.Get(i, 0)
		if str != "█" {
			t.Errorf("At col %d, expected '█', got '%s'", i, str)
		}
	}
	for i := 5; i < 10; i++ {
		str, _, _ := s.Get(i, 0)
		if str != "░" {
			t.Errorf("At col %d, expected '░', got '%s'", i, str)
		}
	}
}
