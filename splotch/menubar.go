package splotch

import "github.com/gdamore/tcell/v2"

// MenuItem represents an item in a menu.
type MenuItem struct {
	Label    string
	Shortcut tcell.Key // For special keys like F1, F2
	AltRune  rune      // For Alt+Letter shortcuts (e.g., 'f' for Alt+F)
	Action   func()
	Disabled bool
}

// Menu represents a single menu in the menu bar.
type Menu struct {
	Title   string
	Items   []MenuItem
	AltRune rune // Shortcut to open this menu (e.g., 'f' for Alt+F)
}

// MenuBar is a component that displays a horizontal bar of menus.
type MenuBar struct {
	Style Style
	Menus []Menu
}

func (m *MenuBar) node() {}

// NewMenuBar creates a new MenuBar component.
func NewMenuBar(style Style, menus []Menu) *MenuBar {
	return &MenuBar{
		Style: style,
		Menus: menus,
	}
}

// MenuBarState stores the interactive state of a MenuBar.
type MenuBarState struct {
	OpenMenuIndex    int // -1 means no menu open
	FocusedItemIndex int
}
