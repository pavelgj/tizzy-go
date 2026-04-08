package splotch

// RadioButton is a node that allows user to select a single option from a set.
type RadioButton struct {
	Style          Style
	Label          string
	Value          string
	Selected       bool
	SelectedChar   string // Default "*"
	UnselectedChar string // Default " "
	OnChange       func(string)
}

// NewRadioButton creates a new RadioButton node.
func NewRadioButton(style Style, label string, value string, selected bool, onChange func(string)) *RadioButton {
	return &RadioButton{
		Style:          style,
		Label:          label,
		Value:          value,
		Selected:       selected,
		SelectedChar:   "*",
		UnselectedChar: " ",
		OnChange:       onChange,
	}
}

// node implements the Node interface.
func (r *RadioButton) node() {}
