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
		func(ctx *splotch.RenderContext) splotch.Node {
			valObj, setVal := ctx.UseState("Type here...")
			inputValue := valObj.(string)

			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Text Input Sample"),
				splotch.NewText(splotch.Style{}, "Press Tab to focus input, ESC to quit."),
				splotch.NewTextInput(
					splotch.Style{
						ID:        "input1",
						Focusable: true,
						Border:    true,
						Padding:   splotch.Padding{Left: 1, Right: 1},
						Width:     20,
					},
					inputValue,
					func(newValue string) {
						setVal(newValue)
					},
				),
				splotch.NewText(splotch.Style{}, "You typed: "+inputValue),
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
