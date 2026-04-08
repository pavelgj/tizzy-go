package main

import (
	"fmt"
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
	count := 0

	err = app.Run(
		func(ctx *splotch.RenderContext) splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Button Sample"),
				splotch.NewText(splotch.Style{}, "Press Tab to focus the button, then Enter to click."),
				splotch.NewButton(
					splotch.Style{
						ID:        "btn1",
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Left: 1, Right: 1},
					},
					"Click Me++",
					func() {
						count++
					},
				),
				splotch.NewButton(
					splotch.Style{
						ID:        "btn2",
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Left: 1, Right: 1},
					},
					"Click Me--",
					func() {
						count--
					},
				),
				splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, fmt.Sprintf("Button clicked %d times", count)),
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
