package splotch

// Button is a node that allows user interaction.
type Button struct {
	Style   Style
	Label   string
	OnClick func()
}

// NewButton creates a new Button node.
func NewButton(style Style, label string, onClick func()) *Button {
	return &Button{
		Style:   style,
		Label:   label,
		OnClick: onClick,
	}
}

// node implements the Node interface.
func (b *Button) node() {}
