package main

import (
	"fmt"
	"log"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	percent := 0.0
	percent2 := 0.0

	err = app.Run(
		func(ctx *tz.RenderContext) tz.Node {
			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Progress Bar Sample"),
				tz.NewText(tz.Style{}, "Animates automatically using the tick event."),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					tz.NewProgressBar(tz.Style{Width: 30, Color: tcell.ColorGreen}, percent),
					tz.NewText(tz.Style{Margin: tz.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent*100))),
				),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					tz.NewButton(tz.Style{ID: "increase", Focusable: true, Border: true}, "Increase", func() {
						percent2 += 0.1
						if percent2 > 1.0 {
							percent2 = 0.0
						}
					}),
					tz.NewButton(tz.Style{ID: "decrease", Focusable: true, Border: true}, "Decrease", func() {
						percent2 -= 0.1
						if percent2 < 0.0 {
							percent2 = 1.0
						}
					}),
				),
				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					tz.NewProgressBar(tz.Style{Width: 100, Color: tcell.ColorBlue}, percent2),
					tz.NewText(tz.Style{Margin: tz.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent2*100))),
				),
			)
		},
		func(ev tcell.Event) {
			switch ev.(type) {
			case *tz.EventTick:
				percent += 0.02
				if percent > 1.0 {
					percent = 0.0
				}
				app.MarkDirty()
			}
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
