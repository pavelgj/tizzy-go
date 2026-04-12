package tz

import (
	"testing"
)

func TestUseState(t *testing.T) {
	app, err := NewTestApp(40, 10)
	if err != nil {
		t.Fatal(err)
	}

	var capturedCount int
	var setCount func(any)

	renderFn := func(ctx *RenderContext) Node {
		v, set := ctx.UseState(0)
		capturedCount = v.(int)
		setCount = set
		return NewText(Style{ID: "t1"}, "Hi")
	}

	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if capturedCount != 0 {
		t.Errorf("Expected initial count 0, got %d", capturedCount)
	}

	setCount(1)

	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if capturedCount != 1 {
		t.Errorf("Expected count 1 after update, got %d", capturedCount)
	}
}

func TestUseEffect(t *testing.T) {
	app, err := NewTestApp(40, 10)
	if err != nil {
		t.Fatal(err)
	}

	mounted := false
	unmounted := false
	var setState func(any)

	renderFn := func(ctx *RenderContext) Node {
		stateObj, set := ctx.UseState(0)
		setState = set

		if stateObj.(int) == 0 {
			ctx.UseEffect(func() func() {
				mounted = true
				return func() { unmounted = true }
			})
			return NewText(Style{ID: "effect1"}, "I exist")
		}
		return NewText(Style{ID: "other"}, "I don't exist")
	}

	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if !mounted {
		t.Error("Expected effect to be mounted after first render")
	}

	setState(1)

	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if !unmounted {
		t.Error("Expected effect cleanup to run when effect leaves the tree")
	}
}

func TestRenderOptimization(t *testing.T) {
	app, err := NewTestApp(40, 10)
	if err != nil {
		t.Fatal(err)
	}

	renderCount := 0
	var setState func(any)

	renderFn := func(ctx *RenderContext) Node {
		renderCount++
		_, set := ctx.UseState(0)
		setState = set
		return NewText(Style{ID: "t1"}, "Hi")
	}

	// Initial render.
	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if renderCount != 1 {
		t.Errorf("Expected 1 render after init, got %d", renderCount)
	}

	// State change triggers exactly one more render.
	setState(1)
	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if renderCount != 2 {
		t.Errorf("Expected 2 renders after state change, got %d", renderCount)
	}

	// No state change — render function must NOT be called again.
	if _, err := app.Step(renderFn, nil); err != nil {
		t.Fatal(err)
	}
	if renderCount != 2 {
		t.Errorf("Expected render count to stay at 2, got %d", renderCount)
	}
}
