package main

import (
	"fmt"
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
			count, setCount := tizzy.UseState(ctx, 0)

			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Button Sample"),
				tizzy.NewText(tizzy.Style{}, "Press Tab to focus the button, then Enter to click."),
				tizzy.NewButton(
					tizzy.Style{
						ID:        "btn1",
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Left: 1, Right: 1},
					},
					"Click Me++",
					func() {
						setCount(count + 1)
					},
				),
				tizzy.NewButton(
					tizzy.Style{
						ID:        "btn2",
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Left: 1, Right: 1},
					},
					"Click Me--",
					func() {
						setCount(count - 1)
					},
				),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, fmt.Sprintf("Button clicked %d times", count)),
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
