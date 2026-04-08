package main

import (
	"fmt"
	"log"
	"splotch/splotch"
	"strings"

	"github.com/gdamore/tcell/v2"
)

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	initialHistory := []string{
		"ls -la",
		"git status",
		"go build ./...",
		"help @feature",
	}

	allSuggestions := []string{
		"feature-x",
		"bugfix-y",
		"documentation",
		"tests",
		"refactor",
	}

	var setInputValue func(string)
	var setHistoryIndex func(int)
	var setPopupOpen func(bool)
	var setSelectedSug func(int)
	var setCursorOverride func(*int)

	var inputValue string
	var history []string
	var historyIndex int
	var popupOpen bool
	var selectedSug int
	var filteredSuggestions []string
	var cursorOverride *int

	render := func(ctx *splotch.RenderContext) splotch.Node {
		inputValue, setInputValue = splotch.UseState(ctx, "")

		history, _ = splotch.UseState(ctx, initialHistory)

		historyIndex, setHistoryIndex = splotch.UseState(ctx, -1)

		popupOpen, setPopupOpen = splotch.UseState(ctx, false)

		selectedSug, setSelectedSug = splotch.UseState(ctx, 0)

		cursorOverride, setCursorOverride = splotch.UseState[*int](ctx, nil)

		filteredSuggestions = []string{}
		if popupOpen {
			filteredSuggestions = allSuggestions
		}

		inputStyle := splotch.Style{
			Focusable: true,
			Border:    true,
			Padding:   splotch.Padding{Left: 1, Right: 1},
			Width:     50,
			Multiline: true,
			MaxHeight: 5,
		}

		inputNode := splotch.NewTextInput(
			ctx,
			inputStyle,
			inputValue,
			func(newValue string) {
				setInputValue(newValue)

				if strings.HasSuffix(newValue, "@") || strings.HasSuffix(newValue, "/") {
					setPopupOpen(true)
					setSelectedSug(0)
				} else if popupOpen && !strings.Contains(newValue, "@") && !strings.Contains(newValue, "/") {
					setPopupOpen(false)
				}
			},
		)

		// Apply cursor override
		inputNode.Cursor = cursorOverride

		// Reset override for next render if it was set
		if cursorOverride != nil {
			setCursorOverride(nil)
		}

		var popupNode splotch.Node
		if popupOpen && len(filteredSuggestions) > 0 {
			listItems := []splotch.Node{}
			for i, sug := range filteredSuggestions {
				style := splotch.Style{Padding: splotch.Padding{Left: 1, Right: 1}}
				if i == selectedSug {
					style.Background = tcell.ColorYellow
					style.Color = tcell.ColorBlack
				}
				listItems = append(listItems, splotch.NewText(style, sug))
			}

			popupNode = splotch.NewPopup(
				ctx,
				splotch.Style{
					Border:     true,
					Background: tcell.ColorGray,
					Width:      20,
				},
				splotch.NewBox(
					splotch.Style{FlexDirection: "column"},
					listItems...,
				),
				10, // X
				5,  // Y
				popupOpen,
			)
		}

		return splotch.NewBox(
			splotch.Style{FlexDirection: "column", Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2}},
			splotch.NewText(splotch.Style{}, "Advanced Prompt Box Sample"),
			splotch.NewText(splotch.Style{}, "Type @ or / to trigger popup. Use Up/Down to navigate history or popup."),
			splotch.NewText(splotch.Style{}, fmt.Sprintf("History Index: %d", historyIndex)),
			inputNode,
			popupNode,
			splotch.NewText(splotch.Style{}, "Press Ctrl+C to exit."),
		)
	}

	update := func(ev tcell.Event) {
		if evKey, ok := ev.(*tcell.EventKey); ok {
			if popupOpen {
				if evKey.Key() == tcell.KeyDown {
					setSelectedSug((selectedSug + 1) % len(filteredSuggestions))
				} else if evKey.Key() == tcell.KeyUp {
					setSelectedSug((selectedSug - 1 + len(filteredSuggestions)) % len(filteredSuggestions))
				} else if evKey.Key() == tcell.KeyEnter {
					sug := filteredSuggestions[selectedSug]
					newVal := inputValue + sug
					setInputValue(newVal)
					setPopupOpen(false)

					// Set cursor override to end of new text
					newOffset := len(newVal)
					setCursorOverride(&newOffset)
				} else if evKey.Key() == tcell.KeyEscape {
					setPopupOpen(false)
				}
			} else {
				if evKey.Key() == tcell.KeyUp {
					if historyIndex < len(history)-1 {
						newIdx := historyIndex + 1
						setHistoryIndex(newIdx)
						setInputValue(history[newIdx])

						// Move cursor to end of history item
						newOffset := len(history[newIdx])
						setCursorOverride(&newOffset)
					}
				} else if evKey.Key() == tcell.KeyDown {
					if historyIndex >= 0 {
						newIdx := historyIndex - 1
						setHistoryIndex(newIdx)
						if newIdx >= 0 {
							setInputValue(history[newIdx])
							newOffset := len(history[newIdx])
							setCursorOverride(&newOffset)
						} else {
							setInputValue("")
							newOffset := 0
							setCursorOverride(&newOffset)
						}
					}
				}
			}
		}
	}

	if err := app.Run(render, update); err != nil {
		log.Fatal(err)
	}
}
