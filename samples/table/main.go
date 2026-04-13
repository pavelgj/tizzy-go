package main

import (
	"log"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	headers := []string{"ID", "Name", "Role", "Status"}
	rows := [][]string{
		{"1", "Alice", "Developer", "Active"},
		{"2", "Bob", "Designer", "Away"},
		{"3", "Charlie", "Manager", "Busy"},
		{"4", "David", "DevOps", "Offline"},
	}

	numericHeaders := []string{"#", "Item", "Qty", "Price"}
	numericRows := [][]string{
		{"1", "Widget", "100", "$9.99"},
		{"2", "Gadget", "50", "$24.99"},
		{"3", "Doohickey", "200", "$4.49"},
		{"4", "Thingamajig", "75", "$14.99"},
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		return tz.NewBox(
			tz.Style{
				FlexDirection: "column",
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				Background:    tcell.ColorReset,
			},
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Plain table:"),
			tz.NewTable(tz.Style{Color: tcell.ColorWhite}, headers, rows),

			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "With border and dividers:"),
			func() tz.Node {
				t := tz.NewTable(tz.Style{Color: tcell.ColorWhite, Border: true}, headers, rows)
				t.Dividers = true
				return t
			}(),

			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Zebra striped, fill width:"),
			func() tz.Node {
				t := tz.NewTable(tz.Style{
					Color:     tcell.ColorWhite,
					Border:    true,
					FillWidth: true,
					Title:     "Employees",
				}, headers, rows)
				t.StripeBackground = tcell.NewRGBColor(40, 40, 60)
				t.Dividers = true
				return t
			}(),

			tz.NewText(tz.Style{Color: tcell.ColorWhite}, ""),
			tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Column alignment (right-aligned Qty and Price):"),
			func() tz.Node {
				t := tz.NewTable(tz.Style{
					Color:     tcell.ColorWhite,
					Border:    true,
					FillWidth: true,
				}, numericHeaders, numericRows)
				t.ColAligns = []string{"right", "left", "right", "right"}
				t.StripeBackground = tcell.NewRGBColor(30, 40, 30)
				return t
			}(),
		)
	}

	update := func(ev tcell.Event) {}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
