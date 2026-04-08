package main

import (
	"log"

	tz "github.com/pavelgj/tizzy-go/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		checked, setChecked := tz.UseState(ctx, false)
		checkedCustom, setCheckedCustom := tz.UseState(ctx, false)

		status := "Unchecked"
		if checked {
			status = "Checked"
		}

		customStatus := "Unchecked"
		if checkedCustom {
			customStatus = "Checked"
		}

		cbCustom := tz.NewCheckbox(ctx, tz.Style{Focusable: true, Color: tcell.ColorGreen}, "Accept Terms", checkedCustom, func(v bool) {
			setCheckedCustom(v)
		})
		cbCustom.CheckedChar = "✔"
		cbCustom.UncheckedChar = "✘"

		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Checkbox Sample"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "----------------"),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Standard Checkbox:"),
			tz.NewCheckbox(ctx, tz.Style{Focusable: true, Color: tcell.ColorWhite}, "Enable Feature", checked, func(v bool) {
				setChecked(v)
			}),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Status: "+status),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Custom Characters Checkbox:"),
			cbCustom,
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Status: "+customStatus),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to toggle, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
