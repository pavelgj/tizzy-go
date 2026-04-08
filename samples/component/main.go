package main

import (
	"fmt"
	"log"
	"splotch/splotch"

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

func (c *CounterComponent) Render() splotch.Node {
	return splotch.NewBox(
		splotch.Style{
			Border:  true,
			Padding: splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:  splotch.Margin{Top: 1},
		},
		splotch.NewText(splotch.Style{Color: tcell.ColorYellow}, fmt.Sprintf("Counter [%s]", c.id)),
		splotch.NewText(splotch.Style{}, fmt.Sprintf("Clicks: %d", c.count)),
		splotch.NewButton(
			splotch.Style{ID: c.id + "_btn", Focusable: true, Margin: splotch.Margin{Top: 1}},
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

func (s *SidebarComponent) Render() splotch.Node {
	items := []string{"Dashboard", "Settings", "About"}
	var children []splotch.Node
	children = append(children, splotch.NewText(splotch.Style{Color: tcell.ColorLightCyan}, " NAVIGATION "))

	for i, item := range items {
		style := splotch.Style{
			Focusable: true,
			ID:        fmt.Sprintf("side_%d", i),
			Margin:    splotch.Margin{Top: 1},
		}
		label := "  " + item
		if i == s.selectedTab {
			label = "> " + item
			style.Color = tcell.ColorYellow
		}

		idx := i // capture for closure
		children = append(children, splotch.NewButton(style, label, func() {
			s.selectedTab = idx
			if s.onSelect != nil {
				s.onSelect(idx)
			}
		}))
	}

	return splotch.NewBox(
		splotch.Style{
			Border:        true,
			Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
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

func (m *MainContentComponent) Render() splotch.Node {
	var content splotch.Node

	switch m.currentView {
	case 0:
		content = splotch.NewBox(
			splotch.Style{FlexDirection: "column"},
			splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "Dashboard View"),
			splotch.NewText(splotch.Style{}, "Welcome to the main dashboard."),
			m.counter1.Render(),
		)
	case 1:
		content = splotch.NewBox(
			splotch.Style{FlexDirection: "column"},
			splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "Settings View"),
			splotch.NewText(splotch.Style{}, "Configure your application here."),
			m.counter2.Render(),
		)
	case 2:
		content = splotch.NewBox(
			splotch.Style{FlexDirection: "column"},
			splotch.NewText(splotch.Style{Color: tcell.ColorGreen}, "About View"),
			splotch.NewText(splotch.Style{}, "Splotch Component Demo"),
			splotch.NewText(splotch.Style{}, "This demonstrates state isolation."),
		)
	}

	return splotch.NewBox(
		splotch.Style{
			Border:        true,
			Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:        splotch.Margin{Left: 1},
			FlexDirection: "column",
			FillWidth:     true,
			FillHeight:    true,
		},
		content,
	)
}

func main() {
	app, err := splotch.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	mainContent := NewMainContentComponent()
	sidebar := &SidebarComponent{
		onSelect: func(idx int) {
			mainContent.currentView = idx
		},
	}

	render := func(ctx *splotch.RenderContext) splotch.Node {
		return splotch.NewBox(
			splotch.Style{
				Padding:       splotch.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
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
