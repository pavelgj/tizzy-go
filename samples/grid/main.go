package main

import (
	"fmt"
	"os"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	root := tz.NewGridBox(
		tz.Style{
			Border:     true,
			FillWidth:  true,
			FillHeight: true,
			Title:      "Main Grid",
		},
		// Columns: Fixed 15 cells for sidebar, 1fraction for main, Fixed 15 for right panel
		[]tz.GridTrack{tz.Fixed(15), tz.Flex(1), tz.Fixed(15)},
		// Rows: Fixed 3 for header, 1fraction for main area, Fixed 3 for footer
		[]tz.GridTrack{tz.Fixed(3), tz.Flex(1), tz.Fixed(3)},

		// Header spanning all 3 columns
		tz.NewBox(tz.Style{GridRow: 0, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorBlue, FillWidth: true, FillHeight: true, Title: "Header"},
			tz.NewText(tz.Style{}, "Spans 3 Columns"),
		),

		// Left Sidebar
		tz.NewBox(tz.Style{GridRow: 1, GridCol: 0, Border: true, Background: tcell.ColorGreen, FillWidth: true, FillHeight: true, Title: "Sidebar"},
			tz.NewText(tz.Style{}, "Navigation"),
		),

		// Main Content
		tz.NewBox(tz.Style{GridRow: 1, GridCol: 1, Border: true, FillWidth: true, FillHeight: true, Title: "Main Content"},
			tz.NewBox(tz.Style{Padding: tz.Padding{Left: 1, Right: 1, Top: 1, Bottom: 1}},
				tz.NewText(tz.Style{}, "This area takes up the remaining space."),
			),
		),

		// Right Panel
		tz.NewBox(tz.Style{GridRow: 1, GridCol: 2, Border: true, Background: tcell.ColorYellow, FillWidth: true, FillHeight: true, Title: "Right Panel"},
			tz.NewText(tz.Style{}, "Info"),
		),

		// Footer spanning all 3 columns
		tz.NewBox(tz.Style{GridRow: 2, GridCol: 0, GridColSpan: 3, Border: true, Background: tcell.ColorRed, FillWidth: true, FillHeight: true, Title: "Footer"},
			tz.NewText(tz.Style{}, "Spans 3 Columns"),
		),
	)

	app, err := tz.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
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
