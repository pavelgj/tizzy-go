package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

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
	res := Layout(box, 0, 0, Constraints{MaxW: 10, MaxH: 1})

	if len(res.Children) != 1 {
		t.Fatalf("Expected 1 child, got %d", len(res.Children))
	}
	if res.Children[0].X != 4 {
		t.Errorf("Expected child X=4, got %d", res.Children[0].X)
	}
}

func TestLayoutMarginAccumulation(t *testing.T) {
	root := NewBox(Style{FlexDirection: "row"},
		NewText(Style{Margin: Margin{Left: 2, Right: 1}}, "A"),
		NewText(Style{Margin: Margin{Left: 3, Right: 2}}, "B"),
	)

	res := Layout(root, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if len(res.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(res.Children))
	}

	c0 := res.Children[0]
	if c0.X != 2 {
		t.Errorf("Expected child 0 X=2, got %d", c0.X)
	}

	c1 := res.Children[1]
	if c1.X != 7 {
		t.Errorf("Expected child 1 X=7, got %d", c1.X)
	}
}

func TestLayoutFillWidth(t *testing.T) {
	text := NewText(Style{}, "Hello")
	box := NewBox(Style{FillWidth: true}, text)
	res := Layout(box, 0, 0, Constraints{MaxW: 50, MaxH: 100})

	if res.W != 50 {
		t.Errorf("Expected W=50, got %d", res.W)
	}
}

func TestLayoutFillWidthRow(t *testing.T) {
	sidebar := NewBox(Style{Width: 20}, NewText(Style{}, "Sidebar"))
	content := NewBox(Style{FillWidth: true}, NewText(Style{}, "Content"))
	root := NewBox(Style{FlexDirection: "row", FillWidth: true}, sidebar, content)

	res := Layout(root, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 100 {
		t.Errorf("Expected root W=100, got %d", res.W)
	}
	if len(res.Children) != 2 {
		t.Fatalf("Expected 2 children, got %d", len(res.Children))
	}

	sidebarRes := res.Children[0]
	contentRes := res.Children[1]

	if sidebarRes.W != 7 {
		t.Errorf("Expected sidebar W=7, got %d", sidebarRes.W)
	}

	if contentRes.W != 93 {
		t.Errorf("Expected content W=93, got %d", contentRes.W)
	}
}

func TestRenderBorder(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	box := NewBox(Style{Border: true},
		NewText(Style{}, "Hi"),
	)

	layout := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 10)

	renderToScreen(s, layout, "", nil)
	s.Show()

	mainc, _, _, _ := s.GetContent(0, 0)
	if mainc != '┌' {
		t.Errorf("Expected '┌' at 0,0, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(1, 0)
	if mainc != '─' {
		t.Errorf("Expected '─' at 1,0, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(1, 1)
	if mainc != 'H' {
		t.Errorf("Expected 'H' at 1,1, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(2, 1)
	if mainc != 'i' {
		t.Errorf("Expected 'i' at 2,1, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(3, 2)
	if mainc != '┘' {
		t.Errorf("Expected '┘' at 3,2, got '%c'", mainc)
	}
}

func TestRenderFocusStyle(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	box := NewBox(Style{ID: "box-1", Border: true, FocusColor: tcell.ColorRed},
		NewText(Style{}, "Hi"),
	)

	layout := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 10)

	renderToScreen(s, layout, "box-1", nil)
	s.Show()

	_, _, style, _ := s.GetContent(0, 0)
	expectedStyle := tcell.StyleDefault.Foreground(tcell.ColorRed)
	if style != expectedStyle {
		t.Errorf("Expected focus style %v, got %v", expectedStyle, style)
	}
}

func TestRenderTitledBorder(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}

	box := NewBox(Style{Border: true, Title: "Title"},
		NewText(Style{}, "Hello World Wide"),
	)

	layout := Layout(box, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	s.SetSize(10, 10)

	renderToScreen(s, layout, "", nil)
	s.Show()

	mainc, _, _, _ := s.GetContent(0, 0)
	if mainc != '┌' {
		t.Errorf("Expected '┌' at 0,0, got '%c'", mainc)
	}

	mainc, _, _, _ = s.GetContent(1, 0)
	if mainc != ' ' {
		t.Errorf("Expected ' ' at 1,0, got '%c'", mainc)
	}

	expectedTitle := "Title"
	for i, r := range expectedTitle {
		mainc, _, _, _ = s.GetContent(2+i, 0)
		if mainc != r {
			t.Errorf("Expected '%c' at %d,0, got '%c'", r, 2+i, mainc)
		}
	}

	mainc, _, _, _ = s.GetContent(2+len(expectedTitle), 0)
	if mainc != ' ' {
		t.Errorf("Expected ' ' at %d,0, got '%c'", 2+len(expectedTitle), mainc)
	}
}
