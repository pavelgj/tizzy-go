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
				tz.Style{FlexDirection: "column", Border: true},
				tz.NewText(tz.Style{}, "Flexbox Centering Sample"),
				tz.NewBox(
					tz.Style{
						FlexDirection:  "row",
						JustifyContent: "center",
						Border:         true,
						Color:          tcell.ColorYellow,
					},
					tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Centered Text"),
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
