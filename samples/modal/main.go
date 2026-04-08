package main

import (
	"fmt"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		fmt.Println(err)
		return
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		simpleOpen, setSimpleOpen := tizzy.UseState(ctx, false)
		confirmOpen, setConfirmOpen := tizzy.UseState(ctx, false)
		formOpen, setFormOpen := tizzy.UseState(ctx, false)

		return tizzy.NewBox(
			tizzy.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
				Padding:    tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "Modal / Dialog Samples"),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "Press buttons to open different modals."),
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGray}, "------------------------------------------"),

			tizzy.NewBox(
				tizzy.Style{FlexDirection: "row"},
				tizzy.NewButton(
					tizzy.Style{ID: "btn_simple", Focusable: true, Color: tcell.ColorWhite},
					"Simple Modal",
					func() { setSimpleOpen(true) },
				),
				tizzy.NewButton(
					tizzy.Style{ID: "btn_confirm", Focusable: true, Color: tcell.ColorWhite},
					"Confirmation",
					func() { setConfirmOpen(true) },
				),
				tizzy.NewButton(
					tizzy.Style{ID: "btn_form", Focusable: true, Color: tcell.ColorWhite},
					"Form Modal",
					func() { setFormOpen(true) },
				),
			),

			// Simple Modal
			tizzy.NewModal(
				ctx,
				tizzy.Style{Color: tcell.ColorWhite, Background: tcell.ColorBlue},
				tizzy.NewBox(
					tizzy.Style{Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "Simple Modal"),
					tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "This is a simple message."),
					tizzy.NewButton(
						tizzy.Style{ID: "close_simple", Focusable: true, Color: tcell.ColorWhite},
						"Close",
						func() { setSimpleOpen(false) },
					),
				),
				simpleOpen,
			),

			// Confirmation Dialog
			tizzy.NewModal(
				ctx,
				tizzy.Style{Color: tcell.ColorWhite, Background: tcell.ColorGreen},
				tizzy.NewBox(
					tizzy.Style{Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "Are you sure?"),
					tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Do you want to proceed?"),
					tizzy.NewBox(
						tizzy.Style{FlexDirection: "row", Margin: tizzy.Margin{Top: 1}},
						tizzy.NewButton(
							tizzy.Style{ID: "confirm_yes", Focusable: true, Color: tcell.ColorWhite},
							"Yes",
							func() { setConfirmOpen(false) },
						),
						tizzy.NewButton(
							tizzy.Style{ID: "confirm_no", Focusable: true, Color: tcell.ColorWhite},
							"No",
							func() { setConfirmOpen(false) },
						),
					),
				),
				confirmOpen,
			),

			// Form Modal
			tizzy.NewModal(
				ctx,
				tizzy.Style{Color: tcell.ColorWhite, Background: tcell.ColorDarkMagenta},
				tizzy.NewBox(
					tizzy.Style{Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, "Interactive Form"),
					tizzy.NewText(tizzy.Style{Color: tcell.ColorWhite}, "Tab between fields:"),

					tizzy.NewBox(
						tizzy.Style{Margin: tizzy.Margin{Top: 1, Bottom: 1}},
						tizzy.NewButton(tizzy.Style{ID: "form_opt1", Focusable: true, Color: tcell.ColorWhite}, "Option 1", nil),
						tizzy.NewButton(tizzy.Style{ID: "form_opt2", Focusable: true, Color: tcell.ColorWhite}, "Option 2", nil),
					),

					tizzy.NewBox(
						tizzy.Style{FlexDirection: "row"},
						tizzy.NewButton(
							tizzy.Style{ID: "form_submit", Focusable: true, Color: tcell.ColorWhite},
							"Submit",
							func() { setFormOpen(false) },
						),
						tizzy.NewButton(
							tizzy.Style{ID: "form_cancel", Focusable: true, Color: tcell.ColorWhite},
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
