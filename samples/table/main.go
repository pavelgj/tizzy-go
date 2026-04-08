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

	headers := []string{"ID", "Name", "Role", "Status"}
	rows := [][]string{
		{"1", "Alice", "Developer", "Active"},
		{"2", "Bob", "Designer", "Away"},
		{"3", "Charlie", "Manager", "Busy"},
		{"4", "David", "DevOps", "Offline"},
	}

	render := func(ctx *splotch.RenderContext) splotch.Node {
		return splotch.NewBox(
			splotch.Style{
				FlexDirection: "column",
				Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Table Sample"),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "------------"),
			splotch.NewTable(splotch.Style{Color: tcell.ColorWhite}, headers, rows),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Table with Border:"),
			splotch.NewTable(splotch.Style{Color: tcell.ColorWhite, Border: true}, headers, rows),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Table filling width:"),
			splotch.NewTable(splotch.Style{Color: tcell.ColorWhite, Border: true, FillWidth: true}, headers, rows),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
