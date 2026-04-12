package tz

import "github.com/gdamore/tcell/v2"

// NewTestApp creates an App backed by a simulation screen sized w×h.
// Use with Step for synchronous, step-through testing without a blocking event loop.
func NewTestApp(w, h int) (*App, error) {
	s := tcell.NewSimulationScreen("")
	if err := s.Init(); err != nil {
		return nil, err
	}
	s.SetSize(w, h)
	return NewAppWithScreen(s), nil
}

// Step renders one frame and, if ev is non-nil, dispatches it synchronously.
// Returns the rendered Grid so callers can inspect visual output.
//
// A typical test loop looks like:
//
//	app, _ := NewTestApp(80, 24)
//	app.Step(renderFn, nil)                                          // initial render
//	app.Step(renderFn, tcell.NewEventKey(tcell.KeyEnter, 0, 0))     // dispatch Enter
//	app.Step(renderFn, nil)                                          // re-render after state change
func (a *App) Step(renderFn func(*RenderContext) Node, ev tcell.Event) (*Grid, error) {
	grid, root, layout, focusableIDs, err := a.RenderFrame(renderFn)
	if err != nil {
		return nil, err
	}
	if ev != nil {
		switch ev := ev.(type) {
		case *tcell.EventKey:
			a.handleKeyEvent(ev, root, layout, focusableIDs)
		case *tcell.EventMouse:
			a.handleMouseEvent(ev, root, layout)
		}
	}
	return grid, nil
}
