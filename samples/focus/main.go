package main

import (
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
			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{}, "Focus Management Sample"),
				tz.NewText(tz.Style{}, "Press Tab to cycle focus, ESC to quit."),
				tz.NewBox(
					tz.Style{
						ID:        "box1",
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
						Margin:    tz.Margin{Top: 1, Bottom: 1},
					},
					tz.NewText(tz.Style{}, "Box 1 (Focusable)"),
				),
				tz.NewBox(
					tz.Style{
						ID:        "box2",
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
					},
					tz.NewText(tz.Style{}, "Box 2 (Focusable)"),
				),
			)
		},
		func(ev tcell.Event) {
			// Static sample, focus is handled by App!
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
