package splotch

import (
	"github.com/gdamore/tcell/v2"
	"testing"
)

func TestMenuBarLayout(t *testing.T) {
	ctx := &RenderContext{app: &App{componentStates: make(map[string]any)}}
	mb := NewMenuBar(ctx, Style{FillWidth: true}, []Menu{
		{Title: "File"},
		{Title: "Edit"},
	})

	res := Layout(mb, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	if res.W != 80 {
		t.Errorf("Expected width 80, got %d", res.W)
	}
	if res.H != 1 {
		t.Errorf("Expected height 1, got %d", res.H)
	}
}

func TestMenuBarLayoutAutoWidth(t *testing.T) {
	ctx := &RenderContext{app: &App{componentStates: make(map[string]any)}}
	mb := NewMenuBar(ctx, Style{}, []Menu{
		{Title: "File"}, // len 4 + 4 = 8
		{Title: "Edit"}, // len 4 + 4 = 8
	})

	res := Layout(mb, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	expectedW := 16
	if res.W != expectedW {
		t.Errorf("Expected width %d, got %d", expectedW, res.W)
	}
}

func TestMenuBarInteraction(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	defer s.Fini()

	app := &App{
		screen:          s,
		componentStates: make(map[string]any),
		focusedID:       "hook-0", // Focus the menu bar (auto-generated ID)
	}
	ctx := &RenderContext{app: app}

	actionTriggered := false
	mb := NewMenuBar(ctx, Style{Focusable: true}, []Menu{
		{Title: "File", AltRune: 'f', Items: []MenuItem{
			{Label: "New", Action: func() { actionTriggered = true }},
		}},
	})

	app.componentStates["hook-0"] = &MenuBarState{OpenMenuIndex: -1, FocusedItemIndex: -1}

	root := mb
	layout := Layout(root, 0, 0, Constraints{MaxW: 80, MaxH: 24})
	focusableIDs := []string{"menubar"}

	// Test direct letter shortcut 'f'
	ev := tcell.NewEventKey(tcell.KeyRune, 'f', tcell.ModNone)
	handled := app.handleKeyEvent(ev, root, layout, focusableIDs)
	if handled {
		t.Error("Expected handleKeyEvent to return false (no exit) for 'f'")
	}

	state := app.componentStates["hook-0"].(*MenuBarState)
	if state.OpenMenuIndex != 0 {
		t.Errorf("Expected open menu index 0, got %d", state.OpenMenuIndex)
	}

	// Test arrow down to select item
	ev = tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone)
	app.handleKeyEvent(ev, root, layout, focusableIDs)
	if state.FocusedItemIndex != 0 {
		t.Errorf("Expected focused item index 0, got %d", state.FocusedItemIndex)
	}

	// Test Enter to trigger action
	ev = tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	app.handleKeyEvent(ev, root, layout, focusableIDs)
	if !actionTriggered {
		t.Error("Expected action to be triggered")
	}
	if state.OpenMenuIndex != -1 {
		t.Error("Expected menu to close after action")
	}
}
