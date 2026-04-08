package splotch

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
		countObj, setCount := ctx.UseState("counter", 0)
		count := countObj.(int)

		if renderCount == 1 {
			if count != 0 {
				t.Errorf("Expected initial count 0, got %d", count)
			}
			setCount(1)
		} else if renderCount == 2 {
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

	renderCount := 0
	err := app.Run(func(ctx *RenderContext) Node {
		renderCount++
		
		ctx.UseEffect("effect1", func() func() {
			mounted = true
			return func() {
				unmounted = true
			}
		})
		
		if renderCount == 1 {
			return NewText(Style{ID: "effect1"}, "I exist")
		} else {
			return NewText(Style{ID: "other"}, "I don't exist")
		}
	}, func(ev tcell.Event) {})

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
