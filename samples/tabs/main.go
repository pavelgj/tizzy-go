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
		count, setCount := tizzy.UseState(ctx, 0)
		countOutside, setCountOutside := tizzy.UseState(ctx, 0)
		notificationsEnabled, setNotificationsEnabled := tizzy.UseState(ctx, true)

		return tizzy.NewBox(
			tizzy.Style{
				Padding: tizzy.Padding{Top: 1, Left: 2},
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "Tabs Sample"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Click on tabs or use Left/Right arrows to switch when focused."),

			tizzy.NewTabs(
				ctx,
				tizzy.Style{
					Focusable: true,
					Margin:    tizzy.Margin{Top: 1},
					Color:     tcell.ColorWhite,
				},
				[]tizzy.Tab{
					{
						Label: "Home",
						Content: tizzy.NewBox(
							tizzy.Style{Border: true, Padding: tizzy.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Welcome to Home Tab!"),
							tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, fmt.Sprintf("Button clicks: %d", count)),
							tizzy.NewButton(tizzy.Style{ID: "btn_home", Focusable: true, Margin: tizzy.Margin{Top: 1}}, "Home Action", func() {
								setCount(count + 1)
							}),
						),
					},
					{
						Label: "Settings",
						Content: tizzy.NewBox(
							tizzy.Style{Border: true, Padding: tizzy.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tizzy.NewText(tizzy.Style{Color: tcell.ColorBlue}, "Settings Tab"),
							tizzy.NewCheckbox(ctx, tizzy.Style{Focusable: true}, "Enable Notifications", notificationsEnabled, func(val bool) {
								setNotificationsEnabled(val)
							}),
							tizzy.NewTextInput(ctx, tizzy.Style{Focusable: true, Width: 20, Margin: tizzy.Margin{Top: 1}}, "Initial Value", func(val string) {}),
						),
					},
					{
						Label: "About",
						Content: tizzy.NewBox(
							tizzy.Style{Border: true, Padding: tizzy.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tizzy.NewText(tizzy.Style{Color: tcell.ColorDarkMagenta}, "About Tab"),
							tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Tizzy TUI Library v0.1.0"),
						),
					},
				},
			),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite, Margin: tizzy.Margin{Top: 1}}, fmt.Sprintf("Outside clicks: %d", countOutside)),
			tizzy.NewButton(tizzy.Style{ID: "btn1", Focusable: true}, "Focusable Button", func() {
				setCountOutside(countOutside + 1)
			}),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
