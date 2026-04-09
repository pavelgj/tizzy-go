package tz

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

func TestUseState(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	app := &App{
		screen:          s,
		componentStates: make(map[string]any),
		activeCleanups:  make(map[string]func()),
	}

	go func() {
		time.Sleep(200 * time.Millisecond)
		s.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
		time.Sleep(100 * time.Millisecond)
		s.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	}()

	renderCount := 0
	err := app.Run(func(ctx *RenderContext) Node {
		renderCount++
		countObj, setCount := ctx.UseState(0)
		count := countObj.(int)

		switch renderCount {
		case 1:
			if count != 0 {
				t.Errorf("Expected initial count 0, got %d", count)
			}
			setCount(1)
		case 2:
			if count != 1 {
				t.Errorf("Expected count 1 on second render, got %d", count)
			}
		}
		return NewText(Style{ID: "t1"}, "Hi")
	}, func(ev tcell.Event) {})

	if err != nil {
		t.Fatal(err)
	}
}

func TestUseEffect(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	app := &App{
		screen:          s,
		componentStates: make(map[string]any),
		activeCleanups:  make(map[string]func()),
	}

	go func() {
		time.Sleep(200 * time.Millisecond)
		s.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
		time.Sleep(100 * time.Millisecond)
		s.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	}()

	mounted := false
	unmounted := false

	var setState func(any)
	err := app.Run(func(ctx *RenderContext) Node {
		stateObj, set := ctx.UseState(0)
		setState = set
		state := stateObj.(int)

		if state == 0 {

			ctx.UseEffect(func() func() {
				mounted = true
				return func() {
					unmounted = true
				}
			})

			return NewText(Style{ID: "effect1"}, "I exist")
		} else {
			return NewText(Style{ID: "other"}, "I don't exist")
		}
	}, func(ev tcell.Event) {
		if _, ok := ev.(*tcell.EventKey); ok {
			setState(1)
		}
	})

	if err != nil {
		t.Fatal(err)
	}

	if !mounted {
		t.Errorf("Expected effect to be called (mounted)")
	}
	if !unmounted {
		t.Errorf("Expected cleanup to be called (unmounted)")
	}
}

func TestRenderOptimization(t *testing.T) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		t.Fatal(err)
	}
	app := &App{
		screen:          s,
		componentStates: make(map[string]any),
		activeCleanups:  make(map[string]func()),
	}

	go func() {
		time.Sleep(200 * time.Millisecond)
		s.InjectKey(tcell.KeyRune, 'a', tcell.ModNone)
		time.Sleep(100 * time.Millisecond)
		s.InjectKey(tcell.KeyEscape, 0, tcell.ModNone)
	}()

	ctxRenderCount := 0
	var currentState any
	var setState func(any)

	err := app.Run(func(ctx *RenderContext) Node {
		ctxRenderCount++

		stateObj, set := ctx.UseState(0)
		currentState = stateObj
		setState = set

		return NewText(Style{ID: "t1"}, "Hi")
	}, func(ev tcell.Event) {
		if _, ok := ev.(*tcell.EventKey); ok {
			if currentState.(int) == 0 {
				setState(1)
			}
		}
	})

	if err != nil {
		t.Fatal(err)
	}

	// Initial render (1) + state change re-render (2).
	// Subsequent ticks and keys should NOT trigger render.
	if ctxRenderCount != 2 {
		t.Errorf("Expected renderFn to be called exactly 2 times, got %d", ctxRenderCount)
	}
}
