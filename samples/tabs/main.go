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
		count, setCount := tz.UseState(ctx, 0)
		countOutside, setCountOutside := tz.UseState(ctx, 0)
		notificationsEnabled, setNotificationsEnabled := tz.UseState(ctx, true)

		return tz.NewBox(
			tz.Style{
				Padding: tz.Padding{Top: 1, Left: 2},
			},
			tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Tabs Sample"),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Click on tabs or use Left/Right arrows to switch when focused."),

			tz.NewTabs(
				ctx,
				tz.Style{
					Focusable: true,
					Margin:    tz.Margin{Top: 1},
					Color:     tcell.ColorWhite,
				},
				[]tz.Tab{
					{
						Label: "Home",
						Content: tz.NewBox(
							tz.Style{Border: true, Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Welcome to Home Tab!"),
							tz.NewText(tz.Style{Color: tcell.ColorWhite}, fmt.Sprintf("Button clicks: %d", count)),
							tz.NewButton(tz.Style{ID: "btn_home", Focusable: true, Margin: tz.Margin{Top: 1}}, "Home Action", func() {
								setCount(count + 1)
							}),
						),
					},
					{
						Label: "Settings",
						Content: tz.NewBox(
							tz.Style{Border: true, Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tz.NewText(tz.Style{Color: tcell.ColorBlue}, "Settings Tab"),
							tz.NewCheckbox(ctx, tz.Style{Focusable: true}, "Enable Notifications", notificationsEnabled, func(val bool) {
								setNotificationsEnabled(val)
							}),
							tz.NewTextInput(ctx, tz.Style{Focusable: true, Width: 20, Margin: tz.Margin{Top: 1}}, "Initial Value", func(val string) {}),
						),
					},
					{
						Label: "About",
						Content: tz.NewBox(
							tz.Style{Border: true, Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}},
							tz.NewText(tz.Style{Color: tcell.ColorDarkMagenta}, "About Tab"),
							tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Tizzy TUI Library v0.1.0"),
						),
					},
				},
			),
			tz.NewText(tz.Style{Color: tcell.ColorWhite, Margin: tz.Margin{Top: 1}}, fmt.Sprintf("Outside clicks: %d", countOutside)),
			tz.NewButton(tz.Style{ID: "btn1", Focusable: true}, "Focusable Button", func() {
				setCountOutside(countOutside + 1)
			}),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
