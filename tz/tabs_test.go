package tz

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
	// Height: header (2 rows: label + separator) + content (1) = 3!
	if res.H != 3 {
		t.Errorf("Expected height 3, got %d", res.H)
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

// TestTabsLastVisible verifies the visible range calculation for overflow tabs.
func TestTabsLastVisible(t *testing.T) {
	// Each tab slot = len(label) + 4.
	// "Tab A"=5+4=9, "Tab B"=5+4=9, "Tab C"=5+4=9, "Tab D"=5+4=9 → total 36
	tabs := []Tab{
		{Label: "Tab A"},
		{Label: "Tab B"},
		{Label: "Tab C"},
		{Label: "Tab D"},
	}

	// availW=20, scrollOffset=0: no left arrow.
	// Without right arrow: 9+9=18 ≤ 20, next 9 would make 27 > 20 → last=1 (need right arrow)
	// With right arrow (reserve 1): 9+9=18, 18+9+1=28 > 20 → last=1 still
	last := tabsLastVisible(tabs, 0, 20)
	if last != 1 {
		t.Errorf("scrollOffset=0 availW=20: expected last=1, got %d", last)
	}

	// scrollOffset=1: left arrow costs 1. Available for tabs = 20-1=19, minus right arrow=18.
	// 9 ≤ 18, next 9+9=18 ≤ 18 fits! but then no more tabs after index 2... wait index 3 remains.
	// With left arrow (1) + right arrow (1): 20-2=18. Tabs 1,2 = 9+9=18 fits. Tab 3 = 9 > 0 remaining → last=2
	last = tabsLastVisible(tabs, 1, 20)
	if last != 2 {
		t.Errorf("scrollOffset=1 availW=20: expected last=2, got %d", last)
	}

	// scrollOffset=2: left arrow costs 1. Tabs 2,3 = 9+9=18, need right? No, tab 3 is last.
	// Without right arrow: 1+9+9=19 ≤ 20 → last=3 (all remaining fit)
	last = tabsLastVisible(tabs, 2, 20)
	if last != 3 {
		t.Errorf("scrollOffset=2 availW=20: expected last=3, got %d", last)
	}
}

// TestTabsScrollOffset verifies that ensureActiveVisible adjusts ScrollOffset correctly.
func TestTabsScrollOffset(t *testing.T) {
	ctx := makeTestContext()
	tabs := NewTabs(ctx, Style{ID: "tabs"}, []Tab{
		{Label: "Tab A"},
		{Label: "Tab B"},
		{Label: "Tab C"},
		{Label: "Tab D"},
	})
	// availW=20, scrollOffset=0 shows tabs 0-1 (plus right arrow)

	t.Run("navigate right scrolls into view", func(t *testing.T) {
		s := &TabsState{ActiveTab: 1, ScrollOffset: 0}
		// Move to tab 2 — it's not visible at offset=0 with availW=20
		s.ActiveTab = 2
		tabs.ensureActiveVisible(s, 20)
		if s.ScrollOffset == 0 {
			t.Error("Expected ScrollOffset to advance so tab 2 is visible")
		}
	})

	t.Run("navigate left scrolls back", func(t *testing.T) {
		s := &TabsState{ActiveTab: 2, ScrollOffset: 1}
		s.ActiveTab = 0
		tabs.ensureActiveVisible(s, 20)
		if s.ScrollOffset != 0 {
			t.Errorf("Expected ScrollOffset=0, got %d", s.ScrollOffset)
		}
	})
}
