package splotch

// Tab represents a single tab in a Tabs component.
type Tab struct {
	Label   string
	Content Node
}

// Tabs represents a component that allows switching between different content views.
type Tabs struct {
	Style Style
	Tabs  []Tab
}

// TabsState tracks the active tab index.
type TabsState struct {
	ActiveTab int
}

// NewTabs creates a new Tabs component.
func NewTabs(style Style, tabs []Tab) *Tabs {
	return &Tabs{
		Style: style,
		Tabs:  tabs,
	}
}

// node implements the Node interface.
func (t *Tabs) node() {}
