package main

import (
	"fmt"
	"os"

	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	render := func(ctx *splotch.RenderContext) splotch.Node {
		count, setCount := splotch.UseState(ctx, 0)
		countOutside, setCountOutside := splotch.UseState(ctx, 0)
		notificationsEnabled, setNotificationsEnabled := splotch.UseState(ctx, true)

		return splotch.NewBox(
			splotch.Style{
				Padding: splotch.Padding{Top: 1, Left: 2},
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Tabs Sample"),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Click on tabs or use Left/Right arrows to switch when focused."),

			splotch.NewTabs(
				ctx,
				splotch.Style{
					Focusable: true,
					Margin:    splotch.Margin{Top: 1},
					Color:     tcell.ColorWhite,
				},
				[]splotch.Tab{
					{
						Label: "Home",
						Content: splotch.NewBox(
							splotch.Style{Border: true, Padding: splotch.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "Welcome to Home Tab!"),
							splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, fmt.Sprintf("Button clicks: %d", count)),
							splotch.NewButton(splotch.Style{ID: "btn_home", Focusable: true, Margin: splotch.Margin{Top: 1}}, "Home Action", func() {
								setCount(count + 1)
							}),
						),
					},
					{
						Label: "Settings",
						Content: splotch.NewBox(
							splotch.Style{Border: true, Padding: splotch.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							splotch.NewText(splotch.Style{Color: tcell.ColorBlue}, "Settings Tab"),
							splotch.NewCheckbox(ctx, splotch.Style{Focusable: true}, "Enable Notifications", notificationsEnabled, func(val bool) {
								setNotificationsEnabled(val)
							}),
							splotch.NewTextInput(ctx, splotch.Style{Focusable: true, Width: 20, Margin: splotch.Margin{Top: 1}}, "Initial Value", func(val string) {}),
						),
					},
					{
						Label: "About",
						Content: splotch.NewBox(
							splotch.Style{Border: true, Padding: splotch.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							splotch.NewText(splotch.Style{Color: tcell.ColorDarkMagenta}, "About Tab"),
							splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Splotch TUI Library v0.1.0"),
						),
					},
				},
			),
			splotch.NewText(splotch.Style{Color: tcell.ColorWhite, Margin: splotch.Margin{Top: 1}}, fmt.Sprintf("Outside clicks: %d", countOutside)),
			splotch.NewButton(splotch.Style{ID: "btn1", Focusable: true}, "Focusable Button", func() {
				setCountOutside(countOutside + 1)
			}),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
