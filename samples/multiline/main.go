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

	// State
	inputValue := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7"

	err = app.Run(
		func(ctx *splotch.RenderContext) splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Multiline Text Input Sample"),
				splotch.NewText(splotch.Style{}, "Press Tab to focus, Use Arrows to navigate, Enter for new line."),
				splotch.NewTextInput(
					ctx,
					splotch.Style{
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Left: 1, Right: 1},
						Width:     30,
						Multiline: true,
						MaxHeight: 7,
					},
					inputValue,
					func(newValue string) {
						inputValue = newValue
					},
				),
				splotch.NewText(splotch.Style{}, "Current Value:"),
				splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, inputValue),
			)
		},
		func(ev tcell.Event) {
			// App handles events
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
