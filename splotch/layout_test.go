package splotch

import "testing"

func TestLayoutText(t *testing.T) {
	text := NewText(Style{}, "Hello")
	res := Layout(text, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected X=10, Y=20, got X=%d, Y=%d", res.X, res.Y)
	}
	if res.W != 5 || res.H != 1 {
		t.Errorf("Expected W=5, H=1, got W=%d, H=%d", res.W, res.H)
	}
}

func TestLayoutBoxRow(t *testing.T) {
	box := NewBox(Style{FlexDirection: "row"},
		NewText(Style{}, "Hi"),
		NewText(Style{}, "There"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 7 { // 2 + 5
		t.Errorf("Expected width 7, got %d", res.W)
	}
	if res.H != 1 {
		t.Errorf("Expected height 1, got %d", res.H)
	}
	if len(res.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(res.Children))
	}
	if res.Children[1].X != 2 {
		t.Errorf("Expected second child X=2, got X=%d", res.Children[1].X)
	}
}

func TestLayoutBoxColumn(t *testing.T) {
	box := NewBox(Style{FlexDirection: "column"},
		NewText(Style{}, "Hi"),
		NewText(Style{}, "There"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 5 { // max(2, 5)
		t.Errorf("Expected width 5, got %d", res.W)
	}
	if res.H != 2 { // 1 + 1
		t.Errorf("Expected height 2, got %d", res.H)
	}
	if len(res.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(res.Children))
	}
	if res.Children[1].Y != 1 {
		t.Errorf("Expected second child Y=1, got Y=%d", res.Children[1].Y)
	}
}
func TestLayoutBoxWithBorder(t *testing.T) {
	box := NewBox(Style{Border: true},
		NewText(Style{}, "Hi"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 4 { // 2 (text) + 2 (borders)
		t.Errorf("Expected width 4, got %d", res.W)
	}
	if res.H != 3 { // 1 (text) + 2 (borders)
		t.Errorf("Expected height 3, got %d", res.H)
	}
	if len(res.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(res.Children))
	}
	if res.Children[0].X != 1 || res.Children[0].Y != 1 {
		t.Errorf("Expected child at 1,1, got %d,%d", res.Children[0].X, res.Children[0].Y)
	}
}

func TestLayoutBoxWithPadding(t *testing.T) {
	box := NewBox(Style{Padding: Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
		NewText(Style{}, "Hi"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 6 { // 2 (text) + 4 (padding)
		t.Errorf("Expected width 6, got %d", res.W)
	}
	if res.H != 3 { // 1 (text) + 2 (padding)
		t.Errorf("Expected height 3, got %d", res.H)
	}
	if res.Children[0].X != 2 || res.Children[0].Y != 1 {
		t.Errorf("Expected child at 2,1, got %d,%d", res.Children[0].X, res.Children[0].Y)
	}
}

func TestLayoutBoxWithMargin(t *testing.T) {
	box := NewBox(Style{Margin: Margin{Left: 2, Top: 1}},
		NewText(Style{}, "Hi"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 2 || res.Y != 1 {
		t.Errorf("Expected box at 2,1, got %d,%d", res.X, res.Y)
	}
	if res.W != 2 {
		t.Errorf("Expected width 2, got %d", res.W)
	}
}

func TestLayoutBoxWithMarginChildren(t *testing.T) {
	box := NewBox(Style{FlexDirection: "row"},
		NewText(Style{Margin: Margin{Right: 2}}, "Hi"),
		NewText(Style{}, "There"),
	)
	res := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if len(res.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(res.Children))
	}
	if res.Children[0].X != 0 {
		t.Errorf("Expected first child X=0, got %d", res.Children[0].X)
	}
	if res.Children[1].X != 4 {
		t.Errorf("Expected second child X=4, got %d", res.Children[1].X)
	}
}

func TestLayoutCenter(t *testing.T) {
	box := NewBox(Style{FlexDirection: "row", JustifyContent: "center"},
		NewText(Style{}, "Hi"),
	)
	// MaxW 10, content width 2. Remaining 8 / 2 = 4 shift.
	res := Layout(box, 0, 0, Constraints{MaxW: 10, MaxH: 1})

	if len(res.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(res.Children))
	}
	if res.Children[0].X != 4 {
		t.Errorf("Expected child X=4, got %d", res.Children[0].X)
	}
}

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


