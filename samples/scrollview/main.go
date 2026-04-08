package main

import (
	"log"
	"strconv"

	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	headers := []string{"ID", "Name", "Role", "Status"}
	var rows [][]string
	for i := 1; i <= 50; i++ {
		rows = append(rows, []string{
			strconv.Itoa(i),
			"User " + strconv.Itoa(i),
			"Role " + strconv.Itoa(i),
			"Active",
		})
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		return tizzy.NewBox(
			tizzy.Style{
				FlexDirection: "column",
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
				FillWidth:     true,
				FillHeight:    true,
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "ScrollView Sample (Scroll with Mouse Wheel)"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "------------------------------------------"),
			tizzy.NewScrollView(
				ctx,
				tizzy.Style{
					Color:      tcell.ColorWhite,
					Border:     true,
					FillWidth:  true,
					FillHeight: true,
					Focusable:  true,
					Title:      "Users Table",
				},
				tizzy.NewTable(tizzy.Style{Color: tcell.ColorWhite, FillWidth: true}, headers, rows),
			),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
