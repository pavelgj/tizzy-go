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
				splotch.Style{FlexDirection: "column", Border: true},
				splotch.NewText(splotch.Style{}, "Outer Box with Border"),
				splotch.NewBox(
					splotch.Style{
						Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
						Border:  true,
					},
					splotch.NewText(splotch.Style{}, "Inner Box with Padding & Border"),
				),
				splotch.NewText(splotch.Style{}, "Press ESC to quit."),
			)
		},
		func(ev tcell.Event) {
			// Static sample, no state updates
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
