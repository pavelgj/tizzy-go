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
			inputValue, setVal := tizzy.UseState(ctx, "Type here...")

			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Text Input Sample"),
				tizzy.NewText(tizzy.Style{}, "Press Tab to focus input, ESC to quit."),
				tizzy.NewTextInput(
					ctx,
					tizzy.Style{
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Left: 1, Right: 1},
						Width:     20,
					},
					inputValue,
					func(newValue string) {
						setVal(newValue)
					},
				),
				tizzy.NewText(tizzy.Style{}, "You typed: "+inputValue),
			)
		},
		func(ev tcell.Event) {
			// App handles text input events internally and calls OnChange!
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
