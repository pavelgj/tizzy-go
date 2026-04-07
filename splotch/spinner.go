package splotch

import "time"

// Spinner is a node that displays a loading animation.
type Spinner struct {
	Style    Style
	Frames   []string
	Interval time.Duration
}

// NewSpinner creates a new Spinner node.
func NewSpinner(style Style) *Spinner {
	return &Spinner{
		Style:    style,
		Frames:   []string{"|", "/", "-", "\\"},
		Interval: 100 * time.Millisecond,
	}
}

// node implements the Node interface.
func (s *Spinner) node() {}
