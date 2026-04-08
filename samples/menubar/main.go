package main

import (
	"fmt"
	"os"

	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		lastAction, setLastAction := tizzy.UseState(ctx, "None")

		return tizzy.NewBox(
			tizzy.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
			},
			tizzy.NewMenuBar(
				ctx,
				tizzy.Style{
					Color:      tcell.ColorWhite,
					Background: tcell.ColorTeal,
					FillWidth:  true,
					Focusable:  true,
				},
				[]tizzy.Menu{
					{
						Title:   "File",
						AltRune: 'f',
						Items: []tizzy.MenuItem{
							{Label: "New", Action: func() { setLastAction("New") }},
							{Label: "Open", Action: func() { setLastAction("Open") }},
							{Label: "Save", Action: func() { setLastAction("Save") }},
							{Label: "Exit", Action: func() { app.Stop() }},
						},
					},
					{
						Title:   "Edit",
						AltRune: 'e',
						Items: []tizzy.MenuItem{
							{Label: "Cut", Action: func() { setLastAction("Cut") }},
							{Label: "Copy", Action: func() { setLastAction("Copy") }},
							{Label: "Paste", Action: func() { setLastAction("Paste") }},
						},
					},
					{
						Title:   "Help",
						AltRune: 'h',
						Items: []tizzy.MenuItem{
							{Label: "About", Action: func() { setLastAction("About") }},
						},
					},
				},
			),
			tizzy.NewBox(
				tizzy.Style{
					Padding: tizzy.Padding{Top: 2, Left: 2},
				},
				tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "MenuBar Sample"),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Press Alt+F, Alt+E, or Alt+H to open menus."),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Or click on them with the mouse."),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Or use Tab to focus the MenuBar (opens File menu)."),
				tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen, Margin: tizzy.Margin{Top: 1}}, "Last Action: "+lastAction),
				tizzy.NewButton(tizzy.Style{ID: "btn1", Focusable: true, Margin: tizzy.Margin{Top: 1}}, "Focusable Button", func() {}),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
