// samples/ansitext demonstrates tz.NewANSIText — a node that renders strings
// containing ANSI SGR escape sequences inside a Tizzy layout tree.
//
// Run: go run ./samples/ansitext
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/pavelgj/tizzy-go/tz"
)

// esc builds an ANSI SGR sequence: ESC [ <params> m
func esc(params string) string { return "\x1b[" + params + "m" }
func reset() string            { return esc("0") }

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		// Title and hint live outside the scroll view so they stay fixed.
		title := tz.NewANSIText(tz.Style{},
			esc("1;38;2;180;100;255")+"ANSIText"+reset()+" "+esc("2")+"— ANSI escape sequences inside Tizzy layouts"+reset())
		hint := tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
			"↑/↓ or mouse wheel to scroll  •  Ctrl+C to exit")

		// All demo content goes inside a scrollable column.
		scrollContent := tz.NewBox(
			tz.Style{FlexDirection: "column"},

			separator(),

			// ── Basic 8 foreground colors ──────────────────────────────
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Basic foreground (30–37 / 90–97)"),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, basicColorRow(false)),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, basicColorRow(true)),

			// ── 256-color palette row ───────────────────────────────────
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "256-color palette (38;5;n)"),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, palette256Row(16, 16)),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, palette256Row(52, 16)),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, palette256Row(196, 16)),

			// ── True-color gradient ────────────────────────────────────
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "True-color gradient (38;2;r;g;b)"),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, trueColorGradient(48)),

			// ── Text attributes ────────────────────────────────────────
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Text attributes"),
			tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, attrSamples()),

			// ── Mixed inline styles ────────────────────────────────────
			tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Inline mixed styles inside a Tizzy box"),
			tz.NewBox(
				tz.Style{
					Border:        true,
					Padding:       tz.Padding{Left: 2, Right: 2, Top: 1, Bottom: 1},
					FlexDirection: "column",
					Margin:        tz.Margin{Top: 1},
				},
				tz.NewANSIText(tz.Style{}, mixedParagraph()),
				tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
					"← this box is a plain tz.Box; only the text inside uses ANSIText"),
			),

			separator(),
		)

		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Left: 3, Right: 3, Bottom: 1},
				FillWidth:     true,
				FillHeight:    true,
			},
			title,
			tz.NewScrollView(ctx, tz.Style{
				ID:         "ansitext-scroll",
				Focusable:  true,
				FillWidth:  true,
				FillHeight: true,
				Margin:     tz.Margin{Top: 1},
			}, scrollContent),
			hint,
		)
	}

	if err := app.Run(render, nil); err != nil {
		log.Fatal(err)
	}
}

// separator returns a thin horizontal rule as an ANSIText node.
func separator() tz.Node {
	return tz.NewANSIText(tz.Style{Margin: tz.Margin{Top: 1}}, esc("2")+strings.Repeat("─", 60)+reset())
}

// basicColorRow builds a row of colored "██" swatches for the 8 basic ANSI
// foreground colors. bright=true uses the 90–97 high-intensity codes.
func basicColorRow(bright bool) string {
	names := []string{"black", "red", "green", "yellow", "blue", "magenta", "cyan", "white"}
	var b strings.Builder
	for i, name := range names {
		code := 30 + i
		if bright {
			code = 90 + i
		}
		b.WriteString(fmt.Sprintf("%s%-9s%s", esc(fmt.Sprint(code)), name, reset()))
	}
	return b.String()
}

// palette256Row builds a row of 256-color index swatches starting at `start`.
func palette256Row(start, count int) string {
	var b strings.Builder
	for i := 0; i < count && start+i <= 255; i++ {
		idx := start + i
		b.WriteString(fmt.Sprintf("%s%s%s%3d%s",
			esc(fmt.Sprintf("48;5;%d", idx)),
			esc("38;5;0"),
			" ",
			idx,
			reset()+" "),
		)
	}
	return b.String()
}

// trueColorGradient builds a magenta→cyan gradient using true-color sequences.
func trueColorGradient(steps int) string {
	var b strings.Builder
	for i := 0; i < steps; i++ {
		t := float64(i) / float64(steps-1)
		r := int(180 * (1 - t))
		g := int(80 + 80*t)
		bv := int(100 + 155*t)
		b.WriteString(fmt.Sprintf("%s██%s", esc(fmt.Sprintf("38;2;%d;%d;%d", r, g, bv)), reset()))
	}
	return b.String()
}

// attrSamples builds a string showcasing each supported SGR text attribute.
func attrSamples() string {
	samples := []struct{ code, label string }{
		{"1", "bold"},
		{"2", "dim"},
		{"3", "italic"},
		{"4", "underline"},
		{"9", "strikethrough"},
		{"7", "reverse"},
		{"5", "blink"},
	}
	var b strings.Builder
	for _, s := range samples {
		b.WriteString(fmt.Sprintf("%s%-15s%s  ", esc(s.code), s.label, reset()))
	}
	return b.String()
}

// mixedParagraph demonstrates multiple styles spanning a single string.
func mixedParagraph() string {
	return esc("1;38;2;255;200;50") + "Warning: " + reset() +
		"The quick " +
		esc("3") + "brown" + reset() +
		" fox " +
		esc("1") + "jumps" + reset() +
		" over the " +
		esc("38;2;80;180;255") + "lazy " + esc("4") + "dog" + reset() +
		".\n" +
		esc("2") + "Styles reset cleanly across lines — each row re-applies from scratch." + reset()
}
