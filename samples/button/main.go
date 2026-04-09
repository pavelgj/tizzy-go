package main

import (
	"fmt"
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
			count, setCount := tz.UseState(ctx, 0)

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Button Sample"),
				tz.NewText(tz.Style{}, "Press Tab to focus the button, then Enter to click."),
				tz.NewButton(
					tz.Style{
						ID:        "btn1",
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
					},
					"Click Me++",
					func() {
						setCount(count + 1)
					},
				),
				tz.NewButton(
					tz.Style{
						ID:        "btn2",
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
					},
					"Click Me--",
					func() {
						setCount(count - 1)
					},
				),
				tz.NewText(tz.Style{Color: tcell.ColorGreen}, fmt.Sprintf("Button clicked %d times", count)),
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
