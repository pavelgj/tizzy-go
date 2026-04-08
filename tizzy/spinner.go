package tizzy

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Spinner is a node that displays a loading animation.
type Spinner struct {
	Style    Style
	Frames   []string
	Interval time.Duration
}

// NewSpinner creates a new Spinner node.
func NewSpinner(ctx *RenderContext, style Style) *Spinner {
	hookId := fmt.Sprintf("hook-%d", ctx.hookIndex)
	_, setFrameIdx := UseState(ctx, 0)

	if style.ID == "" {
		style.ID = hookId
	}

	s := &Spinner{
		Style:    style,
		Frames:   []string{"|", "/", "-", "\\"},
		Interval: 100 * time.Millisecond,
	}

	ctx.UseEffect(func() func() {
		ticker := time.NewTicker(s.Interval)
		done := make(chan bool)
		go func() {
			for {
				select {
				case <-ticker.C:
					currentVal := 0
					if stateObj, ok := ctx.app.componentStates[style.ID]; ok {
						currentVal = stateObj.(int)
					}
					setFrameIdx((currentVal + 1) % len(s.Frames))
				case <-done:
					ticker.Stop()
					return
				}
			}
		}()
		return func() {
			done <- true
		}
	})

	return s
}


// Layout calculates the layout for the Spinner node.
func (n *Spinner) Layout(x, y int, c Constraints) LayoutResult {
	pad := n.Style.Padding
	margin := n.Style.Margin
	boxX := x + margin.Left
	boxY := y + margin.Top

	w := 1
	h := 1

	if n.Style.Width > 0 {
		w = n.Style.Width
	}

	borderSize := 0
	if n.Style.Border {
		borderSize = 2
	}

	layoutH := h + pad.Top + pad.Bottom + borderSize
	if n.Style.MaxHeight > 0 && layoutH > n.Style.MaxHeight {
		layoutH = n.Style.MaxHeight
	}

	return LayoutResult{
		Node: n,
		X:    boxX,
		Y:    boxY,
		W:    w + pad.Left + pad.Right + borderSize,
		H:    layoutH,
	}
}

// Render draws the Spinner node to the grid.
func (n *Spinner) Render(grid *Grid, layout LayoutResult, focusedID string, componentStates map[string]any) {
	style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)

	borderOffset := 0
	borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
	if n.Style.Border {
		borderOffset = 1
		drawBorder(grid, layout.X, layout.Y, layout.W, layout.H, "", borderStyle)
	}

	frameIdx := 0
	if stateObj, ok := componentStates[n.Style.ID]; ok {
		frameIdx = stateObj.(int)
	}
	if frameIdx >= len(n.Frames) {
		frameIdx = 0
	}
	val := n.Frames[frameIdx]

	drawText(grid, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, val, style)
}

// GetStyle returns the style of the Spinner node.
func (n *Spinner) GetStyle() Style {
	return n.Style
}
