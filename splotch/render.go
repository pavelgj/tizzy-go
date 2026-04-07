package splotch

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Render draws the layout tree to the screen.
func Render(screen tcell.Screen, layout LayoutResult, focusedID string, componentStates map[string]any) {
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
		
		borderOffset := 0
		if n.Style.Border {
			borderOffset = 1
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}

		if n.Style.Multiline {
			vScrollOffset := 0
			if n.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[n.Style.ID]; ok {
					state := stateObj.(*TextInputState)
					vScrollOffset = state.vScrollOffset
				}
			}

			lines := strings.Split(n.Value, "\n")
			w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
			visibleHeight := layout.H - n.Style.Padding.Top - n.Style.Padding.Bottom - borderOffset*2
			
			for i := 0; i < visibleHeight; i++ {
				lineIdx := i + vScrollOffset
				if lineIdx >= len(lines) {
					break
				}
				val := lines[lineIdx]
				if len(val) > w && w > 0 {
					val = val[:w]
				}
				drawText(screen, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset+i, val, style)
			}
		} else {
			val := n.Value
			scrollOffset := 0
			if n.Style.ID != "" && componentStates != nil {
				if stateObj, ok := componentStates[n.Style.ID]; ok {
					state := stateObj.(*TextInputState)
					scrollOffset = state.scrollOffset
				}
			}
			
			w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
			if scrollOffset < len(val) {
				val = val[scrollOffset:]
			} else {
				val = ""
			}
			if len(val) > w && w > 0 {
				val = val[:w]
			}
			
			drawText(screen, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, val, style)
		}
	case *Button:
		focused := false
		if n.Style.ID != "" && n.Style.ID == focusedID {
			focused = true
		}
		
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		if focused {
			style = tcell.StyleDefault.Foreground(tcell.ColorBlack).Background(tcell.ColorYellow)
		}
		
		borderOffset := 0
		borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		if n.Style.Border {
			borderOffset = 1
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}
		
		label := "[ " + n.Label + " ]"
		drawText(screen, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, label, style)
	case *Spinner:
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		
		borderOffset := 0
		borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		if n.Style.Border {
			borderOffset = 1
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}
		
		now := time.Now()
		frameIdx := int((now.UnixNano() / int64(n.Interval)) % int64(len(n.Frames)))
		val := n.Frames[frameIdx]
		
		drawText(screen, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, val, style)
	case *ProgressBar:
		style := tcell.StyleDefault.Foreground(n.Style.Color).Background(n.Style.Background)
		
		borderOffset := 0
		borderStyle := tcell.StyleDefault.Foreground(tcell.ColorYellow)
		if n.Style.Border {
			borderOffset = 1
			drawBorder(screen, layout.X, layout.Y, layout.W, layout.H, borderStyle)
		}
		
		w := layout.W - n.Style.Padding.Left - n.Style.Padding.Right - borderOffset*2
		if w < 0 {
			w = 0
		}
		
		filledW := int(float64(w) * n.Percent)
		if filledW > w {
			filledW = w
		}
		
		str := ""
		for i := 0; i < filledW; i++ {
			str += n.FilledChar
		}
		for i := filledW; i < w; i++ {
			str += n.EmptyChar
		}
		
		drawText(screen, layout.X+n.Style.Padding.Left+borderOffset, layout.Y+n.Style.Padding.Top+borderOffset, str, style)
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
			Render(screen, child, focusedID, componentStates)
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
	col := 0
	for _, r := range text {
		screen.SetContent(x+col, y, r, nil, style)
		col++
	}
}
