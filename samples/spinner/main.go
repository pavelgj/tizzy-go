package main

import (
	"log"
	"time"

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
			s1 := tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorBlue})

			s2 := tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorGreen})
			s2.Frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			s2.Interval = 40 * time.Millisecond

			s3 := tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorYellow})
			s3.Frames = []string{".  ", ".. ", "...", "   "}
			s3.Interval = 200 * time.Millisecond

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Spinner Samples"),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					s1,
					tz.NewText(tz.Style{Margin: tz.Margin{Left: 1}}, "Default spinner (|/-)"),
				),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					s2,
					tz.NewText(tz.Style{Margin: tz.Margin{Left: 1}}, "Braille spinner"),
				),

				tz.NewBox(
					tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
					s3,
					tz.NewText(tz.Style{Margin: tz.Margin{Left: 1}}, "Dots spinner"),
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
