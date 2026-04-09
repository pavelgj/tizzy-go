package main

import (
	"log"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func(ctx *tz.RenderContext) tz.Node {
			return tz.NewBox(
				tz.Style{FlexDirection: "column", Border: true, Title: "Outer Box"},
				tz.NewText(tz.Style{}, "Content inside outer box"),
				tz.NewBox(
					tz.Style{
						Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
						Border:  true,
						Title:   "Inner Box",
					},
					tz.NewText(tz.Style{}, "Content inside inner box"),
				),
				tz.NewText(tz.Style{}, "Press ESC to quit."),
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
