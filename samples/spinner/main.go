package main

import (
	"log"
	"time"
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
			s1 := tizzy.NewSpinner(ctx, tizzy.Style{Color: tcell.ColorBlue})

			s2 := tizzy.NewSpinner(ctx, tizzy.Style{Color: tcell.ColorGreen})
			s2.Frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			s2.Interval = 40 * time.Millisecond

			s3 := tizzy.NewSpinner(ctx, tizzy.Style{Color: tcell.ColorYellow})
			s3.Frames = []string{".  ", ".. ", "...", "   "}
			s3.Interval = 200 * time.Millisecond

			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Spinner Samples"),

				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					s1,
					tizzy.NewText(tizzy.Style{Margin: tizzy.Margin{Left: 1}}, "Default spinner (|/-)"),
				),

				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					s2,
					tizzy.NewText(tizzy.Style{Margin: tizzy.Margin{Left: 1}}, "Braille spinner"),
				),

				tizzy.NewBox(
					tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
					s3,
					tizzy.NewText(tizzy.Style{Margin: tizzy.Margin{Left: 1}}, "Dots spinner"),
				),
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
