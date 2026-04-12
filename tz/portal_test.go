package tz

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

// ---------------------------------------------------------------------------
// collectPortals
// ---------------------------------------------------------------------------

func TestCollectPortals_NoPortals(t *testing.T) {
	root := NewBox(Style{},
		NewText(Style{}, "hello"),
	)
	layout := Layout(root, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	var portals []collectedPortal
	collectPortals(layout, 80, 24, layout, &portals)

	if len(portals) != 0 {
		t.Errorf("expected 0 portals, got %d", len(portals))
	}
}

func TestCollectPortals_DirectChild(t *testing.T) {
	portal := &Portal{
		Child: NewText(Style{}, "popup"),
		X:     10, Y: 5,
	}
	root := NewBox(Style{}, portal)
	layout := Layout(root, 0, 0, Constraints{MaxW: 80, MaxH: 24})

	var portals []collectedPortal
	collectPortals(layout, 80, 24, layout, &portals)

	if len(portals) != 1 {
		t.Fatalf("expected 1 portal, got %d", len(portals))
	}
	// Content should be laid out at X=10, Y=5.
	if portals[0].content.X != 10 || portals[0].content.Y != 5 {
		t.Errorf("portal content at (%d,%d), want (10,5)", portals[0].content.X, portals[0].content.Y)
	}
}

func TestCollectPortals_NestedInDropdown(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	app := NewAppWithScreen(screen)
	app.componentStates["dd"] = &DropdownState{Open: true, FocusedIndex: 0}

	renderFn := func(ctx *RenderContext) Node {
		return NewDropdown(ctx, Style{ID: "dd", Focusable: true},
			[]string{"A", "B"}, 0, nil)
	}

	_, _, layout, _, err := app.RenderFrame(renderFn)
	if err != nil {
		t.Fatal(err)
	}

	var portals []collectedPortal
	collectPortals(layout, 80, 24, layout, &portals)

	if len(portals) != 1 {
		t.Fatalf("expected 1 portal from open dropdown, got %d", len(portals))
	}
}

// ---------------------------------------------------------------------------
// Portal auto-centering
// ---------------------------------------------------------------------------

func TestPortalAutoCentre(t *testing.T) {
	child := NewBox(Style{Width: 20, Height: 6}, NewText(Style{}, ""))
	portal := &Portal{Child: child, X: -1, Y: -1}

	content := computePortalLayout(portal, 80, 24, LayoutResult{})

	wantX := (80 - 20) / 2 // 30
	wantY := (24 - 6) / 2  // 9
	if content.X != wantX || content.Y != wantY {
		t.Errorf("auto-centred portal at (%d,%d), want (%d,%d)", content.X, content.Y, wantX, wantY)
	}
}

// ---------------------------------------------------------------------------
// OnOutsideClick
// ---------------------------------------------------------------------------

func TestPortalOnOutsideClick(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	dismissed := false
	portal := &Portal{
		Child:          NewBox(Style{Width: 10, Height: 5}, NewText(Style{}, "")),
		X:              10, Y: 10,
		OnOutsideClick: func() { dismissed = true },
	}
	root := NewBox(Style{}, portal)

	app := NewAppWithScreen(screen)
	_, _, layout, _, err := app.RenderFrame(func(ctx *RenderContext) Node { return root })
	if err != nil {
		t.Fatal(err)
	}

	// Click well outside the portal bounds (portal occupies roughly 10,10 to 20,15).
	ev := tcell.NewEventMouse(1, 1, tcell.Button1, tcell.ModNone)
	app.handleMouseEvent(ev, root, layout)

	if !dismissed {
		t.Error("expected OnOutsideClick to be called on click outside portal")
	}
}

func TestPortalInsideClickDoesNotTriggerOutsideClick(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	dismissed := false
	portal := &Portal{
		Child:          NewBox(Style{Width: 10, Height: 5}, NewText(Style{}, "")),
		X:              0, Y: 0,
		OnOutsideClick: func() { dismissed = true },
	}
	root := NewBox(Style{}, portal)

	app := NewAppWithScreen(screen)
	_, _, layout, _, err := app.RenderFrame(func(ctx *RenderContext) Node { return root })
	if err != nil {
		t.Fatal(err)
	}

	// Click inside the portal.
	ev := tcell.NewEventMouse(2, 2, tcell.Button1, tcell.ModNone)
	app.handleMouseEvent(ev, root, layout)

	if dismissed {
		t.Error("expected OnOutsideClick NOT to be called on click inside portal")
	}
}

// ---------------------------------------------------------------------------
// PopupMode → EventContext.PopupOpen
// ---------------------------------------------------------------------------

func TestPopupModeSetInEventContext(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	app := NewAppWithScreen(screen)

	var capturedPopupOpen bool
	input := &testPopupObserver{id: "inp", onEvent: func(ctx EventContext) { capturedPopupOpen = ctx.PopupOpen }}

	portal := &Portal{
		Child:     NewBox(Style{Width: 10, Height: 5}, NewText(Style{}, "")),
		X:         10, Y: 10,
		PopupMode: true,
	}

	root := NewBox(Style{}, input, portal)

	app.focusedID = "inp"
	_, _, layout, focusableIDs, err := app.RenderFrame(func(_ *RenderContext) Node { return root })
	if err != nil {
		t.Fatal(err)
	}

	ev := tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)
	app.handleKeyEvent(ev, root, layout, focusableIDs)

	if !capturedPopupOpen {
		t.Error("expected EventContext.PopupOpen=true when a PopupMode portal is present")
	}
}

