package splotch

// Checkbox is a node that allows user to toggle a boolean value.
type Checkbox struct {
	Style         Style
	Label         string
	Checked       bool
	CheckedChar   string // Default "x"
	UncheckedChar string // Default " "
	OnChange      func(bool)
}

// NewCheckbox creates a new Checkbox node.
func NewCheckbox(style Style, label string, checked bool, onChange func(bool)) *Checkbox {
	return &Checkbox{
		Style:         style,
		Label:         label,
		Checked:       checked,
		CheckedChar:   "x",
		UncheckedChar: " ",
		OnChange:      onChange,
	}
}

// node implements the Node interface.
func (c *Checkbox) node() {}
