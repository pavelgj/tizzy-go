package main

import (
	"log"
	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func() splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Focus Management Sample"),
				splotch.NewText(splotch.Style{}, "Press Tab to cycle focus, ESC to quit."),
				splotch.NewBox(
					splotch.Style{
						ID:        "box1",
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
						Margin:    splotch.Margin{Top: 1, Bottom: 1},
					},
					splotch.NewText(splotch.Style{}, "Box 1 (Focusable)"),
				),
				splotch.NewBox(
					splotch.Style{
						ID:        "box2",
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
					},
					splotch.NewText(splotch.Style{}, "Box 2 (Focusable)"),
				),
			)
		},
		func(ev tcell.Event) {
			// Static sample, focus is handled by App!
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
