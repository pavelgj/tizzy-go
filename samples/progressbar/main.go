package main

import (
	"fmt"
	"log"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	percent := 0.0
	percent2 := 0.0

	err = app.Run(
		func(ctx *tizzy.RenderContext) tizzy.Node {
			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Progress Bar Sample"),
				tizzy.NewText(tizzy.Style{}, "Animates automatically using the tick event."),

				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					tizzy.NewProgressBar(tizzy.Style{Width: 30, Color: tcell.ColorGreen}, percent),
					tizzy.NewText(tizzy.Style{Margin: tizzy.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent*100))),
				),

				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					tizzy.NewButton(tizzy.Style{ID: "increase", Focusable: true, Border: true}, "Increase", func() {
						percent2 += 0.1
						if percent2 > 1.0 {
							percent2 = 0.0
						}
					}),
					tizzy.NewButton(tizzy.Style{ID: "decrease", Focusable: true, Border: true}, "Decrease", func() {
						percent2 -= 0.1
						if percent2 < 0.0 {
							percent2 = 1.0
						}
					}),
				),
				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					tizzy.NewProgressBar(tizzy.Style{Width: 100, Color: tcell.ColorBlue}, percent2),
					tizzy.NewText(tizzy.Style{Margin: tizzy.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent2*100))),
				),
			)
		},
		func(ev tcell.Event) {
			switch ev.(type) {
			case *tizzy.EventTick:
				percent += 0.02
				if percent > 1.0 {
					percent = 0.0
				}
			}
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
