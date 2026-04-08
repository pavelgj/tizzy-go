package main

import (
	"fmt"
	"log"
	"tizzy/tizzy"

	"github.com/gdamore/tcell/v2"
)

// CounterComponent is a simple stateful component
type CounterComponent struct {
	id    string
	count int
}

func NewCounterComponent(id string) *CounterComponent {
	return &CounterComponent{id: id}
}

func (c *CounterComponent) Render() tizzy.Node {
	return tizzy.NewBox(
		tizzy.Style{
			Border:  true,
			Padding: tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:  tizzy.Margin{Top: 1},
		},
		tizzy.NewText(tizzy.Style{Color: tcell.ColorYellow}, fmt.Sprintf("Counter [%s]", c.id)),
		tizzy.NewText(tizzy.Style{}, fmt.Sprintf("Clicks: %d", c.count)),
		tizzy.NewButton(
			tizzy.Style{ID: c.id + "_btn", Focusable: true, Margin: tizzy.Margin{Top: 1}},
			"Increment",
			func() {
				c.count++
			},
		),
	)
}

// SidebarComponent holds state for selected tab
type SidebarComponent struct {
	selectedTab int
	onSelect    func(int)
}

func (s *SidebarComponent) Render() tizzy.Node {
	items := []string{"Dashboard", "Settings", "About"}
	var children []tizzy.Node
	children = append(children, tizzy.NewText(tizzy.Style{Color: tcell.ColorLightCyan}, " NAVIGATION "))

	for i, item := range items {
		style := tizzy.Style{
			Focusable: true,
			ID:        fmt.Sprintf("side_%d", i),
			Margin:    tizzy.Margin{Top: 1},
		}
		label := "  " + item
		if i == s.selectedTab {
			label = "> " + item
			style.Color = tcell.ColorYellow
		}

		idx := i // capture for closure
		children = append(children, tizzy.NewButton(style, label, func() {
			s.selectedTab = idx
			if s.onSelect != nil {
				s.onSelect(idx)
			}
		}))
	}

	return tizzy.NewBox(
		tizzy.Style{
			Border:        true,
			Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
			FlexDirection: "column",
			Width:         20,
		},
		children...,
	)
}

// MainContentComponent displays different views
type MainContentComponent struct {
	currentView int
	counter1    *CounterComponent
	counter2    *CounterComponent
}

func NewMainContentComponent() *MainContentComponent {
	return &MainContentComponent{
		counter1: NewCounterComponent("dash_count"),
		counter2: NewCounterComponent("settings_count"),
	}
}

func (m *MainContentComponent) Render() tizzy.Node {
	var content tizzy.Node

	switch m.currentView {
	case 0:
		content = tizzy.NewBox(
			tizzy.Style{FlexDirection: "column"},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Dashboard View"),
			tizzy.NewText(tizzy.Style{}, "Welcome to the main dashboard."),
			m.counter1.Render(),
		)
	case 1:
		content = tizzy.NewBox(
			tizzy.Style{FlexDirection: "column"},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "Settings View"),
			tizzy.NewText(tizzy.Style{}, "Configure your application here."),
			m.counter2.Render(),
		)
	case 2:
		content = tizzy.NewBox(
			tizzy.Style{FlexDirection: "column"},
			tizzy.NewText(tizzy.Style{Color: tcell.ColorGreen}, "About View"),
			tizzy.NewText(tizzy.Style{}, "Tizzy Component Demo"),
			tizzy.NewText(tizzy.Style{}, "This demonstrates state isolation."),
		)
	}

	return tizzy.NewBox(
		tizzy.Style{
			Border:        true,
			Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:        tizzy.Margin{Left: 1},
			FlexDirection: "column",
			FillWidth:     true,
			FillHeight:    true,
		},
		content,
	)
}

func main() {
	app, err := tizzy.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	mainContent := NewMainContentComponent()
	sidebar := &SidebarComponent{
		onSelect: func(idx int) {
			mainContent.currentView = idx
		},
	}

	render := func(ctx *tizzy.RenderContext) tizzy.Node {
		return tizzy.NewBox(
			tizzy.Style{
				Padding:       tizzy.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
				FlexDirection: "row",
				FillWidth:     true,
				FillHeight:    true,
			},
			sidebar.Render(),
			mainContent.Render(),
		)
	}

	if err := app.Run(render, nil); err != nil {
		log.Fatal(err)
	}
}
