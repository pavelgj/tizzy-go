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

	count := 0

	err = app.Run(
		func(ctx *tizzy.RenderContext) tizzy.Node {
			return tizzy.NewBox(
				tizzy.Style{FlexDirection: "column"},
				tizzy.NewText(tizzy.Style{}, "Welcome to Tizzy!"),
				tizzy.NewText(tizzy.Style{}, fmt.Sprintf("Count: %d", count)),
				tizzy.NewText(tizzy.Style{}, "Press any key to increment. Press ESC to quit."),
			)
		},
		func(ev tcell.Event) {
			switch ev := ev.(type) {
			case *tcell.EventKey:
				if ev.Key() != tcell.KeyEscape && ev.Key() != tcell.KeyCtrlC {
					count++
				}
			}
		},
	)

	if err != nil {
		log.Fatal(err)
	}
}
