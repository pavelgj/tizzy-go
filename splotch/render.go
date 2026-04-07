package splotch

import (
	"github.com/gdamore/tcell/v2"
)

// Render draws the layout tree to the screen.
func Render(screen tcell.Screen, layout LayoutResult, focusedID string) {
	focused := false
	switch n := layout.Node.(type) {
	case *Text:
		if n.Style.ID != "" && n.Style.ID == focusedID {
			focused = true
		}
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		if focused {
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}
		drawText(screen, layout.X+n.Style.Padding.Left, layout.Y+n.Style.Padding.Top, n.Content, style)
	case *TextInput:
		if n.Style.ID != "" && n.Style.ID == focusedID {
			focused = true
		}
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		borderStyle := style
		if focused {
			borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}
		if n.Style.Border {
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}
		drawText(screen, layout.X+n.Style.Padding.Left, layout.Y+n.Style.Padding.Top, n.Value, style)
	case *Box:
		if n.Style.ID != "" && n.Style.ID == focusedID {
			focused = true
		}
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		borderStyle := style
		if focused {
			borderStyle = tcell.StyleDefault.Foreground(tcell.ColorYellow).Background(n.Style.Background)
		}
		if n.Style.Border {
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}
		for _, child := range layout.Children {
			Render(screen, child, focusedID)
		}
	}
}

func drawBorder(screen tcell.Screen, x, y, w, h int, style tcell.Style) {
	// Top and bottom borders
	for i := 0; i < w; i++ {
		screen.SetContent(x+i, y, '─', nil, style)
		screen.SetContent(x+i, y+h-1, '─', nil, style)
	}

	// Left and right borders
	for i := 0; i < h; i++ {
		screen.SetContent(x, y+i, '│', nil, style)
		screen.SetContent(x+w-1, y+i, '│', nil, style)
	}

	// Corners
	screen.SetContent(x, y, '┌', nil, style)
	screen.SetContent(x+w-1, y, '┐', nil, style)
	screen.SetContent(x, y+h-1, '└', nil, style)
	screen.SetContent(x+w-1, y+h-1, '┘', nil, style)
}

func drawText(screen tcell.Screen, x, y int, text string, style tcell.Style) {
	for i, r := range text {
		screen.SetContent(x+i, y, r, nil, style)
	}
}
