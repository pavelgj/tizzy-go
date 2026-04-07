package main

import (
	"fmt"
	"log"
	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	count := 0

	err = app.Run(
		func() splotch.Node {
			return splotch.NewBox(
				splotch.Style{FlexDirection: "column"},
				splotch.NewText(splotch.Style{}, "Welcome to Splotch!"),
				splotch.NewText(splotch.Style{}, fmt.Sprintf("Count: %d", count)),
				splotch.NewText(splotch.Style{}, "Press any key to increment. Press ESC to quit."),
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
