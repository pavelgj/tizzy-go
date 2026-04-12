package tz

import (
	"fmt"
)

// NewModal creates a centered overlay portal with focus trapping.
// When isOpen is false, nil is returned; NewBox silently drops nil children,
// so callers can always pass NewModal into a box without an if-guard.
//
// The caller controls open/close via normal UseState — the modal is simply
// absent from the tree when closed.  No internal ModalState is stored.
//
// One hook slot is always consumed regardless of isOpen so that callers who
// place NewModal alongside other hook-using constructors keep stable indices.
func NewModal(ctx *RenderContext, style Style, child Node, isOpen bool) Node {
	// Consume one hook slot for stable hook ordering across renders.
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	if !isOpen {
		return nil
	}

	if style.ID == "" {
		style.ID = id
	}

	// Wrap the caller-supplied child in a styled border box so the modal has
	// its own background and border drawn from style.
	wrapper := NewBox(Style{
		Border:          true,
		Color:           style.Color,
		Background:      style.Background,
		FocusColor:      style.FocusColor,
		FocusBackground: style.FocusBackground,
	}, child)

	return &Portal{
		Style:     style,
		Child:     wrapper,
		X:         -1, // auto-center
		Y:         -1,
		TrapFocus: true,
	}
}
