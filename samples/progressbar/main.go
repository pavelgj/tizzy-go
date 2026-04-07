package main

import (
	"fmt"
	"log"
	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	percent := 0.0
	percent2 := 0.0

	err = app.Run(
		func() splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Progress Bar Sample"),
				splotch.NewText(splotch.Style{}, "Animates automatically using the tick event."),

				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					splotch.NewProgressBar(splotch.Style{Width: 30, Color: tcell.ColorGreen}, percent),
					splotch.NewText(splotch.Style{Margin: splotch.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent*100))),
				),

				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					splotch.NewButton(splotch.Style{ID: "increase", Focusable: true, Border: true}, "Increase", func() {
						percent2 += 0.1
						if percent2 > 1.0 {
							percent2 = 0.0
						}
					}),
					splotch.NewButton(splotch.Style{ID: "decrease", Focusable: true, Border: true}, "Decrease", func() {
						percent2 -= 0.1
						if percent2 < 0.0 {
							percent2 = 1.0
						}
					}),
				),
				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					splotch.NewProgressBar(splotch.Style{Width: 100, Color: tcell.ColorBlue}, percent2),
					splotch.NewText(splotch.Style{Margin: splotch.Margin{Left: 1}}, fmt.Sprintf("%d%%", int(percent2*100))),
				),
			)
		},
		func(ev tcell.Event) {
			switch ev.(type) {
			case *splotch.EventTick:
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
