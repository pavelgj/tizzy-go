package splotch

// TextInput is a node that allows text input.
type TextInput struct {
	Style    Style
	Value    string
	OnChange func(string)
}

// NewTextInput creates a new TextInput node.
func NewTextInput(style Style, value string, onChange func(string)) *TextInput {
	return &TextInput{
		Style:    style,
		Value:    value,
		OnChange: onChange,
	}
}

// node implements the Node interface.
func (t *TextInput) node() {}
