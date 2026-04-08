package main

import (
	"fmt"
	tz "github.com/pavelgj/tizzy-go/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		fmt.Println(err)
		return
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		simpleOpen, setSimpleOpen := tz.UseState(ctx, false)
		confirmOpen, setConfirmOpen := tz.UseState(ctx, false)
		formOpen, setFormOpen := tz.UseState(ctx, false)

		return tz.NewBox(
			tz.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
				Padding:    tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			},
			tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Modal / Dialog Samples"),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "Press buttons to open different modals."),
			tz.NewText(tz.Style{Color: tcell.ColorGray}, "------------------------------------------"),

			tz.NewBox(
				tz.Style{FlexDirection: "row"},
				tz.NewButton(
					tz.Style{ID: "btn_simple", Focusable: true, Color: tcell.ColorWhite},
					"Simple Modal",
					func() { setSimpleOpen(true) },
				),
				tz.NewButton(
					tz.Style{ID: "btn_confirm", Focusable: true, Color: tcell.ColorWhite},
					"Confirmation",
					func() { setConfirmOpen(true) },
				),
				tz.NewButton(
					tz.Style{ID: "btn_form", Focusable: true, Color: tcell.ColorWhite},
					"Form Modal",
					func() { setFormOpen(true) },
				),
			),

			// Simple Modal
			tz.NewModal(
				ctx,
				tz.Style{Color: tcell.ColorWhite, Background: tcell.ColorBlue},
				tz.NewBox(
					tz.Style{Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Simple Modal"),
					tz.NewText(tz.Style{Color: tcell.ColorWhite}, "This is a simple message."),
					tz.NewButton(
						tz.Style{ID: "close_simple", Focusable: true, Color: tcell.ColorWhite},
						"Close",
						func() { setSimpleOpen(false) },
					),
				),
				simpleOpen,
			),

			// Confirmation Dialog
			tz.NewModal(
				ctx,
				tz.Style{Color: tcell.ColorWhite, Background: tcell.ColorGreen},
				tz.NewBox(
					tz.Style{Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Are you sure?"),
					tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Do you want to proceed?"),
					tz.NewBox(
						tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
						tz.NewButton(
							tz.Style{ID: "confirm_yes", Focusable: true, Color: tcell.ColorWhite},
							"Yes",
							func() { setConfirmOpen(false) },
						),
						tz.NewButton(
							tz.Style{ID: "confirm_no", Focusable: true, Color: tcell.ColorWhite},
							"No",
							func() { setConfirmOpen(false) },
						),
					),
				),
				confirmOpen,
			),

			// Form Modal
			tz.NewModal(
				ctx,
				tz.Style{Color: tcell.ColorWhite, Background: tcell.ColorDarkMagenta},
				tz.NewBox(
					tz.Style{Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Interactive Form"),
					tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Tab between fields:"),

					tz.NewBox(
						tz.Style{Margin: tz.Margin{Top: 1, Bottom: 1}},
						tz.NewButton(tz.Style{ID: "form_opt1", Focusable: true, Color: tcell.ColorWhite}, "Option 1", nil),
						tz.NewButton(tz.Style{ID: "form_opt2", Focusable: true, Color: tcell.ColorWhite}, "Option 2", nil),
					),

					tz.NewBox(
						tz.Style{FlexDirection: "row"},
						tz.NewButton(
							tz.Style{ID: "form_submit", Focusable: true, Color: tcell.ColorWhite},
							"Submit",
							func() { setFormOpen(false) },
						),
						tz.NewButton(
							tz.Style{ID: "form_cancel", Focusable: true, Color: tcell.ColorWhite},
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
