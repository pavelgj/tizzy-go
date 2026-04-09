package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	options := []string{"Option 1", "Option 2", "Option 3", "Option 4", "Option 5", "Option 6"}
	longOptions := []string{
		"Item 1", "Item 2", "Item 3", "Item 4", "Item 5",
		"Item 6", "Item 7", "Item 8", "Item 9", "Item 10",
		"Item 11", "Item 12", "Item 13", "Item 14", "Item 15",
		"Item 16", "Item 17", "Item 18", "Item 19", "Item 20",
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		selectedIndex1, setSelectedIndex1 := tz.UseState(ctx, 0)
		selectedIndex2, setSelectedIndex2 := tz.UseState(ctx, 0)
		selectedIndex3, setSelectedIndex3 := tz.UseState(ctx, 0)

		return tz.NewBox(
			tz.Style{
				Width:          40,
				Height:         25,
				Border:         true,
				FlexDirection:  "column",
				JustifyContent: "center",
				Background:     tcell.ColorBlack,
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Dropdown Sample"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "---------------"),

			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Dropdown 1 (Default Limit 5):"),
			tz.NewDropdown(
				ctx,
				tz.Style{
					Color:     tcell.ColorWhite,
					Border:    true,
					Focusable: true,
				},
				options,
				selectedIndex1,
				func(idx int) {
					setSelectedIndex1(idx)
				},
			),
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Selected index: "+strconv.Itoa(selectedIndex1)),

			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Dropdown 2 (Short List):"),
			tz.NewDropdown(
				ctx,
				tz.Style{
					Color:     tcell.ColorWhite,
					Border:    true,
					Focusable: true,
				},
				options[:3],
				selectedIndex2,
				func(idx int) {
					setSelectedIndex2(idx)
				},
			),
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Selected index: "+strconv.Itoa(selectedIndex2)),

			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Dropdown 3 (Limit 10):"),
			tz.NewDropdown(
				ctx,
				tz.Style{
					Color:     tcell.ColorWhite,
					Border:    true,
					Focusable: true,
				},
				longOptions,
				selectedIndex3,
				func(idx int) {
					setSelectedIndex3(idx)
				},
				10,
			),
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Selected index: "+strconv.Itoa(selectedIndex3)),
		)
	}

	_ = app.Run(render, func(ev tcell.Event) {})
}
