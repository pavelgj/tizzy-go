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
		checked, setChecked := tizzy.UseState(ctx, false)
		checkedCustom, setCheckedCustom := tizzy.UseState(ctx, false)

		status := "Unchecked"
		if checked {
			status = "Checked"
		}

		customStatus := "Unchecked"
		if checkedCustom {
			customStatus = "Checked"
		}

		cbCustom := tizzy.NewCheckbox(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorGreen}, "Accept Terms", checkedCustom, func(v bool) {
			setCheckedCustom(v)
		})
		cbCustom.CheckedChar = "✔"
		cbCustom.UncheckedChar = "✘"

		return tizzy.NewBox(
			tizzy.Style{
				FlexDirection: "column",
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Checkbox Sample"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "----------------"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Standard Checkbox:"),
			tizzy.NewCheckbox(ctx, tizzy.Style{Focusable: true, Color: tcell.ColorWhite}, "Enable Feature", checked, func(v bool) {
				setChecked(v)
			}),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Status: "+status),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Custom Characters Checkbox:"),
			cbCustom,
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Status: "+customStatus),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, ""),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to toggle, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
