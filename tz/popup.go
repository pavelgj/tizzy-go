package tz

import (
	"fmt"
)

// NewPopup creates a portal positioned at an explicit absolute screen position.
// When isOpen is false, nil is returned; NewBox silently drops nil children.
//
// PopupMode is set so that underlying TextInput components know to suppress
// their navigation-key handling (Enter / Up / Down) while the popup is visible.
//
// One hook slot is always consumed regardless of isOpen so that callers keep
// stable hook indices across renders.
func NewPopup(ctx *RenderContext, style Style, child Node, x, y int, isOpen bool) Node {
	// Consume one hook slot for stable hook ordering across renders.
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	if !isOpen {
		return nil
	}

	if style.ID == "" {
		style.ID = id
	}

	// Wrap the child with the popup's visual style (border, background, etc.)
	// so the caller does not have to duplicate those style fields on their child.
	wrapperStyle := Style{
		Border:     style.Border,
		Width:      style.Width,
		Background: style.Background,
		Color:      style.Color,
		Padding:    style.Padding,
	}
	wrapped := NewBox(wrapperStyle, child)

	return &Portal{
		Style:     style,
		Child:     wrapped,
		X:         x,
		Y:         y,
		PopupMode: true,
	}
}
