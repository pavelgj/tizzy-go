package main

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v2"
	"splotch/splotch"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	app.SetState("menubar", &splotch.MenuBarState{OpenMenuIndex: -1})
	app.SetState("last_action", "None")

	render := func(ctx *splotch.RenderContext) splotch.Node {
		lastAction := "None"
		if val := app.GetState("last_action"); val != nil {
			lastAction = val.(string)
		}

		return splotch.NewBox(
			splotch.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
			},
			splotch.NewMenuBar(
				splotch.Style{
					ID:         "menubar",
					Color:      tcell.ColorWhite,
					Background: tcell.ColorTeal,
					FillWidth:  true,
					Focusable:  true,
				},
				[]splotch.Menu{
					{
						Title:   "File",
						AltRune: 'f',
						Items: []splotch.MenuItem{
							{Label: "New", Action: func() { app.SetState("last_action", "New") }},
							{Label: "Open", Action: func() { app.SetState("last_action", "Open") }},
							{Label: "Save", Action: func() { app.SetState("last_action", "Save") }},
							{Label: "Exit", Action: func() { app.Stop() }},
						},
					},
					{
						Title:   "Edit",
						AltRune: 'e',
						Items: []splotch.MenuItem{
							{Label: "Cut", Action: func() { app.SetState("last_action", "Cut") }},
							{Label: "Copy", Action: func() { app.SetState("last_action", "Copy") }},
							{Label: "Paste", Action: func() { app.SetState("last_action", "Paste") }},
						},
					},
					{
						Title:   "Help",
						AltRune: 'h',
						Items: []splotch.MenuItem{
							{Label: "About", Action: func() { app.SetState("last_action", "About") }},
						},
					},
				},
			),
			splotch.NewBox(
				splotch.Style{
					Padding: splotch.Padding{Top: 2, Left: 2},
				},
				splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "MenuBar Sample"),
				splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Press Alt+F, Alt+E, or Alt+H to open menus."),
				splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Or click on them with the mouse."),
				splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Or use Tab to focus the MenuBar (opens File menu)."),
				splotch.NewText(splotch.Style{Color: tcell.ColorGreen, Margin: splotch.Margin{Top: 1}}, "Last Action: "+lastAction),
				splotch.NewButton(splotch.Style{ID: "btn1", Focusable: true, Margin: splotch.Margin{Top: 1}}, "Focusable Button", func() {}),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
