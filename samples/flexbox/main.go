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
				tizzy.NewText(tizzy.Style{}, "Flexbox Centering Sample"),
				tizzy.NewBox(
					tizzy.Style{
						FlexDirection:  "row",
						JustifyContent: "center",
						Border:         true,
						Color:          tcell.ColorYellow,
					},
					tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Centered Text"),
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
