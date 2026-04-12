package tz

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"reflect"

	"github.com/gdamore/tcell/v2"
)

// DumpToPNG saves the grid as a PNG image.
// Each cell is rendered as a block of cellW x cellH pixels.
func (g *Grid) DumpToPNG(filename string, cellW, cellH int) error {
	img := image.NewRGBA(image.Rect(0, 0, g.W*cellW, g.H*cellH))

	for y := 0; y < g.H; y++ {
		for x := 0; x < g.W; x++ {
			cell := g.Cells[y][x]
			v := reflect.ValueOf(cell.Style)
			if v.Kind() == reflect.Interface {
				v = v.Elem()
			}
			
			var fgVal, bgVal uint64
			if v.Kind() == reflect.Struct && v.NumField() >= 2 {
				fgVal = v.Field(0).Uint()
				bgVal = v.Field(1).Uint()
			}
			
			bgColor := tcellToColor(tcell.Color(bgVal))
			fgColor := tcellToColor(tcell.Color(fgVal))

			// Fill background
			for cy := 0; cy < cellH; cy++ {
				for cx := 0; cx < cellW; cx++ {
					img.Set(x*cellW+cx, y*cellH+cy, bgColor)
				}
			}

			// Render rune using bitmap font if it's ASCII
			if cell.Rune > 0 && cell.Rune < 128 {
				bitmap := Font8x8Basic[cell.Rune]
				for cy := 0; cy < 8; cy++ {
					b := bitmap[cy]
					for cx := 0; cx < 8; cx++ {
						if (b & (1 << cx)) != 0 {
							sw := cellW / 8
							sh := cellH / 8
							if sw == 0 { sw = 1 }
							if sh == 0 { sh = 1 }
							for sy := 0; sy < sh; sy++ {
								for sx := 0; sx < sw; sx++ {
									img.Set(x*cellW+cx*sw+sx, y*cellH+cy*sh+sy, fgColor)
								}
							}
						}
					}
				}
			} else if cell.Rune != ' ' {
				// Handle box drawing characters procedurally
				switch cell.Rune {
				case '─':
					cy := cellH / 2
					for cx := 0; cx < cellW; cx++ {
						img.Set(x*cellW+cx, y*cellH+cy, fgColor)
					}
				case '│':
					cx := cellW / 2
					for cy := 0; cy < cellH; cy++ {
						img.Set(x*cellW+cx, y*cellH+cy, fgColor)
					}
				case '┌':
					cx := cellW / 2
					cy := cellH / 2
					for x2 := cx; x2 < cellW; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
					for y2 := cy; y2 < cellH; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
				case '┐':
					cx := cellW / 2
					cy := cellH / 2
					for x2 := 0; x2 <= cx; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
					for y2 := cy; y2 < cellH; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
				case '└':
					cx := cellW / 2
					cy := cellH / 2
					for y2 := 0; y2 <= cy; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
					for x2 := cx; x2 < cellW; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
				case '┘':
					cx := cellW / 2
					cy := cellH / 2
					for y2 := 0; y2 <= cy; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
					for x2 := 0; x2 <= cx; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
				// Rounded corner variants (╭╮╯╰) — rendered identically to square corners
				case '╭':
					cx := cellW / 2
					cy := cellH / 2
					for x2 := cx; x2 < cellW; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
					for y2 := cy; y2 < cellH; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
				case '╮':
					cx := cellW / 2
					cy := cellH / 2
					for x2 := 0; x2 <= cx; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
					for y2 := cy; y2 < cellH; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
				case '╯':
					cx := cellW / 2
					cy := cellH / 2
					for y2 := 0; y2 <= cy; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
					for x2 := 0; x2 <= cx; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
				case '╰':
					cx := cellW / 2
					cy := cellH / 2
					for y2 := 0; y2 <= cy; y2++ { img.Set(x*cellW+cx, y*cellH+y2, fgColor) }
					for x2 := cx; x2 < cellW; x2++ { img.Set(x*cellW+x2, y*cellH+cy, fgColor) }
				case '▼':
					// Filled triangle pointing down (widest at top, narrows to point at bottom)
					for cy := 0; cy < cellH; cy++ {
						halfW := (cellW/2) * (cellH - 1 - cy) / (cellH - 1)
						midX := cellW / 2
						for cx := midX - halfW; cx <= midX+halfW; cx++ {
							img.Set(x*cellW+cx, y*cellH+cy, fgColor)
						}
					}
				case '▲':
					// Filled triangle pointing up (point at top, widest at bottom)
					for cy := 0; cy < cellH; cy++ {
						halfW := (cellW / 2) * cy / (cellH - 1)
						midX := cellW / 2
						for cx := midX - halfW; cx <= midX+halfW; cx++ {
							img.Set(x*cellW+cx, y*cellH+cy, fgColor)
						}
					}
				case '✓':
					// Checkmark: short left leg down-right to pivot, long right leg up-right
					pivot := cellW / 3
					drawSeg := func(x0, y0, x1, y1 int) {
						dx, dy := x1-x0, y1-y0
						steps := dx
						if dy > steps { steps = dy }
						if dy < -steps { steps = -dy }
						if dx < -steps { steps = -dx }
						if steps == 0 {
							img.Set(x*cellW+x0, y*cellH+y0, fgColor)
							return
						}
						for i := 0; i <= steps; i++ {
							px := x0 + dx*i/steps
							py := y0 + dy*i/steps
							img.Set(x*cellW+px, y*cellH+py, fgColor)
							if px+1 < cellW {
								img.Set(x*cellW+px+1, y*cellH+py, fgColor)
							}
						}
					}
					drawSeg(0, cellH/2, pivot, cellH-1)
					drawSeg(pivot, cellH-1, cellW-1, 1)
				default:
					// Fallback for other characters (fill whole cell)
					for cy := 0; cy < cellH; cy++ {
						for cx := 0; cx < cellW; cx++ {
							img.Set(x*cellW+cx, y*cellH+cy, fgColor)
						}
					}
				}
			}
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func tcellToColor(c tcell.Color) color.Color {
	if c == tcell.ColorDefault || c == tcell.ColorReset {
		return color.RGBA{0, 0, 0, 255} // Default to black
	}
	r, g, b := c.RGB()
	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}
