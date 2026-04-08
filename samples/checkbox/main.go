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
		checked, setChecked := splotch.UseState(ctx, false)
		checkedCustom, setCheckedCustom := splotch.UseState(ctx, false)

		status := "Unchecked"
		if checked {
			status = "Checked"
		}

		customStatus := "Unchecked"
		if checkedCustom {
			customStatus = "Checked"
		}

		cbCustom := splotch.NewCheckbox(ctx, splotch.Style{Focusable: true, Color: tcell.ColorGreen}, "Accept Terms", checkedCustom, func(v bool) {
			setCheckedCustom(v)
		})
		cbCustom.CheckedChar = "✔"
		cbCustom.UncheckedChar = "✘"

		return splotch.NewBox(
			splotch.Style{
				FlexDirection: "column",
				Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Checkbox Sample"),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "----------------"),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Standard Checkbox:"),
			splotch.NewCheckbox(ctx, splotch.Style{Focusable: true, Color: tcell.ColorWhite}, "Enable Feature", checked, func(v bool) {
				setChecked(v)
			}),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Status: "+status),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Custom Characters Checkbox:"),
			cbCustom,
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Status: "+customStatus),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, ""),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Use Tab/Shift-Tab to navigate, Space/Enter to toggle, or mouse click."),
		)
	}

	update := func(ev tcell.Event) {
		if evKey, ok := ev.(*tcell.EventKey); ok {
			if evKey.Key() == tcell.KeyEscape {
				// We don't have an explicit Exit method on App yet, but we could add it.
				// For now we just let the user use Ctrl-C which is handled by default usually.
			}
		}
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
