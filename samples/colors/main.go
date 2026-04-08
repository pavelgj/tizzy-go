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
				tizzy.Style{FlexDirection: "column", Border: true},
				tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Green Text"),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorRed, Background: tcell.ColorWhite}, "Red on White"),
				tizzy.NewBox(
					tizzy.Style{
						Border: true,
						Color:  tcell.ColorYellow,
					},
					tizzy.NewText(tizzy.Style{}, "Yellow Border Box"),
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
