package main

import (
	"fmt"
	tz "github.com/pavelgj/tizzy-go/tizzy"
	"log"

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

func (c *CounterComponent) Render() tz.Node {
	return tz.NewBox(
		tz.Style{
			Border:  true,
			Padding: tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:  tz.Margin{Top: 1},
		},
		tz.NewText(tz.Style{Color: tcell.ColorYellow}, fmt.Sprintf("Counter [%s]", c.id)),
		tz.NewText(tz.Style{}, fmt.Sprintf("Clicks: %d", c.count)),
		tz.NewButton(
			tz.Style{ID: c.id + "_btn", Focusable: true, Margin: tz.Margin{Top: 1}},
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

func (s *SidebarComponent) Render() tz.Node {
	items := []string{"Dashboard", "Settings", "About"}
	var children []tz.Node
	children = append(children, tz.NewText(tz.Style{Color: tcell.ColorLightCyan}, " NAVIGATION "))

	for i, item := range items {
		style := tz.Style{
			Focusable: true,
			ID:        fmt.Sprintf("side_%d", i),
			Margin:    tz.Margin{Top: 1},
		}
		label := "  " + item
		if i == s.selectedTab {
			label = "> " + item
			style.Color = tcell.ColorYellow
		}

		idx := i // capture for closure
		children = append(children, tz.NewButton(style, label, func() {
			s.selectedTab = idx
			if s.onSelect != nil {
				s.onSelect(idx)
			}
		}))
	}

	return tz.NewBox(
		tz.Style{
			Border:        true,
			Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 1, Right: 1},
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

func (m *MainContentComponent) Render() tz.Node {
	var content tz.Node

	switch m.currentView {
	case 0:
		content = tz.NewBox(
			tz.Style{FlexDirection: "column"},
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Dashboard View"),
			tz.NewText(tz.Style{}, "Welcome to the main dashboard."),
			m.counter1.Render(),
		)
	case 1:
		content = tz.NewBox(
			tz.Style{FlexDirection: "column"},
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "Settings View"),
			tz.NewText(tz.Style{}, "Configure your application here."),
			m.counter2.Render(),
		)
	case 2:
		content = tz.NewBox(
			tz.Style{FlexDirection: "column"},
			tz.NewText(tz.Style{Color: tcell.ColorGreen}, "About View"),
			tz.NewText(tz.Style{}, "Tizzy Component Demo"),
			tz.NewText(tz.Style{}, "This demonstrates state isolation."),
		)
	}

	return tz.NewBox(
		tz.Style{
			Border:        true,
			Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
			Margin:        tz.Margin{Left: 1},
			FlexDirection: "column",
			FillWidth:     true,
			FillHeight:    true,
		},
		content,
	)
}

func main() {
	app, err := tz.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	mainContent := NewMainContentComponent()
	sidebar := &SidebarComponent{
		onSelect: func(idx int) {
			mainContent.currentView = idx
		},
	}

	render := func(ctx *tz.RenderContext) tz.Node {
		return tz.NewBox(
			tz.Style{
				Padding:       tz.Padding{Top: 1, Bottom: 1, Left: 2, Right: 2},
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
