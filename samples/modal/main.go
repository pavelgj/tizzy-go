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

	// Initial states
	app.SetState("modal_simple", &splotch.ModalState{Open: false})
	app.SetState("modal_confirm", &splotch.ModalState{Open: false})
	app.SetState("modal_form", &splotch.ModalState{Open: false})

	render := func(ctx *splotch.RenderContext) splotch.Node {
		var stateSimple *splotch.ModalState
		if s, ok := app.GetState("modal_simple").(*splotch.ModalState); ok {
			stateSimple = s
		}
		var stateConfirm *splotch.ModalState
		if s, ok := app.GetState("modal_confirm").(*splotch.ModalState); ok {
			stateConfirm = s
		}
		var stateForm *splotch.ModalState
		if s, ok := app.GetState("modal_form").(*splotch.ModalState); ok {
			stateForm = s
		}

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
					func() {
						if stateSimple != nil {
							stateSimple.Open = true
						}
					},
				),
				splotch.NewButton(
					splotch.Style{ID: "btn_confirm", Focusable: true, Color: tcell.ColorWhite},
					"Confirmation",
					func() {
						if stateConfirm != nil {
							stateConfirm.Open = true
						}
					},
				),
				splotch.NewButton(
					splotch.Style{ID: "btn_form", Focusable: true, Color: tcell.ColorWhite},
					"Form Modal",
					func() {
						if stateForm != nil {
							stateForm.Open = true
						}
					},
				),
			),

			// Simple Modal
			splotch.NewModal(
				splotch.Style{ID: "modal_simple", Color: tcell.ColorWhite, Background: tcell.ColorBlue},
				splotch.NewBox(
					splotch.Style{Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Simple Modal"),
					splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "This is a simple message."),
					splotch.NewButton(
						splotch.Style{ID: "close_simple", Focusable: true, Color: tcell.ColorWhite},
						"Close",
						func() {
							if stateSimple != nil {
								stateSimple.Open = false
							}
						},
					),
				),
			),

			// Confirmation Dialog
			splotch.NewModal(
				splotch.Style{ID: "modal_confirm", Color: tcell.ColorWhite, Background: tcell.ColorGreen},
				splotch.NewBox(
					splotch.Style{Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
					splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, "Are you sure?"),
					splotch.NewText(splotch.Style{Color: tcell.ColorWhite}, "Do you want to proceed?"),
					splotch.NewBox(
						splotch.Style{FlexDirection: "row", Margin: splotch.Margin{Top: 1}},
						splotch.NewButton(
							splotch.Style{ID: "confirm_yes", Focusable: true, Color: tcell.ColorWhite},
							"Yes",
							func() {
								if stateConfirm != nil {
									stateConfirm.Open = false
								}
							},
						),
						splotch.NewButton(
							splotch.Style{ID: "confirm_no", Focusable: true, Color: tcell.ColorWhite},
							"No",
							func() {
								if stateConfirm != nil {
									stateConfirm.Open = false
								}
							},
						),
					),
				),
			),

			// Form Modal
			splotch.NewModal(
				splotch.Style{ID: "modal_form", Color: tcell.ColorWhite, Background: tcell.ColorDarkMagenta},
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
							func() {
								if stateForm != nil {
									stateForm.Open = false
								}
							},
						),
						splotch.NewButton(
							splotch.Style{ID: "form_cancel", Focusable: true, Color: tcell.ColorWhite},
							"Cancel",
							func() {
								if stateForm != nil {
									stateForm.Open = false
								}
							},
						),
					),
				),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		fmt.Println(err)
	}
}
