package splotch

import "testing"

func TestLayoutTextInput(t *testing.T) {
	input := NewTextInput(Style{Padding: Padding{Left: 1}}, "hello", nil)
	
	res := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	
	if res.W != 6 { // 5 (len) + 1 (left pad)
		t.Errorf("Expected width 6, got %d", res.W)
	}
	if res.H != 1 {
		t.Errorf("Expected height 1, got %d", res.H)
	}
}

func TestLayoutTextInputFixedWidth(t *testing.T) {
	input := NewTextInput(Style{Width: 10}, "hello", nil)
	
	res := Layout(input, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	
	if res.W != 10 {
		t.Errorf("Expected width 10, got %d", res.W)
	}
}
