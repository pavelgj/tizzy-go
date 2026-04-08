package tizzy

import (
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestModalLayout(t *testing.T) {
	ctx := &RenderContext{app: &App{componentStates: make(map[string]any)}}
	modal := NewModal(
		ctx,
		Style{ID: "modal"},
		NewText(Style{Color: tcell.ColorWhite}, "Content"),
		true,
	)

	res := Layout(modal, 10, 20, Constraints{MaxW: 100, MaxH: 100})

	if res.X != 10 || res.Y != 20 {
		t.Errorf("Expected position (10, 20), got (%d, %d)", res.X, res.Y)
	}
	if res.W != 0 || res.H != 0 {
		t.Errorf("Expected size (0, 0) for Modal in tree, got (%d, %d)", res.W, res.H)
	}
}

func TestModalFocusTrapping(t *testing.T) {
	ctx := &RenderContext{app: &App{componentStates: make(map[string]any)}}
	modal := NewModal(
		ctx,
		Style{ID: "modal"},
		NewBox(
			Style{},
			NewButton(Style{ID: "btn1", Focusable: true}, "Btn1", nil),
			NewButton(Style{ID: "btn2", Focusable: true}, "Btn2", nil),
		),
		true,
	)

	// Simulate App.Run behavior when modal is open:
	// It calls findFocusableIDs on activeModal.Child!
	ids := findFocusableIDs(modal.Child, nil)

	if len(ids) != 2 {
		t.Fatalf("Expected 2 focusable IDs, got %d", len(ids))
	}
	if ids[0] != "btn1" || ids[1] != "btn2" {
		t.Errorf("Expected [btn1, btn2], got %v", ids)
	}
}
