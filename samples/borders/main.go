package main

import (
	"log"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func(ctx *tizzy.RenderContext) tizzy.Node {
			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Border: true, Title: "Outer Box"},
				tizzy.NewText(tizzy.Style{}, "Content inside outer box"),
				tizzy.NewBox(
					tizzy.Style{
						Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
						Border:  true,
						Title:   "Inner Box",
					},
					tizzy.NewText(tizzy.Style{}, "Content inside inner box"),
				),
				tizzy.NewText(tizzy.Style{}, "Press ESC to quit."),
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