// testPopupObserver is a minimal focusable EventHandler that records the
// EventContext it receives, used to inspect PopupOpen.
type testPopupObserver struct {
	id      string
	onEvent func(EventContext)
}

func (n *testPopupObserver) GetStyle() Style       { return Style{ID: n.id, Focusable: true} }
func (n *testPopupObserver) IsFocusable() bool      { return true }
func (n *testPopupObserver) DefaultState() any       { return nil }
func (n *testPopupObserver) HandleEvent(_ tcell.Event, _ any, ctx EventContext) bool {
	if n.onEvent != nil {
		n.onEvent(ctx)
	}
	return false
}

// ---------------------------------------------------------------------------
// Dismissable / dismissOthers
// ---------------------------------------------------------------------------

func TestDismissOthers_ClosesOpenDropdown(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	app := NewAppWithScreen(screen)
	app.componentStates["dd"] = &DropdownState{Open: true}
	app.componentStates["btn"] = nil

	ctx := &RenderContext{app: app}
	dd := NewDropdown(ctx, Style{ID: "dd", Focusable: true}, []string{"A"}, 0, nil)
	btn := NewButton(Style{ID: "btn", Focusable: true}, "OK", nil)
	root := NewBox(Style{}, dd, btn)

	app.focusedID = "dd"
	app.setFocus("btn", root)

	state := app.componentStates["dd"].(*DropdownState)
	if state.Open {
		t.Error("expected dropdown to be dismissed after focus moved to btn")
	}
}

func TestDismissOthers_KeepsFocusedComponentOpen(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	_ = screen.Init()
	defer screen.Fini()
	screen.SetSize(80, 24)

	app := NewAppWithScreen(screen)
	app.componentStates["dd"] = &DropdownState{Open: true}

	ctx := &RenderContext{app: app}
	dd := NewDropdown(ctx, Style{ID: "dd", Focusable: true}, []string{"A"}, 0, nil)
	root := NewBox(Style{}, dd)

	// Focus the dropdown itself — it should NOT be dismissed.
	app.focusedID = ""
	app.setFocus("dd", root)

	state := app.componentStates["dd"].(*DropdownState)
	if !state.Open {
		t.Error("expected focused dropdown to remain open")
	}
}
