package main

import (
	"log"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/pavelgj/tizzy-go/tz"
)

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		return tz.NewBox(
			tz.Style{
				FillWidth:     true,
				FillHeight:    true,
				Color:         tcell.ColorWhite,
				Background:    tcell.ColorBlack,
				FlexDirection: "column",
			},
			// Header
			tz.NewBox(
				tz.Style{
					Height:        1,
					Background:    tcell.ColorNavy,
					Color:         tcell.ColorWhite,
					FlexDirection: "row",
					Padding:       tz.Padding{Left: 1, Right: 1},
					FillWidth:     true,
				},
				tz.NewText(tz.Style{}, "TIZZY DASHBOARD | "+time.Now().Format("15:04:05")),
			),
			// Main Body
			tz.NewBox(
				tz.Style{
					FillWidth:     true,
					FillHeight:    true,
					FlexDirection: "row",
					Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
				},
				// Sidebar
				tz.NewList(ctx, tz.Style{
					Width:         15,
					Border:        true,
					Title:         "System",
					Color:         tcell.ColorDarkCyan,
					Margin:        tz.Margin{Right: 1},
					Focusable:     true,
				}, "sidebar-list", []any{"Overview", "Processes", "Network", "Logs"}, 0, func(item any, index int, selected bool, cursor bool) tz.Node {
					label := item.(string)
					style := tz.Style{FillWidth: true}
					textColor := tcell.ColorWhite
					if selected {
						textColor = tcell.ColorYellow
					}
					if cursor {
						style.Background = tcell.ColorNavy
					}
					return tz.NewBox(style, tz.NewText(tz.Style{Color: textColor, Background: style.Background}, "  "+label))
				}, func(idx int) {}),
				// Center Content
				tz.NewBox(
					tz.Style{
						FlexDirection: "column",
						Margin:        tz.Margin{Right: 1},
					},
					// Metrics
					tz.NewBox(
						tz.Style{
							Border:        true,
							Title:         "Metrics",
							Color:         tcell.ColorGreen,
							Padding:       tz.Padding{Left: 1, Right: 1},
							FlexDirection: "column",
						},
						tz.NewText(tz.Style{Color: tcell.ColorWhite}, "CPU Usage:"),
						tz.NewProgressBar(tz.Style{Color: tcell.ColorDarkCyan, Width: 30}, 0.45),
						tz.NewText(tz.Style{Color: tcell.ColorWhite, Margin: tz.Margin{Top: 1}}, "Memory Usage:"),
						tz.NewProgressBar(tz.Style{Color: tcell.ColorBlue, Width: 30}, 0.72),
					),
					// Table
					tz.NewBox(
						tz.Style{
							Border:    true,
							Title:     "Processes",
							Color:     tcell.ColorYellow,
							Margin:    tz.Margin{Top: 1},
						},
						tz.NewTable(
							tz.Style{Color: tcell.ColorWhite},
							[]string{"PID", "Name", "CPU", "MEM"},
							[][]string{
								{"1024", "tizzy", "12%", "45MB"},
								{"2048", "dockerd", "5%", "120MB"},
								{"4096", "chrome", "25%", "450MB"},
							},
						),
					),
				),
				// Right Content (Form & Logs)
				tz.NewBox(
					tz.Style{
						Width:         30,
						FlexDirection: "column",
						FillWidth:     true,
					},
					// Form
					tz.NewBox(
						tz.Style{
							Border:        true,
							Title:         "Quick Action",
							Color:         tcell.ColorBlue,
							Padding:       tz.Padding{Left: 1, Right: 1},
							FlexDirection: "column",
							Margin:        tz.Margin{Bottom: 1},
							FillWidth:     true,
						},
						tz.NewText(tz.Style{Color: tcell.ColorWhite}, "Command:"),
						tz.NewTextInput(ctx, tz.Style{Focusable: true, Border: true, Width: 20}, "reboot", nil),
						tz.NewCheckbox(ctx, tz.Style{Focusable: true, Margin: tz.Margin{Top: 1}}, "Force", false, nil),
						tz.NewButton(tz.Style{Focusable: true, Border: true, Margin: tz.Margin{Top: 1}}, " Execute ", nil),
					),
					// Logs
					tz.NewBox(
						tz.Style{
							Border:        true,
							Title:         "Logs",
							Color:         tcell.ColorRed,
							FlexDirection: "column",
							Padding:       tz.Padding{Left: 1, Right: 1},
							FillHeight:    true,
							FillWidth:     true,
						},
						tz.NewText(tz.Style{Color: tcell.ColorGreen}, "[INFO] System up"),
						tz.NewText(tz.Style{Color: tcell.ColorYellow}, "[WARN] High load"),
						tz.NewText(tz.Style{Color: tcell.ColorRed}, "[ERR] Disk full"),
						// Spinner
						tz.NewBox(
							tz.Style{FlexDirection: "row", Margin: tz.Margin{Top: 1}},
							tz.NewSpinner(ctx, tz.Style{Color: tcell.ColorDarkCyan}),
							tz.NewText(tz.Style{Color: tcell.ColorWhite, Margin: tz.Margin{Left: 1}}, "Syncing..."),
						),
					),
				),
			),
		)
	}

	if err := app.Run(render, nil); err != nil {
		log.Fatal(err)
	}
}
