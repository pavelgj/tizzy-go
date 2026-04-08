package main

import (
	"fmt"
	"splotch/splotch"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		fmt.Println(err)
		return
	}

	render := func(ctx *splotch.RenderContext) splotch.Node {
		simpleOpen, setSimpleOpen := splotch.UseState(ctx, false)
		confirmOpen, setConfirmOpen := splotch.UseState(ctx, false)
		formOpen, setFormOpen := splotch.UseState(ctx, false)

		return splotch.NewBox(
			splotch.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
				Padding:    splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			},
			splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Modal / Dialog Samples"),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "Press buttons to open different modals."),
			splotch.NewText(splotch.Style{Color: tcell.ColorGray}, "------------------------------------------"),

			splotch.NewBox(
				splotch.Style{FlexDirection: "row"},
				splotch.NewButton(
					splotch.Style{ID: "btn_simple", Focusable: true, Color: tcell.ColorWhite},
					"Simple Modal",
					func() { setSimpleOpen(true) },
				),
				splotch.NewButton(
					splotch.Style{ID: "btn_confirm", Focusable: true, Color: tcell.ColorWhite},
					"Confirmation",
					func() { setConfirmOpen(true) },
				),
				splotch.NewButton(
					splotch.Style{ID: "btn_form", Focusable: true, Color: tcell.ColorWhite},
					"Form Modal",
					func() { setFormOpen(true) },
				),
			),

			// Simple Modal
			splotch.NewModal(
				ctx,
				splotch.Style{Color: tcell.ColorWhite, Background: tcell.ColorBlue},
				splotch.NewBox(
					splotch.Style{Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Simple Modal"),
					splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "This is a simple message."),
					splotch.NewButton(
						splotch.Style{ID: "close_simple", Focusable: true, Color: tcell.ColorWhite},
						"Close",
						func() { setSimpleOpen(false) },
					),
				),
				simpleOpen,
			),

			// Confirmation Dialog
			splotch.NewModal(
				ctx,
				splotch.Style{Color: tcell.ColorWhite, Background: tcell.ColorGreen},
				splotch.NewBox(
					splotch.Style{Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Are you sure?"),
					splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Do you want to proceed?"),
					splotch.NewBox(
						splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
						splotch.NewButton(
							splotch.Style{ID: "confirm_yes", Focusable: true, Color: tcell.ColorWhite},
							"Yes",
							func() { setConfirmOpen(false) },
						),
						splotch.NewButton(
							splotch.Style{ID: "confirm_no", Focusable: true, Color: tcell.ColorWhite},
							"No",
							func() { setConfirmOpen(false) },
						),
					),
				),
				confirmOpen,
			),

			// Form Modal
			splotch.NewModal(
				ctx,
				splotch.Style{Color: tcell.ColorWhite, Background: tcell.ColorDarkMagenta},
				splotch.NewBox(
					splotch.Style{Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Interactive Form"),
					splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Tab between fields:"),

					splotch.NewBox(
						splotch.Style{Margin: splotch.Margin{Top: 1, Bottom: 1}},
						splotch.NewButton(splotch.Style{ID: "form_opt1", Focusable: true, Color: tcell.ColorWhite}, "Option 1", nil),
						splotch.NewButton(splotch.Style{ID: "form_opt2", Focusable: true, Color: tcell.ColorWhite}, "Option 2", nil),
					),

					splotch.NewBox(
						splotch.Style{FlexDirection: "row"},
						splotch.NewButton(
							splotch.Style{ID: "form_submit", Focusable: true, Color: tcell.ColorWhite},
							"Submit",
							func() { setFormOpen(false) },
						),
						splotch.NewButton(
							splotch.Style{ID: "form_cancel", Focusable: true, Color: tcell.ColorWhite},
							"Cancel",
							func() { setFormOpen(false) },
						),
					),
				),
				formOpen,
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
	}
}
