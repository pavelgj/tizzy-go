package main

import (
	"log"
	"splotch/splotch"
	"time"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func() splotch.Node {
			s1 := splotch.NewSpinner(splotch.Style{Color: tcell.ColorBlue})

			s2 := splotch.NewSpinner(splotch.Style{Color: tcell.ColorGreen})
			s2.Frames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			s2.Interval = 40 * time.Millisecond

			s3 := splotch.NewSpinner(splotch.Style{Color: tcell.ColorYellow})
			s3.Frames = []string{".  ", ".. ", "...", "   "}
			s3.Interval = 200 * time.Millisecond

			return splotch.NewBox(
				splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				splotch.NewText(splotch.Style{}, "Spinner Samples"),

				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					s1,
					splotch.NewText(splotch.Style{Margin: splotch.Margin{Left: 1}}, "Default spinner (|/-)"),
				),

				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					s2,
					splotch.NewText(splotch.Style{Margin: splotch.Margin{Left: 1}}, "Braille spinner"),
				),

				splotch.NewBox(
					splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
					s3,
					splotch.NewText(splotch.Style{Margin: splotch.Margin{Left: 1}}, "Dots spinner"),
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
