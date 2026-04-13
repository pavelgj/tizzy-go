package main

import (
	"fmt"
	"log"

	"github.com/pavelgj/tizzy-go/tz"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	err = app.Run(
		func(ctx *tz.RenderContext) tz.Node {
			count, setCount := tz.UseState(ctx, 0)

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{TextAttrs: tcell.AttrBold}, "Button Sample"),
				tz.NewText(tz.Style{}, "Press Tab to focus, Enter/Space to click."),
				tz.NewText(tz.Style{}, ""),

				// Default style
				tz.NewText(tz.Style{}, "Default:"),
				tz.NewButton(
					tz.Style{ID: "btn-default", Focusable: true},
					"Click Me",
					func() { setCount(count + 1) },
				),
				tz.NewText(tz.Style{}, ""),

				// With border
				tz.NewText(tz.Style{}, "With border:"),
				tz.NewButton(
					tz.Style{
						ID:        "btn-border",
						Focusable: true,
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
					},
					"Click Me",
					func() { setCount(count + 1) },
				),
				tz.NewText(tz.Style{}, ""),

				// Custom focus colors
				tz.NewText(tz.Style{}, "Custom focus colors (blue bg):"),
				tz.NewButton(
					tz.Style{
						ID:              "btn-custom",
						Focusable:       true,
						Border:          true,
						Padding:         tz.Padding{Left: 1, Right: 1},
						FocusColor:      tcell.ColorWhite,
						FocusBackground: tcell.NewRGBColor(0, 80, 200),
					},
					"Focus Me",
					func() { setCount(count + 1) },
				),
				tz.NewText(tz.Style{}, ""),

				// Bold label via TextAttrs
				tz.NewText(tz.Style{}, "Bold label via TextAttrs:"),
				tz.NewButton(
					tz.Style{
						ID:        "btn-bold",
						Focusable: true,
						TextAttrs: tcell.AttrBold,
					},
					"Bold Button",
					func() { setCount(count + 1) },
				),
				tz.NewText(tz.Style{}, ""),

				// Fixed width — label is centered
				tz.NewText(tz.Style{}, "Fixed width (centered label):"),
				tz.NewButton(
					tz.Style{
						ID:        "btn-wide",
						Focusable: true,
						Border:    true,
						Width:     30,
					},
					"Centered",
					func() { setCount(count + 1) },
				),
				tz.NewText(tz.Style{}, ""),

				// Disabled
				tz.NewText(tz.Style{}, "Disabled (not focusable, dimmed):"),
				tz.NewButton(
					tz.Style{
						ID:        "btn-disabled",
						Focusable: true, // ignored when Disabled=true
						Border:    true,
						Padding:   tz.Padding{Left: 1, Right: 1},
					},
					"Can't Click",
					func() { setCount(count + 1) },
				).WithDisabled(true),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{Color: tcell.ColorGreen}, fmt.Sprintf("Clicked %d times", count)),
			)
		},
		func(ev tcell.Event) {},
	)

	if err != nil {
		log.Fatal(err)
	}
}
