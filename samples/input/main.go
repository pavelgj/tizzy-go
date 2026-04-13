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
			name, setName := tz.UseState(ctx, "")
			password, setPassword := tz.UseState(ctx, "")
			limited, setLimited := tz.UseState(ctx, "")
			custom, setCustom := tz.UseState(ctx, "")
			bold, setBold := tz.UseState(ctx, "")
			titled, setTitled := tz.UseState(ctx, "")
			scrollable, setScrollable := tz.UseState(ctx, "Try typing a long line here to see the scroll indicators appear")

			// Placeholder: shown in dim gray when value is empty.
			nameInput := tz.NewTextInput(ctx,
				tz.Style{ID: "name", Focusable: true, Border: true,
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28},
				name, func(v string) { setName(v) })
			nameInput.Placeholder = "Enter your name…"

			// Mask: actual characters are hidden, replaced by the mask rune.
			passInput := tz.NewTextInput(ctx,
				tz.Style{ID: "pass", Focusable: true, Border: true, Title: " Password ",
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28},
				password, func(v string) { setPassword(v) })
			passInput.Placeholder = "Enter password…"
			passInput.Mask = '*'

			// MaxLen: rejects input beyond the limit.
			limitedInput := tz.NewTextInput(ctx,
				tz.Style{ID: "limited", Focusable: true, Border: true,
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28},
				limited, func(v string) { setLimited(v) })
			limitedInput.Placeholder = "Max 10 characters"
			limitedInput.MaxLen = 10

			// Disabled: visually dimmed, not focusable, ignores all input.
			disabledInput := tz.NewTextInput(ctx,
				tz.Style{ID: "disabled", Focusable: true, Border: true,
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28},
				"Read-only value", nil)
			disabledInput.Disabled = true

			// Custom focus colors: override the default yellow.
			customInput := tz.NewTextInput(ctx,
				tz.Style{ID: "custom", Focusable: true, Border: true,
					Padding:    tz.Padding{Left: 1, Right: 1}, Width: 28,
					FocusColor: tcell.NewRGBColor(0, 200, 100),
				},
				custom, func(v string) { setCustom(v) })
			customInput.Placeholder = "Focus me (green border)"

			// TextAttrs: styling applied to the text itself.
			boldInput := tz.NewTextInput(ctx,
				tz.Style{ID: "bold", Focusable: true, Border: true,
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28,
					TextAttrs: tcell.AttrBold},
				bold, func(v string) { setBold(v) })
			boldInput.Placeholder = "Bold text input"

			// Border title: label embedded in the top border.
			titledInput := tz.NewTextInput(ctx,
				tz.Style{ID: "titled", Focusable: true, Border: true, Title: " Search ",
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 28},
				titled, func(v string) { setTitled(v) })
			titledInput.Placeholder = "Type to search…"

			// Scroll indicators: narrow input — '<' and '>' appear in the border
			// when content is scrolled. Try typing a long line.
			scrollInput := tz.NewTextInput(ctx,
				tz.Style{ID: "scroll", Focusable: true, Border: true,
					Padding: tz.Padding{Left: 1, Right: 1}, Width: 20},
				scrollable, func(v string) { setScrollable(v) })

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{TextAttrs: tcell.AttrBold}, "Text Input Sample"),
				tz.NewText(tz.Style{Color: tcell.ColorGray}, "Ctrl+Left/Right: word jump  |  Ctrl+W: delete word  |  Tab: focus"),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Placeholder (shown when empty):"),
				nameInput,
				tz.NewText(tz.Style{Color: tcell.ColorGray}, "Value: "+name),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Password (Mask: '*'):"),
				passInput,
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					fmt.Sprintf("Length: %d chars", len([]rune(password)))),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "MaxLen = 10:"),
				limitedInput,
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					fmt.Sprintf("%d / 10", len([]rune(limited)))),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Disabled (not focusable, dimmed):"),
				disabledInput,
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Custom focus color (green border):"),
				customInput,
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Bold TextAttrs:"),
				boldInput,
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Border Title:"),
				titledInput,
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Scroll indicators (narrow — overflow shows '<' '>'):"),
				scrollInput,
			)
		},
		func(ev tcell.Event) {},
	)

	if err != nil {
		log.Fatal(err)
	}
}
