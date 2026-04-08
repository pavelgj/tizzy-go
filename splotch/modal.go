package splotch

// Modal represents a dialog overlay component.
type Modal struct {
	Style Style
	Child Node
}

func (m *Modal) node() {}

// NewModal creates a new Modal component.
func NewModal(style Style, child Node) *Modal {
	return &Modal{
		Style: style,
		Child: child,
	}
}

// ModalState stores the interactive state of a Modal.
type ModalState struct {
	Open bool
}
