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

	render := func(ctx *splotch.RenderContext) splotch.Node {
		selectedValue, setSelectedVal := splotch.UseState(ctx, "option1")
		selectedCustomValue, setSelectedCustomVal := splotch.UseState(ctx, "apple")

		rbCustom1 := splotch.NewRadioButton(ctx, splotch.Style{Focusable: true, Color: tcell.ColorGreen}, "Apple", "apple", selectedCustomValue == "apple", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom1.SelectedChar = "⚫"
		rbCustom1.UnselectedChar = "⚪"

		rbCustom2 := splotch.NewRadioButton(ctx, splotch.Style{Focusable: true, Color: tcell.ColorYellow}, "Banana", "banana", selectedCustomValue == "banana", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom2.SelectedChar = "⚫"
		rbCustom2.UnselectedChar = "⚪"

		return splotch.NewBox(
			splotch.Style{
				FlexDirection: "column",
				Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Radio Button Sample"),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "-------------------"),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Standard Radio Buttons:"),
			splotch.NewRadioButton(ctx, splotch.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 1", "option1", selectedValue == "option1", func(v string) {
				setSelectedVal(v)
			}),
			splotch.NewRadioButton(ctx, splotch.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 2", "option2", selectedValue == "option2", func(v string) {
				setSelectedVal(v)
			}),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Selected: "+selectedValue),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Custom Characters Radio Buttons:"),
			rbCustom1,
			rbCustom2,
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Selected: "+selectedCustomValue),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to select, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
