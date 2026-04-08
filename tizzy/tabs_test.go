package tizzy

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestTabsLayout(t *testing.T) {
	ctx := makeTestContext()
	tabs := NewTabs(ctx, Style{ID: "tabs"}, []Tab{
		{Label: "Home", Content: NewText(Style{}, "Home Content")},
		{Label: "About", Content: NewText(Style{}, "About Content")},
	})

	// Headers width: len("Home") + 4 = 8, len("About") + 4 = 9. Total = 17.
	// Content width: "Home Content" len 12, "About Content" len 13. Max = 13.
	// So total width should be max(17, 13) = 17!

	res := Layout(tabs, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	if res.W != 17 {
		t.Errorf("Expected width 17, got %d", res.W)
	}
	// Height: header (1) + content (1) = 2!
	if res.H != 2 {
		t.Errorf("Expected height 2, got %d", res.H)
	}
}

func TestTabsInteraction(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Fini()

	app := &App{
		screen:          s,
		componentStates: make(map[string]any),
		focusedID:       "tabs",
	}

	ctx := &RenderContext{app: app}
	tabs := NewTabs(ctx, Style{ID: "tabs", Focusable: true}, []Tab{
		{Label: "Home", Content: NewText(Style{}, "Home Content")},
		{Label: "About", Content: NewText(Style{}, "About Content")},
	})
	app.componentStates["tabs"] = &TabsState{ActiveTab: 0}

	root := tabs
	layout := Layout(root, 0, 0, Constraints{MaxW: 80, MaxH: 24})
	focusableIDs := []string{"tabs"}

	// Test Right arrow to switch tab
	ev := tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone)
	handled := app.handleKeyEvent(ev, root, layout, focusableIDs)
	if handled {
		t.Error("Expected handleKeyEvent to return false (no exit)")
	}

	state := app.componentStates["tabs"].(*TabsState)
	if state.ActiveTab != 1 {
		t.Errorf("Expected active tab 1, got %d", state.ActiveTab)
	}

	// Test Left arrow to switch back
	ev = tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone)
	app.handleKeyEvent(ev, root, layout, focusableIDs)
	if state.ActiveTab != 0 {
		t.Errorf("Expected active tab 0, got %d", state.ActiveTab)
	}
}
