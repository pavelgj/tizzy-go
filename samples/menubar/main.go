package main

import (
	"fmt"
	"os"

	tz "github.com/pavelgj/tizzy-go/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		lastAction, setLastAction := tz.UseState(ctx, "None")

		return tz.NewBox(
			tz.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
			},
			tz.NewMenuBar(
				ctx,
				tz.Style{
					Color:      tcell.ColorWhite,
					Background: tcell.ColorTeal,
					FillWidth:  true,
					Focusable:  true,
				},
				[]tz.Menu{
					{
						Title:   "File",
						AltRune: 'f',
						Items: []tz.MenuItem{
							{Label: "New", Action: func() { setLastAction("New") }},
							{Label: "Open", Action: func() { setLastAction("Open") }},
							{Label: "Save", Action: func() { setLastAction("Save") }},
							{Label: "Exit", Action: func() { app.Stop() }},
						},
					},
					{
						Title:   "Edit",
						AltRune: 'e',
						Items: []tz.MenuItem{
							{Label: "Cut", Action: func() { setLastAction("Cut") }},
							{Label: "Copy", Action: func() { setLastAction("Copy") }},
							{Label: "Paste", Action: func() { setLastAction("Paste") }},
						},
					},
					{
						Title:   "Help",
						AltRune: 'h',
						Items: []tz.MenuItem{
							{Label: "About", Action: func() { setLastAction("About") }},
						},
					},
				},
			),
			tz.NewBox(
				tz.Style{
					Padding: tz.Padding{Top: 2, Left: 2},
				},
				tz.NewText(tz.Style{Color: tcell.ColorYellow}, "MenuBar Sample"),
				tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Press Alt+F, Alt+E, or Alt+H to open menus."),
				tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Or click on them with the mouse."),
				tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Or use Tab to focus the MenuBar (opens File menu)."),
				tz.NewText(tz.Style{Color: tcell.ColorGreen, Margin: tz.Margin{Top: 1}}, "Last Action: "+lastAction),
				tz.NewButton(tz.Style{ID: "btn1", Focusable: true, Margin: tz.Margin{Top: 1}}, "Focusable Button", func() {}),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
