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
				splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "Green Text"),
				splotch.NewText(splotch.Style{Color: tcell.ColorRed, Background: tcell.ColorWhite}, "Red on White"),
				splotch.NewBox(
					splotch.Style{
						Border: true,
						Color:  tcell.ColorYellow,
					},
					splotch.NewText(splotch.Style{}, "Yellow Border Box"),
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
