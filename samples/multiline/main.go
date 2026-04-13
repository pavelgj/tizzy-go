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
			content, setContent := tz.UseState(ctx,
				"Line 1\nLine 2\nLine 3\nLine 4\nLine 5\nLine 6\nLine 7")
			notes, setNotes := tz.UseState(ctx, "")

			// Main editor: scroll indicators (↑ ↓) appear in the border when
			// content extends beyond the visible window.
			editor := tz.NewTextInput(ctx,
				tz.Style{
					ID:        "editor",
					Focusable: true,
					Border:    true,
					Title:     " Editor ",
					Padding:   tz.Padding{Left: 1, Right: 1},
					Width:     30,
					Multiline: true,
					MaxHeight: 7,
				},
				content, func(v string) { setContent(v) })

			// Placeholder: shown in gray when the textarea is empty.
			notesInput := tz.NewTextInput(ctx,
				tz.Style{
					ID:        "notes",
					Focusable: true,
					Border:    true,
					Title:     " Notes ",
					Padding:   tz.Padding{Left: 1, Right: 1},
					Width:     30,
					Multiline: true,
					MaxHeight: 4,
				},
				notes, func(v string) { setNotes(v) })
			notesInput.Placeholder = "Start typing notes…"

			lines := 0
			for _, c := range content {
				if c == '\n' {
					lines++
				}
			}
			lines++

			return tz.NewBox(
				tz.Style{FlexDirection: "column", Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
				tz.NewText(tz.Style{TextAttrs: tcell.AttrBold}, "Multiline Text Input Sample"),
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					"Arrows: navigate  |  Home/End: line ↔ absolute  |  PgUp/PgDn: scroll"),
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					"Ctrl+Left/Right: word jump  |  Ctrl+W: delete word  |  Tab: focus"),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "Editor (↑↓ in border = hidden lines):"),
				editor,
				tz.NewText(tz.Style{Color: tcell.ColorGray},
					fmt.Sprintf("%d lines  |  %d chars", lines, len([]rune(content)))),
				tz.NewText(tz.Style{}, ""),

				tz.NewText(tz.Style{}, "With placeholder:"),
				notesInput,
			)
		},
		func(ev tcell.Event) {},
	)

	if err != nil {
		log.Fatal(err)
	}
}
