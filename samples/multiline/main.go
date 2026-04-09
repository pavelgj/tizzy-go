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

	// State
	inputValue := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7"

	err = app.Run(
		func(ctx *tz.RenderContext) tz.Node {
			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Multiline Text Input Sample"),
				tz.NewText(tz.Style{}, "Press Tab to focus, Use Arrows to navigate, Enter for new line."),
				tz.NewTextInput(
					ctx,
					tz.Style{
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
						Width:     30,
						Multiline: true,
						MaxHeight: 7,
					},
					inputValue,
					func(newValue string) {
						inputValue = newValue
					},
				),
				tz.NewText(tz.Style{}, "Current Value:"),
				tz.NewText(tz.Style{Color: tcell.ColorGreen}, inputValue),
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
