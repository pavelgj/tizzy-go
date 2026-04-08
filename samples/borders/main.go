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
		func(ctx *splotch.RenderContext) splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Border: true, Title: "Outer Box"},
				splotch.NewText(splotch.Style{}, "Content inside outer box"),
				splotch.NewBox(
					splotch.Style{
						Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
						Border:  true,
						Title:   "Inner Box",
					},
					splotch.NewText(splotch.Style{}, "Content inside inner box"),
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
