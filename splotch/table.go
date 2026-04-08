package splotch

// Table is a node that displays tabular data.
type Table struct {
	Style               Style
	Headers             []string
	Rows                [][]string
	ColWidths           []int // Optional, if empty calculate based on content
	CalculatedColWidths []int // Set during layout
}

// NewTable creates a new Table node.
func NewTable(style Style, headers []string, rows [][]string) *Table {
	return &Table{
		Style:   style,
		Headers: headers,
		Rows:    rows,
	}
}

// node implements the Node interface.
func (t *Table) node() {}
