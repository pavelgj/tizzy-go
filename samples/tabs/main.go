package main

import (
	"fmt"
	"os"

	"github.com/pavelgj/tizzy-go/tz"

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
		notificationsEnabled, setNotificationsEnabled := tz.UseState(ctx, true)
		textValue, setTextValue := tz.UseState(ctx, "Hello!")

		// --- Normal tabs (3 tabs, no overflow) ---
		normalTabs := tz.NewTabs(
			ctx,
			tz.Style{
				ID:        "normal-tabs",
				Focusable: true,
				Color:     tcell.ColorWhite,
			},
			[]tz.Tab{
				{
					Label: "Home",
					Content: tz.NewBox(
						tz.Style{
							Border:  true,
							Width:   50,
							Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2},
						},
						tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Welcome to the Home tab!"),
						tz.NewText(tz.Style{Color: tcell.ColorWhite}, fmt.Sprintf("Button presses: %d", count)),
						tz.NewButton(tz.Style{
							ID:        "btn_home",
							Focusable: true,
							Margin:    tz.Margin{Top: 1},
						}, "Press Me", func() {
							setCount(count + 1)
						}),
					),
				},
				{
					Label: "Settings",
					Content: tz.NewBox(
						tz.Style{
							Border:  true,
							Width:   50,
							Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2},
						},
						tz.NewCheckbox(ctx, tz.Style{
							ID:        "cb_notif",
							Focusable: true,
						}, "Enable notifications", notificationsEnabled, func(val bool) {
							setNotificationsEnabled(val)
						}),
						tz.NewText(tz.Style{
							Color:  tcell.ColorGray,
							Margin: tz.Margin{Top: 1},
						}, "Username:"),
						tz.NewTextInput(ctx, tz.Style{
							ID:        "ti_user",
							Focusable: true,
							Width:     24,
						}, textValue, func(val string) {
							setTextValue(val)
						}),
						tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
							fmt.Sprintf("Notifications: %v  |  User: %s", notificationsEnabled, textValue)),
					),
				},
				{
					Label: "About",
					Content: tz.NewBox(
						tz.Style{
							Border:  true,
							Width:   50,
							Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2},
						},
						tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Tizzy TUI Library"),
						tz.NewText(tz.Style{Color: tcell.ColorWhite}, "A declarative TUI framework for Go."),
						tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}},
							"Arrow keys or click tabs to switch."),
					),
				},
			},
		)

		// --- Overflow tabs (many tabs, fixed width forces < / > arrows) ---
		overflowTabs := tz.NewTabs(
			ctx,
			tz.Style{
				ID:        "overflow-tabs",
				Focusable: true,
				Width:     50,
				Color:     tcell.ColorWhite,
			},
			[]tz.Tab{
				{Label: "Alpha", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Alpha content — first tab")},
				{Label: "Beta", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Beta content")},
				{Label: "Gamma", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Gamma content")},
				{Label: "Delta", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Delta content")},
				{Label: "Epsilon", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Epsilon content")},
				{Label: "Zeta", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Zeta content")},
				{Label: "Eta", Content: tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Eta content — last tab")},
			},
		)

		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Left: 2},
			},
			tz.NewText(tz.Style{Color: tcell.ColorYellow}, "─── Normal tabs (3) ───"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Left/Right arrows to switch • Tab to focus inner controls"),
			normalTabs,

			tz.NewText(tz.Style{
				Color:  tcell.ColorYellow,
				Margin: tz.Margin{Top: 1},
			}, "─── Overflow tabs (7, width=50) ───"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Left/Right arrows to switch • < > arrows to scroll"),
			overflowTabs,
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
