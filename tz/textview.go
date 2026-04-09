package tz

import "strings"

// NewTextView creates a component that displays multiline text by splitting it into lines.
func NewTextView(style Style, text string) Node {
	lines := strings.Split(text, "\n")
	var children []Node
	for _, line := range lines {
		children = append(children, NewText(Style{}, line))
	}
	return NewBox(style, children...)
}
