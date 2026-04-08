package splotch

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestLayoutDropdown(t *testing.T) {
	drp := NewDropdown(Style{ID: "drp"}, []string{"Option 1", "Option Longest"}, 0, nil)
	
	res := Layout(drp, 0, 0, Constraints{MaxW: 100, MaxH: 100})
	
	expectedW := len("Option Longest") + 6 // 14 + 6 = 20
	if res.W != expectedW {
		t.Errorf("Expected width %d, got %d", expectedW, res.W)
	}
	
	if res.H != 1 {
		t.Errorf("Expected height 1, got %d", res.H)
	}
}

func TestFindNodeByID_Dropdown(t *testing.T) {
	drp := NewDropdown(Style{ID: "drp"}, []string{"Option 1"}, 0, nil)
	root := NewBox(Style{ID: "box"}, drp)
	
	found := findNodeByID(root, "drp")
	if found != drp {
		t.Errorf("Expected to find dropdown node, got %v", found)
	}
}

func TestDropdownScrolling(t *testing.T) {
	app := &App{
		componentStates: make(map[string]interface{}),
		focusedID:       "mydropdown",
	}

	drp := NewDropdown(Style{ID: "mydropdown"}, []string{"1", "2", "3", "4", "5", "6"}, 0, nil)
	state := &DropdownState{Open: true}
	app.componentStates["mydropdown"] = state

	// Press Down 5 times (reaches index 5, which is 6th item)
	for i := 0; i < 5; i++ {
		ev := tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
		app.handleKeyEvent(ev, drp, LayoutResult{Node: drp}, []string{"mydropdown"})
	}

	if state.FocusedIndex != 5 {
		t.Errorf("Expected FocusedIndex 5, got %d", state.FocusedIndex)
	}

	// ScrollOffset should have advanced to 1! (FocusedIndex 5 >= 0 + 5)
	if state.ScrollOffset != 1 {
		t.Errorf("Expected ScrollOffset 1, got %d", state.ScrollOffset)
	}

	// Press Down again (should NOT wrap, should stay at 5)
	ev := tcell.NewEventKey(tcell.KeyDown, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, drp, LayoutResult{Node: drp}, []string{"mydropdown"})

	if state.FocusedIndex != 5 {
		t.Errorf("Expected FocusedIndex 5, got %d", state.FocusedIndex)
	}

	// ScrollOffset should remain 1!
	if state.ScrollOffset != 1 {
		t.Errorf("Expected ScrollOffset 1, got %d", state.ScrollOffset)
	}
}

func TestDropdownPageUpDown(t *testing.T) {
	app := &App{
		componentStates: make(map[string]interface{}),
		focusedID:       "mydropdown",
	}

	drp := NewDropdown(Style{ID: "mydropdown"}, []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}, 0, nil)
	state := &DropdownState{Open: true}
	app.componentStates["mydropdown"] = state

	// Default limit is 5.
	
	// Press Page Down
	ev := tcell.NewEventKey(tcell.KeyPgDn, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, drp, LayoutResult{Node: drp}, []string{"mydropdown"})
	
	if state.FocusedIndex != 5 {
		t.Errorf("Expected FocusedIndex 5, got %d", state.FocusedIndex)
	}
	
	if state.ScrollOffset != 5 {
		t.Errorf("Expected ScrollOffset 5, got %d", state.ScrollOffset)
	}
	
	// Press Page Up
	ev = tcell.NewEventKey(tcell.KeyPgUp, ' ', tcell.ModNone)
	app.handleKeyEvent(ev, drp, LayoutResult{Node: drp}, []string{"mydropdown"})
	
	if state.FocusedIndex != 0 {
		t.Errorf("Expected FocusedIndex 0, got %d", state.FocusedIndex)
	}
	
	if state.ScrollOffset != 0 {
		t.Errorf("Expected ScrollOffset 0, got %d", state.ScrollOffset)
	}
}
