package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		// State for various components
		count, setCount := tz.UseState(ctx, 0)
		checked, setChecked := tz.UseState(ctx, false)
		radioVal, setRadioVal := tz.UseState(ctx, "opt1")
		dropdownIdx, setDropdownIdx := tz.UseState(ctx, 0)
		textVal, setTextVal := tz.UseState(ctx, "Edit me!")
		lastAction, setLastAction := tz.UseState(ctx, "None")

		dropdownOptions := []string{"Option A", "Option B", "Option C"}

		return tz.NewBox(
			tz.Style{
				FillWidth:  true,
				FillHeight: true,
				Color:      tcell.ColorWhite,
				Background: tcell.ColorBlack,
				FlexDirection: "column",
			},
			// MenuBar at the top
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
							{Label: "New", Action: func() { setLastAction("New File") }},
							{Label: "Save", Action: func() { setLastAction("Save File") }},
							{Label: "Exit", Action: func() { app.Stop() }},
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
			// Main content area with sidebar and content
			tz.NewBox(
				tz.Style{
					FlexDirection: "row",
					FillWidth:     true,
					FillHeight:    true,
					Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
				},
				// Sidebar
				tz.NewBox(
					tz.Style{
						Width:         15,
						Border:        true,
						Title:         "Sidebar",
						FlexDirection: "column",
						Margin:        tz.Margin{Right: 1},
					},
					tz.NewText(tz.Style{Color: tcell.ColorYellow}, " Navigation"),
					tz.NewText(tz.Style{Color: tcell.ColorWhite}, " Dashboard"),
					tz.NewText(tz.Style{Color: tcell.ColorGray}, " Settings"),
					tz.NewText(tz.Style{Color: tcell.ColorGray}, " Profile"),
				),
				// Main Content (Tabs)
				tz.NewBox(
					tz.Style{
						FillWidth:     true,
						FillHeight:    true,
						FlexDirection: "column",
					},
					tz.NewTabs(
						ctx,
						tz.Style{
							Focusable: true,
							Color:     tcell.ColorWhite,
						},
						[]tz.Tab{
							{
								Label: "Form Controls",
								Content: tz.NewBox(
									tz.Style{Border: true, Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}, FlexDirection: "column"},
									tz.NewText(tz.Style{Color: tcell.ColorYellow}, "Interactive Components"),
									
									// TextInput
									tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Text Input:"),
									tz.NewTextInput(ctx, tz.Style{Focusable: true, Border: true, Width: 20}, textVal, func(v string) {
										setTextVal(v)
									}),
									
									// Checkbox
									tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Checkbox:"),
									tz.NewCheckbox(ctx, tz.Style{Focusable: true}, "Enable feature", checked, func(v bool) {
										setChecked(v)
									}),
									
									// RadioButtons
									tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Radio Buttons:"),
									tz.NewRadioButton(ctx, tz.Style{Focusable: true}, "Option 1", "opt1", radioVal == "opt1", func(v string) {
										setRadioVal(v)
									}),
									tz.NewRadioButton(ctx, tz.Style{Focusable: true}, "Option 2", "opt2", radioVal == "opt2", func(v string) {
										setRadioVal(v)
									}),
									
									// Dropdown
									tz.NewText(tz.Style{Color: tcell.ColorGray, Margin: tz.Margin{Top: 1}}, "Dropdown:"),
									tz.NewDropdown(ctx, tz.Style{Focusable: true, Border: true, Width: 15}, dropdownOptions, dropdownIdx, func(idx int) {
										setDropdownIdx(idx)
									}),
								),
							},
							{
								Label: "Buttons & State",
								Content: tz.NewBox(
									tz.Style{Border: true, Padding: tz.Padding{Top: 1, Left: 2, Bottom: 1, Right: 2}, FlexDirection: "column"},
									tz.NewText(tz.Style{Color: tcell.ColorYellow}, "State Management"),
									
									tz.NewText(tz.Style{Color: tcell.ColorWhite, Margin: tz.Margin{Top: 1}}, "Click the buttons to change state:"),
									
									tz.NewBox(
										tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
										tz.NewButton(tz.Style{Focusable: true, Border: true, ID: "btn_inc"}, " Increment ", func() {
											setCount(count + 1)
										}),
										tz.NewButton(tz.Style{Focusable: true, Border: true, ID: "btn_dec", Margin: tz.Margin{Left: 1}}, " Decrement ", func() {
											setCount(count - 1)
										}),
									),
									
									tz.NewText(tz.Style{Color: tcell.ColorGreen, Margin: tz.Margin{Top: 1}}, "Counter: "+strconv.Itoa(count)),
									tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Checkbox: "+fmt.Sprintf("%t", checked)),
									tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Radio: "+radioVal),
									tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Dropdown: "+dropdownOptions[dropdownIdx]),
									tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Text: "+textVal),
									tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Last Action: "+lastAction),
								),
							},
						},
					),
				),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		log.Fatal(err)
	}
}
