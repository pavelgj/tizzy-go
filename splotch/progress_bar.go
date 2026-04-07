package splotch

// ProgressBar is a node that displays a visual representation of completion percentage.
type ProgressBar struct {
	Style      Style
	Percent    float64 // 0.0 to 1.0
	FilledChar string
	EmptyChar  string
}

// NewProgressBar creates a new ProgressBar node.
func NewProgressBar(style Style, percent float64) *ProgressBar {
	return &ProgressBar{
		Style:      style,
		Percent:    percent,
		FilledChar: "█",
		EmptyChar:  "░",
	}
}

// node implements the Node interface.
func (p *ProgressBar) node() {}
