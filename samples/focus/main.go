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
			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column", Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tizzy.NewText(tizzy.Style{}, "Focus Management Sample"),
				tizzy.NewText(tizzy.Style{}, "Press Tab to cycle focus, ESC to quit."),
				tizzy.NewBox(
					tizzy.Style{
						ID:        "box1",
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
						Margin:    tizzy.Margin{Top: 1, Bottom: 1},
					},
					tizzy.NewText(tizzy.Style{}, "Box 1 (Focusable)"),
				),
				tizzy.NewBox(
					tizzy.Style{
						ID:        "box2",
						Focusable: true,
						Border:    true,
						Padding:   tizzy.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
					},
					tizzy.NewText(tizzy.Style{}, "Box 2 (Focusable)"),
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
