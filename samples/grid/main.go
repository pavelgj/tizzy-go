package main

import (
	"fmt"
	"os"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	root := tizzy.NewGridBox(
		tizzy.Style{
			Border:     true,
			FillWidth:  true,
			FillHeight: true,
			Title:      "Main Grid",
		},
		// Columns: Fixed 15 cells for sidebar, 1fraction for main, Fixed 15 for right panel
		[]tizzy.GridTrack{tizzy.Fixed(15), tizzy.Flex(1), tizzy.Fixed(15)},
		// Rows: Fixed 3 for header, 1fraction for main area, Fixed 3 for footer
		[]tizzy.GridTrack{tizzy.Fixed(3), tizzy.Flex(1), tizzy.Fixed(3)},

		// Header spanning all 3 columns
		tizzy.NewBox(tizzy.Style{GridRow: 0, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorBlue, FillWidth: true, FillHeight: true, Title: "Header"},
			tizzy.NewText(tizzy.Style{}, "Spans 3 Columns"),
		),

		// Left Sidebar
		tizzy.NewBox(tizzy.Style{GridRow: 1, GridCol: 0, Border: true, Background: tcell.ColorGreen, FillWidth: true, FillHeight: true, Title: "Sidebar"},
			tizzy.NewText(tizzy.Style{}, "Navigation"),
		),

		// Main Content
		tizzy.NewBox(tizzy.Style{GridRow: 1, GridCol: 1, Border: true, FillWidth: true, FillHeight: true, Title: "Main Content"},
			tizzy.NewBox(tizzy.Style{Padding: tizzy.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1}},
				tizzy.NewText(tizzy.Style{}, "This area takes up the remaining space."),
			),
		),

		// Right Panel
		tizzy.NewBox(tizzy.Style{GridRow: 1, GridCol: 2, Border: true, Background: tcell.ColorYellow, FillWidth: true, FillHeight: true, Title: "Right Panel"},
			tizzy.NewText(tizzy.Style{}, "Info"),
		),

		// Footer spanning all 3 columns
		tizzy.NewBox(tizzy.Style{GridRow: 2, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorRed, FillWidth: true, FillHeight: true, Title: "Footer"},
			tizzy.NewText(tizzy.Style{}, "Spans 3 Columns"),
		),
	)

	app, err := tizzy.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
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
