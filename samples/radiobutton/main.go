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

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		selectedValue, setSelectedVal := tizzy.UseState(ctx, "option1")
		selectedCustomValue, setSelectedCustomVal := tizzy.UseState(ctx, "apple")

		rbCustom1 := tizzy.NewRadioButton(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorGreen}, "Apple", "apple", selectedCustomValue == "apple", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom1.SelectedChar = "⚫"
		rbCustom1.UnselectedChar = "⚪"

		rbCustom2 := tizzy.NewRadioButton(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorYellow}, "Banana", "banana", selectedCustomValue == "banana", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom2.SelectedChar = "⚫"
		rbCustom2.UnselectedChar = "⚪"

		return tizzy.NewBox(
			tizzy.Style{
				FlexDirection: "column",
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Radio Button Sample"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "-------------------"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Standard Radio Buttons:"),
			tizzy.NewRadioButton(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 1", "option1", selectedValue == "option1", func(v string) {
				setSelectedVal(v)
			}),
			tizzy.NewRadioButton(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 2", "option2", selectedValue == "option2", func(v string) {
				setSelectedVal(v)
			}),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Selected: "+selectedValue),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Custom Characters Radio Buttons:"),
			rbCustom1,
			rbCustom2,
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Selected: "+selectedCustomValue),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to select, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
