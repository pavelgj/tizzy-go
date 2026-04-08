package main

import (
	tz "github.com/pavelgj/tizzy-go/tizzy"
	"log"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func(ctx *tz.RenderContext) tz.Node {
			inputValue, setVal := tz.UseState(ctx, "Type here...")

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Text Input Sample"),
				tz.NewText(tz.Style{}, "Press Tab to focus input, ESC to quit."),
				tz.NewTextInput(
					ctx,
					tz.Style{
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
						Width:     20,
					},
					inputValue,
					func(newValue string) {
						setVal(newValue)
					},
				),
				tz.NewText(tz.Style{}, "You typed: "+inputValue),
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
