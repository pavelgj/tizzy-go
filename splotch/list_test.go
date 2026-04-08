package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutList(t *testing.T) {
	ctx := makeTestContext()
	list := NewList(ctx, Style{ID: "list"}, []any{"Item 1", "Item 2"}, func(item any, index int, selected bool, cursor bool) Node {
		return NewText(Style{}, item.(string))
	}, nil)

	res := Layout(list, 0, 0, Constraints{MaxW: 100, MaxH: 100})

	if res.W != 100 {
		t.Errorf("Expected width 100, got %d", res.W)
	}

	if res.H != 100 {
		t.Errorf("Expected height 100, got %d", res.H)
	}
}

func TestListArrowNavigation(t *testing.T) {
	app := &App{
		componentStates: make(map[string]any),
		focusedID:       "mylist",
	}

	list := &List{
		Style: Style{ID: "mylist"},
		Items: []any{"1", "2", "3"},
	}
	state := &ListState{SelectedIndex: -1, CursorIndex: 0}
	app.componentStates["mylist"] = state

	// Simulate KeyDown
	ev := tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, list, LayoutResult{Node: list, H: 10}, []string{"mylist"})

	if state.CursorIndex != 1 {
		t.Errorf("Expected CursorIndex 1, got %d", state.CursorIndex)
	}
	if state.SelectedIndex != -1 {
		t.Errorf("Expected SelectedIndex -1, got %d", state.SelectedIndex)
	}

	// Simulate KeyUp
	ev = tcell.NewEventKey(tcell.KeyUp, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, list, LayoutResult{Node: list, H: 10}, []string{"mylist"})

	if state.CursorIndex != 0 {
		t.Errorf("Expected CursorIndex 0, got %d", state.CursorIndex)
	}
}

func TestListEnterSelection(t *testing.T) {
	app := &App{
		componentStates: make(map[string]any),
		focusedID:       "mylist",
	}

	var selectedIdx = -1
	list := &List{
		Style: Style{ID: "mylist"},
		Items: []any{"1", "2", "3"},
		OnSelect: func(idx int) {
			selectedIdx = idx
		},
	}
	state := &ListState{SelectedIndex: -1, CursorIndex: 1}
	app.componentStates["mylist"] = state

	// Simulate KeyEnter
	ev := tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, list, LayoutResult{Node: list, H: 10}, []string{"mylist"})

	if state.SelectedIndex != 1 {
		t.Errorf("Expected SelectedIndex 1, got %d", state.SelectedIndex)
	}
	if selectedIdx != 1 {
		t.Errorf("Expected OnSelect to be called with 1, got %d", selectedIdx)
	}
}

func TestFindListAt(t *testing.T) {
	list := &List{Style: Style{ID: "mylist"}}
	res := LayoutResult{
		Node: list,
		X:    10,
		Y:    10,
		W:    20,
		H:    20,
	}

	found := findListAt(res, 15, 15, nil)
	if found != list {
		t.Errorf("Expected to find list, got %v", found)
	}

	found = findListAt(res, 5, 5, nil)
	if found != nil {
		t.Errorf("Expected not to find list, got %v", found)
	}
}

func TestListPageUpDown(t *testing.T) {
	app := &App{
		componentStates: make(map[string]any),
		focusedID:       "mylist",
	}

	list := &List{
		Style: Style{ID: "mylist"},
		Items: []any{"1", "2", "3", "4", "5"},
	}
	state := &ListState{SelectedIndex: -1, CursorIndex: 0}
	app.componentStates["mylist"] = state

	// Simulate KeyPgDn (viewportH is 2 if H=2 and no border)
	ev := tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, list, LayoutResult{Node: list, H: 2}, []string{"mylist"})

	if state.CursorIndex != 2 {
		t.Errorf("Expected CursorIndex 2, got %d", state.CursorIndex)
	}

	// Simulate KeyPgUp
	ev = tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, list, LayoutResult{Node: list, H: 2}, []string{"mylist"})

	if state.CursorIndex != 0 {
		t.Errorf("Expected CursorIndex 0, got %d", state.CursorIndex)
	}
}
