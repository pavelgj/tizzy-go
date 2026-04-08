package main

import (
	"log"
	"strconv"

	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
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

	render := func(ctx *splotch.RenderContext) splotch.Node {
		return splotch.NewBox(
			splotch.Style{
				FlexDirection: "column",
				Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
				FillWidth:     true,
				FillHeight:    true,
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "ScrollView Sample (Scroll with Mouse Wheel)"),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "------------------------------------------"),
			splotch.NewScrollView(
				ctx,
				splotch.Style{
					Color:      tcell.ColorWhite,
					Border:     true,
					FillWidth:  true,
					FillHeight: true,
					Focusable:  true,
				},
				splotch.NewTable(splotch.Style{Color: tcell.ColorWhite, FillWidth: true}, headers, rows),
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
