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

	render := func(ctx *tz.RenderContext) tz.Node {
		selectedValue, setSelectedVal := tz.UseState(ctx, "option1")
		selectedCustomValue, setSelectedCustomVal := tz.UseState(ctx, "apple")

		rbCustom1 := tz.NewRadioButton(ctx, tz.Style{Focusable: true, Color: tcell.ColorGreen}, "Apple", "apple", selectedCustomValue == "apple", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom1.SelectedChar = "⚫"
		rbCustom1.UnselectedChar = "⚪"

		rbCustom2 := tz.NewRadioButton(ctx, tz.Style{Focusable: true, Color: tcell.ColorYellow}, "Banana", "banana", selectedCustomValue == "banana", func(v string) {
			setSelectedCustomVal(v)
		})
		rbCustom2.SelectedChar = "⚫"
		rbCustom2.UnselectedChar = "⚪"

		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Radio Button Sample"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "-------------------"),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Standard Radio Buttons:"),
			tz.NewRadioButton(ctx, tz.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 1", "option1", selectedValue == "option1", func(v string) {
				setSelectedVal(v)
			}),
			tz.NewRadioButton(ctx, tz.Style{Focusable: true, Color: tcell.ColorWhite}, "Option 2", "option2", selectedValue == "option2", func(v string) {
				setSelectedVal(v)
			}),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Selected: "+selectedValue),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Custom Characters Radio Buttons:"),
			rbCustom1,
			rbCustom2,
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Selected: "+selectedCustomValue),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to select, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {
		// Custom global event handling if needed
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
