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

	headers := []string{"ID", "Name", "Role", "Status"}
	rows := [][]string{
		{"1", "Alice", "Developer", "Active"},
		{"2", "Bob", "Designer", "Away"},
		{"3", "Charlie", "Manager", "Busy"},
		{"4", "David", "DevOps", "Offline"},
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Table Sample"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "------------"),
			tz.NewTable(tz.Style{Color: tcell.ColorWhite}, headers, rows),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Table with Border:"),
			tz.NewTable(tz.Style{Color: tcell.ColorWhite, Border: true}, headers, rows),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Table filling width:"),
			tz.NewTable(tz.Style{Color: tcell.ColorWhite, Border: true, FillWidth: true}, headers, rows),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
