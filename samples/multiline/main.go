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

	// State
	inputValue := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7"

	err = app.Run(
		func(ctx *tizzy.RenderContext) tizzy.Node {
			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Multiline Text Input Sample"),
				tizzy.NewText(tizzy.Style{}, "Press Tab to focus, Use Arrows to navigate, Enter for new line."),
				tizzy.NewTextInput(
					ctx,
					tizzy.Style{
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Left: 1, Right: 1},
						Width:     30,
						Multiline: true,
						MaxHeight: 7,
					},
					inputValue,
					func(newValue string) {
						inputValue = newValue
					},
				),
				tizzy.NewText(tizzy.Style{}, "Current Value:"),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, inputValue),
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
