package tz

import "fmt"

// EffectRecord stores a lifecycle effect and its stable hook ID.
type EffectRecord struct {
	ID     string
	Effect func() func()
}

// RenderContext provides hooks and state scoping during the render function.
// It is created fresh each frame and passed to the root render function.
type RenderContext struct {
	app       *App
	effects   []EffectRecord
	hookIndex int
}

// GetFocusedID returns the ID of the currently focused component.
func (ctx *RenderContext) GetFocusedID() string {
	return ctx.app.GetFocusedID()
}

// UseState retrieves or initializes per-component state, keyed by call order.
// Returns the current value and a setter that triggers a re-render.
func (ctx *RenderContext) UseState(initial any) (any, func(any)) {
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	state := ctx.app.componentStates[id]
	if state == nil {
		state = initial
		ctx.app.componentStates[id] = state
	}
	setter := func(newVal any) {
		ctx.app.mu.Lock()
		ctx.app.componentStates[id] = newVal
		ctx.app.mu.Unlock()
		ctx.app.MarkDirty()
	}
	return state, setter
}

// UseState is a type-safe wrapper around RenderContext.UseState.
func UseState[T any](ctx *RenderContext, initial T) (T, func(T)) {
	stateObj, setter := ctx.UseState(initial)
	return stateObj.(T), func(newVal T) {
		setter(newVal)
	}
}

// UseEffect registers a lifecycle effect. The effect function runs on mount
// and its return value (if non-nil) is called on unmount.
func (ctx *RenderContext) UseEffect(effect func() func()) {
	id := fmt.Sprintf("hook-%d", ctx.hookIndex)
	ctx.hookIndex++

	ctx.effects = append(ctx.effects, EffectRecord{ID: id, Effect: effect})
}
