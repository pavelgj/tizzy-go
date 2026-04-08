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

	headers := []string{"ID", "Name", "Role", "Status"}
	rows := [][]string{
		{"1", "Alice", "Developer", "Active"},
		{"2", "Bob", "Designer", "Away"},
		{"3", "Charlie", "Manager", "Busy"},
		{"4", "David", "DevOps", "Offline"},
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		return tizzy.NewBox(
			tizzy.Style{
				FlexDirection: "column",
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Table Sample"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "------------"),
			tizzy.NewTable(tizzy.Style{Color: tcell.ColorWhite}, headers, rows),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Table with Border:"),
			tizzy.NewTable(tizzy.Style{Color: tcell.ColorWhite, Border: true}, headers, rows),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Table filling width:"),
			tizzy.NewTable(tizzy.Style{Color: tcell.ColorWhite, Border: true, FillWidth: true}, headers, rows),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
