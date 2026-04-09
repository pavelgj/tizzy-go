package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

type GroceryItem struct {
	Name        string
	Description string
}

func main() {
	app, err := tz.NewApp()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating app: %v\n", err)
		os.Exit(1)
	}

	items := []any{
		GroceryItem{Name: "Pocky", Description: "Expensive"},
		GroceryItem{Name: "Ginger", Description: "Exquisite"},
		GroceryItem{Name: "Plantains", Description: "Questionable"},
		GroceryItem{Name: "Honey Dew", Description: "Delectable"},
		GroceryItem{Name: "Pineapple", Description: "Kind of spicy"},
		GroceryItem{Name: "Snow Peas", Description: "Bold flavor"},
		GroceryItem{Name: "Party Gherkin", Description: "My favorite"},
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		selectedIndex, setSelectedIndex := tz.UseState(ctx, -1)

		list := tz.NewList(
			ctx,
			tz.Style{
				Width:     40,
				Height:    16,
				Border:    true,
				Title:     "Groceries",
				Focusable: true,
			},
			"grocery-list",
			items,
			selectedIndex,
			func(item any, index int, selected bool, cursor bool) tz.Node {
				g := item.(GroceryItem)
				
				bg := tcell.ColorBlack
				if cursor {
					bg = tcell.ColorDarkCyan
				}
				if selected {
					bg = tcell.ColorBlue
				}
				
				titleColor := tcell.ColorWhite
				descColor := tcell.ColorGray
				
				if cursor || selected {
					titleColor = tcell.ColorWhite
					descColor = tcell.ColorLightGray
				}

				return tz.NewBox(
					tz.Style{
						FlexDirection: "column",
						Background:    bg,
						FillWidth:     true,
					},
					tz.NewText(tz.Style{Color: titleColor, Background: bg}, g.Name),
					tz.NewText(tz.Style{Color: descColor, Background: bg}, g.Description),
				)
			},
			func(idx int) {
				setSelectedIndex(idx)
			},
		)
		list.ItemHeight = 2

		return tz.NewBox(
			tz.Style{
				Width:          50,
				Height:         20,
				JustifyContent: "center",
				FlexDirection:  "column",
			},
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Use Arrows to navigate, Enter to select"),
			list,
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Selected Index: "+strconv.Itoa(selectedIndex)),
		)
	}

	_ = app.Run(render, func(ev tcell.Event) {})
}
