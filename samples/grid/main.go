package main

import (
	"fmt"
	"os"
	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	root := splotch.NewGridBox(
		splotch.Style{
			Border:     true,
			FillWidth:  true,
			FillHeight: true,
			Title:      "Main Grid",
		},
		// Columns: Fixed 15 cells for sidebar, 1fraction for main, Fixed 15 for right panel
		[]splotch.GridTrack{splotch.Fixed(15), splotch.Flex(1), splotch.Fixed(15)},
		// Rows: Fixed 3 for header, 1fraction for main area, Fixed 3 for footer
		[]splotch.GridTrack{splotch.Fixed(3), splotch.Flex(1), splotch.Fixed(3)},

		// Header spanning all 3 columns
		splotch.NewBox(splotch.Style{GridRow: 0, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorBlue, FillWidth: true, FillHeight: true, Title: "Header"},
			splotch.NewText(splotch.Style{}, "Spans 3 Columns"),
		),

		// Left Sidebar
		splotch.NewBox(splotch.Style{GridRow: 1, GridCol: 0, Border: true, Background: tcell.ColorGreen, FillWidth: true, FillHeight: true, Title: "Sidebar"},
			splotch.NewText(splotch.Style{}, "Navigation"),
		),

		// Main Content
		splotch.NewBox(splotch.Style{GridRow: 1, GridCol: 1, Border: true, FillWidth: true, FillHeight: true, Title: "Main Content"},
			splotch.NewBox(splotch.Style{Padding: splotch.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1}},
				splotch.NewText(splotch.Style{}, "This area takes up the remaining space."),
			),
		),

		// Right Panel
		splotch.NewBox(splotch.Style{GridRow: 1, GridCol: 2, Border: true, Background: tcell.ColorYellow, FillWidth: true, FillHeight: true, Title: "Right Panel"},
			splotch.NewText(splotch.Style{}, "Info"),
		),

		// Footer spanning all 3 columns
		splotch.NewBox(splotch.Style{GridRow: 2, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorRed, FillWidth: true, FillHeight: true, Title: "Footer"},
			splotch.NewText(splotch.Style{}, "Spans 3 Columns"),
		),
	)

	app, err := splotch.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	render := func(ctx *splotch.RenderContext) splotch.Node {
		return root
	}

	update := func(ev tcell.Event) {
		if key, ok := ev.(*tcell.EventKey); ok {
			if key.Key() == tcell.KeyEscape || key.Key() == tcell.KeyCtrlC {
				app.Stop()
			}
		}
	}

	if err := app.Run(render, update); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
