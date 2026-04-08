package splotch

import (
	"testing"
	"github.com/gdamore/tcell/v2"
)

func TestFindFocusableIDs(t *testing.T) {
	ctx := makeTestContext()
	root := NewBox(Style{},
		NewText(Style{ID: "t1", Focusable: true}, "Text 1"),
		NewBox(Style{ID: "b1", Focusable: true},
			NewText(Style{ID: "t2", Focusable: true}, "Text 2"),
		),
		NewScrollView(ctx, Style{ID: "s1", Focusable: true},
			NewText(Style{ID: "t4", Focusable: true}, "Text 4"),
		),
		NewText(Style{ID: "t3"}, "Not Focusable"), // Should not be found
	)

	ids := findFocusableIDs(root, nil)

	expected := []string{"t1", "b1", "t2", "s1", "t4"}
	if len(ids) != len(expected) {
		t.Fatalf("Expected %d IDs, got %d", len(expected), len(ids))
	}
	for i, id := range ids {
		if id != expected[i] {
			t.Errorf("Expected ID at %d to be %s, got %s", i, expected[i], id)
		}
	}
}

func TestNextPrevFocus(t *testing.T) {
	ids := []string{"1", "2", "3"}

	if nextFocus("1", ids) != "2" {
		t.Errorf("Expected next after 1 to be 2")
	}
	if nextFocus("3", ids) != "1" {
		t.Errorf("Expected next after 3 to be 1")
	}
	if nextFocus("", ids) != "1" {
		t.Errorf("Expected next after empty to be 1")
	}

	if prevFocus("2", ids) != "1" {
		t.Errorf("Expected prev before 2 to be 1")
	}
	if prevFocus("1", ids) != "3" {
		t.Errorf("Expected prev before 1 to be 3")
	}
	if prevFocus("", ids) != "3" {
		t.Errorf("Expected prev before empty to be 3")
	}
}

func TestOffsetToLineCol(t *testing.T) {
	text := "abc\ndef\nghi"
	
	tests := []struct {
		offset int
		line   int
		col    int
	}{
		{0, 0, 0},
		{1, 0, 1},
		{3, 0, 3},
		{4, 1, 0},
		{5, 1, 1},
		{7, 1, 3},
		{8, 2, 0},
		{9, 2, 1},
		{11, 2, 3},
		{12, 2, 3}, // Beyond end
	}
	
	for _, tc := range tests {
		l, c := offsetToLineCol(text, tc.offset)
		if l != tc.line || c != tc.col {
			t.Errorf("For offset %d, expected line %d, col %d; got line %d, col %d", tc.offset, tc.line, tc.col, l, c)
		}
	}
}

func TestLineColToOffset(t *testing.T) {
	text := "abc\ndef\nghi"
	
	tests := []struct {
		line   int
		col    int
		offset int
	}{
		{0, 0, 0},
		{0, 1, 1},
		{0, 3, 3},
		{0, 5, 3}, // Clamped
		{1, 0, 4},
		{1, 1, 5},
		{1, 3, 7},
		{2, 0, 8},
		{2, 3, 11},
		{3, 0, 11}, // Beyond lines
	}
	
	for _, tc := range tests {
		off := lineColToOffset(text, tc.line, tc.col)
		if off != tc.offset {
			t.Errorf("For line %d, col %d, expected offset %d; got %d", tc.line, tc.col, tc.offset, off)
		}
	}
}

func TestScrollViewKeyboardScrolling(t *testing.T) {
	app := &App{
		componentStates: make(map[string]any),
		focusedID:       "sv",
	}
	
	ctx := &RenderContext{app: app}
	sv := NewScrollView(ctx, Style{ID: "sv", Focusable: true}, NewText(Style{}, "Content"))
	root := NewBox(Style{}, sv)
	
	layout := Layout(root, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	
	ev := tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	
	exit := app.handleKeyEvent(ev, root, layout, []string{"sv"})
	
	if exit {
		t.Errorf("Expected handleKeyEvent to return false, got true")
	}
	
	stateObj, ok := app.componentStates["sv"]
	if !ok {
		t.Fatalf("Expected state for 'sv' to be created")
	}
	state := stateObj.(*ScrollViewState)
	if state.ScrollOffset != 1 {
		t.Errorf("Expected ScrollOffset to be 1, got %d", state.ScrollOffset)
	}
}

func TestFindLayoutResultByID(t *testing.T) {
	ctx := makeTestContext()
	list := NewList(ctx, Style{ID: "list-1"}, "key", []any{"item1"}, nil, nil)
	grid := NewGridBox(Style{ID: "grid-1"},
		[]GridTrack{Flex(1)},
		[]GridTrack{Flex(1)},
		NewBox(Style{ID: "box-1"}, list),
	)

	layout := Layout(grid, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	res := findLayoutResultByID(layout, "list-1")
	if res == nil {
		t.Fatalf("Expected to find layout result for 'list-1'")
	}
	if res.Node != list {
		t.Errorf("Expected found node to be the list, got %T", res.Node)
	}
}

func TestListOnFocus(t *testing.T) {
	app := &App{
		componentStates: make(map[string]any),
	}
	ctx := &RenderContext{app: app}
	
	focusCalled := false
	list := NewList(ctx, Style{ID: "list-1", Focusable: true}, "key", []any{"item1"}, nil, nil)
	list.OnFocus = func(state *ListState) {
		focusCalled = true
		state.CursorIndex = 99
	}
	
	root := NewBox(Style{}, list)
	
	app.setFocus("list-1", root)
	
	if !focusCalled {
		t.Errorf("Expected OnFocus to be called")
	}
	
	stateObj, ok := app.componentStates["list-1"]
	if !ok {
		t.Fatalf("Expected state for 'list-1' to be created")
	}
	state := stateObj.(*ListState)
	if state.CursorIndex != 99 {
		t.Errorf("Expected CursorIndex to be 99, got %d", state.CursorIndex)
	}
}


